// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/internal/extstat"
	"github.com/vpnhouse/tunnel/internal/settings"
	"github.com/vpnhouse/tunnel/pkg/control"
	"github.com/vpnhouse/tunnel/pkg/validator"
	"github.com/vpnhouse/tunnel/pkg/version"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"go.uber.org/zap"
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
		// blocks for a certificate issuing, timeout is a LE request timeout is about 10s
		if needCert := setDomainConfig(tun.runtime.Settings, dc); needCert {
			if err := tun.issueCertificateSync(); err != nil {
				return nil, err
			}
		}

		// setting the password resets the "initial setup required" flag.
		if err := tun.runtime.Settings.SetAdminPassword(req.AdminPassword); err != nil {
			return nil, err
		}

		if err := tun.runtime.Settings.Flush(); err != nil {
			return nil, err
		}
		if req.SendStats != nil && *req.SendStats {
			cfg := extstat.Defaults()
			tun.runtime.ExternalStats = extstat.New(tun.runtime.Settings.InstanceID, cfg)
			tun.runtime.Settings.ExternalStats = cfg
		}

		tun.runtime.ExternalStats.OnInstall()
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

		if err := tun.mergeStaticSettings(tun.runtime.Settings, newSettings); err != nil {
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
			DomainName: s.Domain.Name,
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

func (tun *TunnelAPI) mergeStaticSettings(current *settings.Config, s adminAPI.Settings) error {
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
	if s.Domain != nil {
		tmpDC := &xhttp.DomainConfig{
			Name:     s.Domain.DomainName,
			Mode:     string(s.Domain.Mode),
			IssueSSL: s.Domain.IssueSsl,
			Schema:   string(s.Domain.Schema),
		}
		if err := tmpDC.Validate(); err != nil {
			return err
		}

		if needCert := setDomainConfig(tun.runtime.Settings, tmpDC); needCert {
			// blocks for a certificate issuing, timeout is a LE request timeout is about 10s
			if err := tun.issueCertificateSync(); err != nil {
				return err
			}
		}
	} else {
		// consider "domain: null" as "disabled for the whole option set"
		current.Domain = nil
	}

	if s.SendStats != nil && *s.SendStats {
		if current.ExternalStats == nil {
			current.ExternalStats = extstat.Defaults()
		} else {
			current.ExternalStats.Enabled = true
		}
	}

	return nil
}

func (tun *TunnelAPI) issueCertificateSync() error {
	issuer, err := xhttp.NewIssuer(xhttp.IssuerOpts{
		Domain:   tun.runtime.Settings.Domain.Name,
		CacheDir: tun.runtime.Settings.ConfigDir(),
		Router:   tun.runtime.HttpRouter,
		Callback: func(_ *tls.Config) {
			zap.L().Info("ssl certificate issued", zap.String("name", tun.runtime.Settings.Domain.Name))
		},
	})
	if err != nil {
		return err
	}

	// ask for the config (it will be cached inside and re-used after the restart).
	_, err = issuer.TLSConfig()
	return err
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
			oldName = c.Domain.Name
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
		return dc.Name != oldName
	}

	return false
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
