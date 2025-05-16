// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"net/http"

	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/tunnel/internal/iprose"
	"github.com/vpnhouse/tunnel/internal/proxy"
	"github.com/vpnhouse/tunnel/internal/stats"
	"github.com/vpnhouse/tunnel/internal/wireguard"
)

// AdminGetStatus returns current server status
func (tun *TunnelAPI) AdminGetStatus(w http.ResponseWriter, r *http.Request) {
	transform := func(global *stats.Stats) *adminAPI.ProtocolStats {
		return &adminAPI.ProtocolStats{
			PeersActive: &global.PeersActive,
			PeersTotal:  &global.PeersTotal,
			TrafficUp:   &global.UpstreamBytes,
			TrafficDown: &global.DownstreamBytes,
			SpeedDown:   &global.DownstreamSpeed,
			SpeedUp:     &global.UpstreamSpeed,
		}
	}

	xhttp.JSONResponse(w, func() (interface{}, error) {
		global, proto := tun.stats.Stats()
		flags := tun.runtime.Flags
		status := adminAPI.ServiceStatusResponse{
			RestartRequired: flags.RestartRequired,
			StatsGlobal:     *transform(&global),
		}

		if proxy, ok := proto[proxy.ProtoName]; ok {
			status.StatsProxy = transform(&proxy)
		}

		if iprose, ok := proto[iprose.ProtoName]; ok {
			status.StatsIprose = transform(&iprose)
		}

		if wireguard, ok := proto[wireguard.ProtoName]; ok {
			status.StatsWireguard = transform(&wireguard)
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
