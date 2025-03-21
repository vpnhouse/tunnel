// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net"
	"net/http"

	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/common-lib-go/control"
	"github.com/vpnhouse/common-lib-go/validator"
	"github.com/vpnhouse/common-lib-go/version"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/common-lib-go/xnet"
	"github.com/vpnhouse/tunnel/internal/extstat"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/settings"
)

// AdminGetSettings implements handler for GET /api/tunnel/admin/settings request
func (tun *TunnelAPI) AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		s := settingsToOpenAPI(tun.runtime.Settings)
		return s, nil
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

		var dc *xhttp.DomainConfig = nil
		if req.Domain != nil {
			dc = &xhttp.DomainConfig{
				Mode:        string(req.Domain.Mode),
				PrimaryName: req.Domain.DomainName,
				IssueSSL:    req.Domain.IssueSsl,
				Schema:      string(req.Domain.Schema),
				Dir:         tun.runtime.Settings.ConfigDir(),
			}
			if err := dc.Validate(); err != nil {
				return nil, err
			}
		}

		subnet, err := validateSubnet(req.ServerIpMask)
		if err != nil {
			return nil, err
		}
		tun.runtime.Settings.Wireguard.Subnet = validator.Subnet(subnet)
		setDomainConfig(tun.runtime.Settings, dc)

		// setting the password resets the "initial setup required" flag.
		if err := tun.runtime.Settings.SetAdminPassword(req.AdminPassword); err != nil {
			return nil, err
		}

		if err := tun.runtime.Settings.Flush(); err != nil {
			return nil, err
		}
		if req.SendStats != nil && *req.SendStats {
			cfg := extstat.Defaults()
			tun.runtime.ReplaceExternalStatsService(
				extstat.New(tun.runtime.Settings.InstanceID, cfg),
			)
			tun.runtime.Settings.ExternalStats = cfg
		}

		tun.runtime.ExternalStats.OnInstall()
		tun.runtime.Events.EmitEvent(control.EventRestart)
		return nil, nil
	})
}

// AdminUpdateSettings implements handler for PATCH /api/tunnel/admin/settings request
func (tun *TunnelAPI) AdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		newSettings, err := openApiSettingsFromRequest(r)
		if err != nil {
			return nil, err
		}

		if err := tun.mergeStaticSettings(tun.runtime, newSettings); err != nil {
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
	var dc *adminAPI.DomainConfig = nil
	if s.Domain != nil {
		dc = &adminAPI.DomainConfig{
			DomainName: s.Domain.PrimaryName,
			IssueSsl:   s.Domain.IssueSSL,
			Mode:       adminAPI.DomainConfigMode(s.Domain.Mode),
			Schema:     adminAPI.DomainConfigSchema(s.Domain.Schema),
		}
	}
	sendStats := s.ExternalStats != nil && s.ExternalStats.Enabled
	return adminAPI.Settings{
		ConnectionTimeout:  &s.GetPublicAPIConfig().PeerTTL,
		Dns:                &s.Wireguard.DNS,
		PingInterval:       &s.GetPublicAPIConfig().PingInterval,
		WireguardKeepalive: &s.Wireguard.Keepalive,
		//  note: return both ports, allow to update only the `WireguardServerPort` value.
		WireguardListenPort: &s.Wireguard.ListenPort,
		WireguardServerPort: &wgPublicPort,
		WireguardPublicKey:  &public,
		WireguardServerIpv4: &s.Wireguard.ServerIPv4,
		WireguardSubnet:     &subnet,
		Domain:              dc,
		SendStats:           &sendStats,
	}
}

func (tun *TunnelAPI) mergeStaticSettings(rt *runtime.TunnelRuntime, s adminAPI.Settings) error {
	if s.AdminPassword != nil {
		if err := rt.Settings.SetAdminPassword(*s.AdminPassword); err != nil {
			return err
		}
	}
	if s.WireguardServerIpv4 != nil {
		newip := xnet.ParseIP(*s.WireguardServerIpv4)
		if newip.IP == nil {
			return xerror.WInvalidField("settings", "failed to parse IPv4 address", "wireguard_server_ipv4", nil)
		}
		if err := rt.Settings.SetPublicIP(newip); err != nil {
			return err
		}
	}

	if s.Dns != nil {
		rt.Settings.Wireguard.DNS = *s.Dns
	}
	if s.WireguardKeepalive != nil {
		rt.Settings.Wireguard.Keepalive = *s.WireguardKeepalive
	}
	if s.WireguardSubnet != nil {
		subnet, err := validateSubnet(*s.WireguardSubnet)
		if err != nil {
			return err
		}
		rt.Settings.Wireguard.Subnet = validator.Subnet(subnet)
	}
	if s.WireguardServerPort != nil {
		rt.Settings.Wireguard.NATedPort = *s.WireguardServerPort
	}
	if s.Domain != nil {
		tmpDC := &xhttp.DomainConfig{
			PrimaryName: s.Domain.DomainName,
			Mode:        string(s.Domain.Mode),
			IssueSSL:    s.Domain.IssueSsl,
			Schema:      string(s.Domain.Schema),
		}
		if err := tmpDC.Validate(); err != nil {
			return err
		}

		setDomainConfig(tun.runtime.Settings, tmpDC)
	} else {
		// consider "domain: null" as "disabled for the whole option set"
		rt.Settings.Domain = nil
	}

	if s.SendStats != nil {
		if *s.SendStats {
			if rt.Settings.ExternalStats == nil {
				rt.Settings.ExternalStats = extstat.Defaults()
			} else {
				rt.Settings.ExternalStats.Enabled = true
			}
		} else {
			if rt.Settings.ExternalStats != nil {
				rt.Settings.ExternalStats.Enabled = false
			}
		}
		rt.ReplaceExternalStatsService(
			extstat.New(rt.Settings.InstanceID, rt.Settings.ExternalStats),
		)
	}

	return nil
}

// setDomainConfig updates current settings with new domain config,
// return true if the new certificate must be issued.
func setDomainConfig(c *settings.Config, dc *xhttp.DomainConfig) bool {
	if dc == nil {
		return false
	}

	oldName := ""
	if c.Domain != nil {
		if c.Domain.Mode == string(adminAPI.DomainConfigModeDirect) {
			oldName = c.Domain.PrimaryName
		}
	}

	if len(dc.Dir) == 0 {
		dc.Dir = c.ConfigDir()
	}
	c.Domain = dc
	if dc.IssueSSL {
		if c.SSL == nil || len(c.SSL.ListenAddr) == 0 {
			c.SSL = &xhttp.SSLConfig{
				ListenAddr: ":443",
			}
		}
		// notify caller that the name differs
		return dc.PrimaryName != oldName
	}

	return false
}

func validateSubnet(s string) (string, error) {
	_, netw, err := net.ParseCIDR(s)
	if err != nil {
		return "", err
	}
	if v4 := netw.IP.To4(); v4 == nil {
		return "", xerror.EInvalidArgument("non-IPv4 subnet given", nil)
	}
	if ones, _ := netw.Mask.Size(); ones < 8 || ones > 30 {
		return "", xerror.EInvalidArgument("invalid subnet size given, want /8 to /30", nil)
	}

	if !xnet.IsPrivateIPNet(netw) {
		return "", xerror.EInvalidArgument("non-private subnet given", nil)
	}

	return netw.String(), nil
}

// openApiSettingsFromRequest parses settings information from request body.
// WARNING! This function does not do any verification of imported data! Caller must do it itself!
func openApiSettingsFromRequest(r *http.Request) (adminAPI.Settings, error) {
	var oSettings adminAPI.Settings
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&oSettings); err != nil {
		return adminAPI.Settings{}, xerror.EInvalidArgument("invalid settings", err)
	}

	return oSettings, nil
}
