// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/settings"
	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/validator"
	"github.com/Codename-Uranium/tunnel/pkg/version"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
)

// AdminGetSettings implements handler for GET /api/tunnel/admin/settings request
func (tun *TunnelAPI) AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		s := settingsToOpenAPI(tun.runtime.Settings, tun.runtime.DynamicSettings)
		return s, nil
	})
}

type passwordCleaner interface {
	CleanAdminPassword()
}

func (tun *TunnelAPI) TmpResetSettingsToDefault(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if c, ok := tun.runtime.DynamicSettings.(passwordCleaner); ok {
			c.CleanAdminPassword()
		}

		tun.runtime.Settings.Wireguard.Subnet = "10.235.0.0/24"
		if err := tun.runtime.Settings.Write(); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

// AdminInitialSetup POST /api/tunnel/admin/initial-setup
func (tun *TunnelAPI) AdminInitialSetup(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if !version.IsPersonal() {
			return nil, xerror.EForbidden("initial setup is disabled for `" + version.GetVersion() + "`")
		}
		if !tun.runtime.DynamicSettings.InitialSetupRequired() {
			return nil, xerror.EForbidden("the initial configuration has already been applied")
		}

		var req adminAPI.InitialSetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, xerror.EInvalidArgument("failed to unmarshal request", err)
		}

		if err := validateInitialSetupRequest(req); err != nil {
			return nil, err
		}

		tun.runtime.Settings.Wireguard.Subnet = validator.Subnet(req.ServerIpMask)
		if req.EnableSsl {
			tun.runtime.Settings.SSL = &xhttp.SSLConfig{
				ListenAddr: ":443",
				Domain:     req.DomainName,
			}
		}
		if err := tun.runtime.Settings.Write(); err != nil {
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

func validateInitialSetupRequest(req adminAPI.InitialSetupRequest) error {
	if len(req.DomainName) > 0 && !req.EnableSsl {
		return xerror.EInvalidField("domain name without SSL enabled is meaningless", "domain_name", nil)
	}
	if len(req.AdminPassword) < 6 {
		return xerror.EInvalidField("password too short", "admin_password", nil)
	}
	return nil
}

// AdminUpdateSettings implements handler for PATCH /api/tunnel/admin/settings request
func (tun *TunnelAPI) AdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		newSettings, err := openApiSettingsFromRequest(r)
		if err != nil {
			return nil, err
		}

		if err := updateDynamicSettings(tun.runtime.DynamicSettings, newSettings); err != nil {
			return nil, err
		}

		tun.runtime.Settings = mergeStaticSettings(tun.runtime.Settings, newSettings)
		if err := tun.runtime.Settings.Write(); err != nil {
			return nil, err
		}

		tun.runtime.Events.EmitEvent(control.EventRestart)
		updated := settingsToOpenAPI(tun.runtime.Settings, tun.runtime.DynamicSettings)
		return updated, nil
	})
}

func settingsToOpenAPI(s settings.StaticConfig, d settings.DynamicConfig) adminAPI.Settings {
	public := d.GetWireguardPrivateKey().Public().Unwrap().String()
	subnet := string(s.Wireguard.Subnet)
	wgPublicPort := s.Wireguard.ClientPort()
	return adminAPI.Settings{
		AdminUserName:      &s.GetAdminAPConfig().UserName,
		ConnectionTimeout:  &s.GetPublicAPIConfig().PeerTTL,
		Dns:                &s.Wireguard.DNS,
		LogLevel:           (*adminAPI.SettingsLogLevel)(&s.LogLevel),
		PingInterval:       &s.GetPublicAPIConfig().PingInterval,
		WireguardKeepalive: &s.Wireguard.Keepalive,
		//  note: return both ports, allow to update only the `WireguardServerPort` value.
		WireguardListenPort: &s.Wireguard.ListenPort,
		WireguardServerPort: &wgPublicPort,
		WireguardPublicKey:  &public,
		WireguardServerIpv4: &s.Wireguard.ServerIPv4,
		WireguardSubnet:     &subnet,
	}
}

func mergeStaticSettings(current settings.StaticConfig, s adminAPI.Settings) settings.StaticConfig {
	if s.LogLevel != nil {
		current.LogLevel = (string)(*s.LogLevel)
	}

	if s.Dns != nil {
		current.Wireguard.DNS = *s.Dns
	}
	if s.WireguardKeepalive != nil {
		current.Wireguard.Keepalive = *s.WireguardKeepalive
	}
	if s.WireguardServerIpv4 != nil {
		current.Wireguard.ServerIPv4 = *s.WireguardServerIpv4
	}
	if s.WireguardSubnet != nil {
		current.Wireguard.Subnet = validator.Subnet(*s.WireguardSubnet)
	}
	if s.WireguardServerPort != nil {
		current.Wireguard.NATedPort = *s.WireguardServerPort
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
