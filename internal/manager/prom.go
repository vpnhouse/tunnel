// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vishvananda/netlink"
)

var wgPeersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "peers",
	Help:      "number of allocated peers",
})

var wgActiveGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "active",
	Help:      "number of peers with active WG handshake",
})

var wgInterfaceRxBytes = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "rx_bytes",
	Help:      "bytes received by the WG interface",
})

var wgInterfaceTxBytes = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "tx_bytes",
	Help:      "bytes transmitted by the WG interface",
})

var wgInterfaceRxPackets = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "rx_packets",
	Help:      "packets received by the WG interface",
})

var wgInterfaceTxPackets = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "tunnel",
	Subsystem: "wireguard",
	Name:      "tx_packets",
	Help:      "packets transmitted by the WG interface",
})

func init() {
	prometheus.MustRegister(
		wgPeersGauge, wgActiveGauge,
		wgInterfaceRxPackets, wgInterfaceRxBytes,
		wgInterfaceTxPackets, wgInterfaceTxBytes,
	)
}

func updatePrometheusFromLinkStats(ls *netlink.LinkStatistics) {
	wgInterfaceRxPackets.Set(float64(ls.RxPackets))
	wgInterfaceRxBytes.Set(float64(ls.RxBytes))
	wgInterfaceTxPackets.Set(float64(ls.TxPackets))
	wgInterfaceTxBytes.Set(float64(ls.TxBytes))
}
