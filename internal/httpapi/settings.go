package httpapi

import (
	"encoding/json"
	"net/http"
	"os"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/settings"
	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/validator"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// AdminGetSettings implements handler for GET /settings request
func (tun *TunnelAPI) AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		s := settingsToOpenAPI(tun.runtime.Settings, tun.runtime.DynamicSettings)
		return s, nil
	})
}

type initialSetupRequest struct {
	AdminPassword string `json:"admin_password"`
	// ip/netmask actually
	ServerIP string `json:"server_ip_mask"`
}

func validate(req initialSetupRequest) error {
	if len(req.AdminPassword) < 6 {
		return xerror.EInvalidField("password too short", "admin_password", nil)
	}

	ipAddr, ipNet, err := xnet.ParseCIDR(req.ServerIP)
	if err != nil {
		return xerror.EInvalidField("failed to parse subnet", "server_ip_mask", err)
	}

	if !ipAddr.Isv4() {
		return xerror.EInvalidField("only ipv4 network supported", "server_ip_mask", nil)
	}

	if !ipAddr.IP.IsPrivate() {
		return xerror.EInvalidField("not a private subnet given", "server_ip_mask", nil)
	}

	if ipAddr.Equal(ipNet.IP()) {
		return xerror.EInvalidField("ip/mask required, but subnet/mask given", "server_ip_mask", nil)
	}

	size, _ := ipNet.Mask().Size()
	if size < 8 || size > 31 {
		return xerror.EInvalidField("invalid subnet size", "server_ip_mask", nil)
	}

	return nil
}

func (tun *TunnelAPI) AdminInitialSetup(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if !tun.runtime.DynamicSettings.InitialSetupRequired() {
			return nil, xerror.EForbidden("the initial configuration has already been applied")
		}

		var req initialSetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, xerror.EInvalidArgument("failed to unmarshal request", err)
		}

		if err := validate(req); err != nil {
			return nil, err
		}

		tun.runtime.Settings.Wireguard.Subnet = req.ServerIP
		if err := validateAndWriteSettings(tun.runtime.Settings); err != nil {
			return nil, err
		}

		// setting the password resets the "initial setup required" flag.
		if err := tun.runtime.DynamicSettings.SetAdminPassword(req.AdminPassword); err != nil {
			return nil, err
		}

		tun.runtime.Events.EmitEvent(control.EventRestart)
		return nil, nil
	})
}

// AdminUpdateSettings implements handler for PATCH /settings request
func (tun *TunnelAPI) AdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		newSettings, err := openApiSettingsFromRequest(r)
		if err != nil {
			return nil, err
		}

		if err := updateDynamicSettings(tun.runtime.DynamicSettings, newSettings); err != nil {
			return nil, err
		}

		static := mergeStaticSettings(tun.runtime.Settings, newSettings)
		if err := validateAndWriteSettings(static); err != nil {
			return nil, err
		}

		tun.runtime.Events.EmitEvent(control.EventNeedRestart)
		return nil, nil
	})
}

func validateAndWriteSettings(newSettings settings.StaticConfig) error {
	if err := validator.ValidateStruct(newSettings); err != nil {
		return xerror.EInvalidArgument("failed to validate static config", err)
	}

	bs, _ := yaml.Marshal(newSettings)
	path := newSettings.GetPath()
	if err := os.WriteFile(path, bs, 0600); err != nil {
		return xerror.WInternalError("config", "failed to write static config",
			err, zap.String("path", path))
	}
	return nil
}

func settingsToOpenAPI(s settings.StaticConfig, d settings.DynamicConfig) adminAPI.Settings {
	key := d.GetWireguardPublicKey().PublicKey().String()
	return adminAPI.Settings{
		AdminUserName:       &s.AdminAPI.UserName,
		ConnectionTimeout:   &s.PublicAPI.PeerTTL,
		Dns:                 &s.Wireguard.DNS,
		HttpListenAddr:      &s.HTTPListenAddr,
		LogLevel:            (*adminAPI.SettingsLogLevel)(&s.LogLevel),
		PingInterval:        &s.PublicAPI.PingInterval,
		WireguardKeepalive:  &s.Wireguard.Keepalive,
		WireguardListenPort: &s.Wireguard.ServerPort,
		WireguardPublicKey:  &key,
		WireguardServerIpv4: &s.Wireguard.ServerIPv4,
		WireguardSubnet:     &s.Wireguard.Subnet,
	}
}

func mergeStaticSettings(current settings.StaticConfig, s adminAPI.Settings) settings.StaticConfig {
	if s.LogLevel != nil {
		current.LogLevel = (string)(*s.LogLevel)
	}
	if s.HttpListenAddr != nil {
		current.HTTPListenAddr = *s.HttpListenAddr
	}

	if s.ConnectionTimeout != nil {
		current.PublicAPI.PeerTTL = *s.ConnectionTimeout
	}
	if s.PingInterval != nil {
		current.PublicAPI.PingInterval = *s.PingInterval
	}

	if s.Dns != nil {
		current.Wireguard.DNS = *s.Dns
	}
	if s.WireguardKeepalive != nil {
		current.Wireguard.Keepalive = *s.WireguardKeepalive
	}
	if s.WireguardServerPort != nil {
		current.Wireguard.ServerPort = *s.WireguardServerPort
	}
	if s.WireguardServerIpv4 != nil {
		current.Wireguard.ServerIPv4 = *s.WireguardServerIpv4
	}
	if s.WireguardSubnet != nil {
		current.Wireguard.Subnet = *s.WireguardSubnet
	}

	return current
}

func updateDynamicSettings(d settings.DynamicConfig, s adminAPI.Settings) error {
	if s.AdminPassword != nil && len(*s.AdminPassword) > 0 {
		if err := d.SetAdminPassword(*s.AdminPassword); err != nil {
			return err
		}
	}
	return nil
}

// openApiSettingsFromRequest parses settings information from request body.
// WARNING! This function does not do any verification of imported data! Caller must do it itself!
func openApiSettingsFromRequest(r *http.Request) (adminAPI.Settings, error) {
	var oSettings adminAPI.Settings
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&oSettings); err != nil {
		return adminAPI.Settings{}, xerror.EInvalidArgument("invalid settings", err)
	}

	return oSettings, nil
}
