// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/common-lib-go/ipam"
	"github.com/vpnhouse/common-lib-go/statutils"
	"github.com/vpnhouse/common-lib-go/xstats"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/stats"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/internal/wireguard"
	"go.uber.org/zap"
)

const ProtoName = "wireguard"

type Manager struct {
	runtime       *runtime.TunnelRuntime
	lock          sync.RWMutex
	storage       *storage.Storage
	wireguard     *wireguard.Wireguard
	ip4am         *ipam.IPAM
	statsReporter *xstats.Service
	geoipService  *geoip.Instance
	wgStats       atomic.Pointer[wgStats]
	running       atomic.Value
	stop          chan struct{}
	done          chan struct{}

	upstreamSpeedAvg   *statutils.AvgValue
	downstreamSpeedAvg *statutils.AvgValue
}

func New(
	runtime *runtime.TunnelRuntime,
	storage *storage.Storage,
	wireguard *wireguard.Wireguard,
	ip4am *ipam.IPAM,
	statsService *stats.Service,
	geoipService *geoip.Instance,
) (*Manager, error) {
	manager := &Manager{
		runtime:            runtime,
		storage:            storage,
		wireguard:          wireguard,
		ip4am:              ip4am,
		geoipService:       geoipService,
		stop:               make(chan struct{}),
		done:               make(chan struct{}),
		upstreamSpeedAvg:   statutils.NewAvgValue(10),
		downstreamSpeedAvg: statutils.NewAvgValue(10),
	}
	var err error
	manager.statsReporter, err = statsService.Register(ProtoName, func() stats.ExtraStats {
		peers, err := manager.peers()
		if err != nil {
			return stats.ExtraStats{}
		}

		now := time.Now()
		var active int
		for _, peer := range peers {
			if peer.Activity != nil && now.Sub(*peer.Activity.TimePtr()) < time.Minute {
				active += 1
			}
		}

		return stats.ExtraStats{
			PeersTotal:  len(peers),
			PeersActive: active,
		}
	})
	if err != nil {
		return nil, err
	}
	manager.restorePeers()
	manager.running.Store(true)

	// Run background goroutine
	go manager.background()

	return manager, nil
}

func (manager *Manager) Shutdown() error {
	zap.L().Debug("Marking manager as not accepting any requests anymore")
	manager.running.Store(false)

	// Shutdown background goroutine
	zap.L().Debug("Sending stop signal to manager background goroutine")
	close(manager.stop)

	zap.L().Debug("Waiting for shutting down background goroutine")
	<-manager.done

	return nil
}

func (manager *Manager) Running() bool {
	return manager.running.Load().(bool)
}

func (manager *Manager) GetRuntimePeerStat(peer *types.PeerInfo) PeerStats {
	statsPtr := manager.wgStats.Load()
	return (*statsPtr)[*peer.WireguardPublicKey]
}

// admin.Handler implementation
func (manager *Manager) KillActiveUserSessions(userId string) {
	// TODO: add implementation
}
