package httpapi

import (
	"encoding/json"
	"net/http"
	"os"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/settings"
	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/asaskevich/govalidator"
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
		ok, err := govalidator.ValidateStruct(static)
		if err != nil {
			return nil, xerror.EInvalidArgument("failed to validate static config", err)
		}
		if !ok {
			return nil, xerror.EInvalidArgument("failed to validate static config", nil)
		}

		bs, _ := yaml.Marshal(static)
		if err := os.WriteFile(static.GetPath(), bs, 0600); err != nil {
			return nil, xerror.WInternalError("config", "failed to write static config",
				err, zap.String("path", static.GetPath()))
		}

		tun.runtime.Events.EmitEvent(control.EventNeedRestart)
		return nil, nil
	})
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
