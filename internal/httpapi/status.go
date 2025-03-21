// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"net/http"

	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/tunnel/internal/wireguard"
)

// AdminGetStatus returns current server status
func (tun *TunnelAPI) AdminGetStatus(w http.ResponseWriter, r *http.Request) {
	stats := tun.manager.GetCachedStatistics()
	xhttp.JSONResponse(w, func() (interface{}, error) {
		flags := tun.runtime.Flags
		status := adminAPI.ServiceStatusResponse{
			RestartRequired:  flags.RestartRequired,
			PeersTotal:       &stats.PeersTotal,
			PeersConnected:   &stats.PeersWithTraffic,
			PeersActive1h:    &stats.PeersActiveLastHour,
			PeersActive1d:    &stats.PeersActiveLastDay,
			TrafficUp:        &stats.Upstream,
			TrafficDown:      &stats.Downstream,
			TrafficUpSpeed:   &stats.UpstreamSpeed,
			TrafficDownSpeed: &stats.DownstreamSpeed,
		}
		return status, nil
	})
}

func (tun *TunnelAPI) AdminConnectionInfoWireguard(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if len(tun.runtime.Settings.Wireguard.ServerIPv4) == 0 {
			return nil, xerror.EInvalidConfiguration(
				"missing server public ipv4 option, please specify it in settings",
				"wireguard_server_ipv4")
		}
		info := wireguardConnectionInfo(tun.runtime.Settings.Wireguard)
		return info, nil
	})
}

func wireguardConnectionInfo(c wireguard.Config) adminAPI.WireguardOptions {
	return adminAPI.WireguardOptions{
		AllowedIps:      []string{"0.0.0.0/0"},
		Subnet:          string(c.Subnet),
		Dns:             c.DNS,
		Keepalive:       c.Keepalive,
		ServerIpv4:      c.ServerIPv4,
		ServerPort:      c.ClientPort(),
		ServerPublicKey: c.GetPrivateKey().Public().Unwrap().String(),
	}
}
