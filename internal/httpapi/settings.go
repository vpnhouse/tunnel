// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net/http"

	adminAPI "github.com/comradevpn/api/go/server/tunnel_admin"
	"github.com/comradevpn/tunnel/internal/settings"
	"github.com/comradevpn/tunnel/pkg/control"
	"github.com/comradevpn/tunnel/pkg/validator"
	"github.com/comradevpn/tunnel/pkg/version"
	"github.com/comradevpn/tunnel/pkg/xerror"
	"github.com/comradevpn/tunnel/pkg/xhttp"
	"github.com/comradevpn/tunnel/pkg/xnet"
)

// AdminGetSettings implements handler for GET /api/tunnel/admin/settings request
func (tun *TunnelAPI) AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		s := settingsToOpenAPI(tun.runtime.Settings)
		return s, nil
	})
}

func (tun *TunnelAPI) TmpResetSettingsToDefault(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		tun.runtime.Settings.Wireguard.Subnet = "10.235.0.0/24"
		tun.runtime.Settings.CleanAdminPassword()
		return nil, nil
	})
}

// AdminInitialSetup POST /api/tunnel/admin/initial-setup
func (tun *TunnelAPI) AdminInitialSetup(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if !version.IsPersonal() {
			return nil, xerror.EForbidden("initial setup is disabled for `" + version.GetVersion() + "`")
		}
		if !tun.runtime.Settings.InitialSetupRequired() {
			return nil, xerror.EForbidden("the initial configuration has already been applied")
		}

		var req adminAPI.InitialSetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, xerror.EInvalidArgument("failed to unmarshal request", err)
		}

		if err := validateInitialSetupRequest(req); err != nil {
			return nil, err
		}

		var dc *xhttp.DomainConfig = nil
		if req.Domain != nil {
			dc = &xhttp.DomainConfig{
				Mode:     string(req.Domain.Mode),
				Name:     req.Domain.DomainName,
				IssueSSL: req.Domain.IssueSsl,
				Schema:   string(req.Domain.Schema),
				Dir:      tun.runtime.Settings.ConfigDir(),
			}
			if err := dc.Validate(); err != nil {
				return nil, err
			}
		}

		tun.runtime.Settings.Wireguard.Subnet = validator.Subnet(req.ServerIpMask)
		tun.runtime.Settings.Domain = dc
		if dc != nil && dc.IssueSSL {
			tun.runtime.Settings.SSL = &xhttp.SSLConfig{
				ListenAddr: ":443",
			}
		}

		// setting the password resets the "initial setup required" flag.
		if err := tun.runtime.Settings.SetAdminPassword(req.AdminPassword); err != nil {
			return nil, err
		}

		if err := tun.runtime.Settings.Flush(); err != nil {
			return nil, err
		}

		tun.runtime.Events.EmitEvent(control.EventRestart)
		return nil, nil
	})
}

func validateInitialSetupRequest(req adminAPI.InitialSetupRequest) error {
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

		if err := mergeStaticSettings(tun.runtime.Settings, newSettings); err != nil {
			return nil, err
		}

		if err := tun.runtime.Settings.Flush(); err != nil {
			return nil, err
		}

		tun.runtime.Events.EmitEvent(control.EventRestart)
		updated := settingsToOpenAPI(tun.runtime.Settings)
		return updated, nil
	})
}

func settingsToOpenAPI(s *settings.Config) adminAPI.Settings {
	public := s.Wireguard.GetPrivateKey().Public().Unwrap().String()
	subnet := string(s.Wireguard.Subnet)
	wgPublicPort := s.Wireguard.ClientPort()
	return adminAPI.Settings{
		AdminUserName:      &s.AdminAPI.UserName,
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

func mergeStaticSettings(current *settings.Config, s adminAPI.Settings) error {
	if s.LogLevel != nil {
		current.LogLevel = (string)(*s.LogLevel)
	}

	if s.AdminPassword != nil {
		if err := current.SetAdminPassword(*s.AdminPassword); err != nil {
			return err
		}
	}
	if s.WireguardServerIpv4 != nil {
		newip := xnet.ParseIP(*s.WireguardServerIpv4)
		if newip.IP == nil {
			return xerror.WInvalidField("settings", "failed to parse IPv4 address", "wireguard_server_ipv4", nil)
		}
		if err := current.SetPublicIP(newip); err != nil {
			return err
		}
	}

	if s.Dns != nil {
		current.Wireguard.DNS = *s.Dns
	}
	if s.WireguardKeepalive != nil {
		current.Wireguard.Keepalive = *s.WireguardKeepalive
	}
	if s.WireguardSubnet != nil {
		current.Wireguard.Subnet = validator.Subnet(*s.WireguardSubnet)
	}
	if s.WireguardServerPort != nil {
		current.Wireguard.NATedPort = *s.WireguardServerPort
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
