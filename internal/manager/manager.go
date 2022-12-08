// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/internal/wireguard"
	"github.com/vpnhouse/tunnel/pkg/ipam"
	"go.uber.org/zap"
)

type CachedStatistics struct {
	// PeersTotal is a number of peers
	// being authorized to connect to this node
	PeersTotal int
	// PeersWithTraffic is a number of peers
	// being actually connected to this node
	PeersWithTraffic int
	// PeersActiveLastHour is number of peers
	// having any exchange during last hour
	PeersActiveLastHour int
	// PeersActiveLastDay is number of peers
	// having any exchange during last 24 hours
	PeersActiveLastDay int
	// Wireguard link statistic, may be nil
	LinkStat *netlink.LinkStatistics
	// Upstream traffic totally
	Upstream int64
	// Upstream traffic speed
	UpstreamSpeed int64
	// Downstream traffic totally
	Downstream int64
	// Downstream traffic speed
	DownstreamSpeed int64

	// The time in seconds then statistics was collected
	Collected int64
}

func (s *CachedStatistics) UpdateSpeeds(prevStats *CachedStatistics) {
	if s.Collected == 0 || prevStats.Collected >= s.Collected || prevStats == nil {
		return
	}

	seconds := s.Collected - prevStats.Collected

	if s.Upstream >= prevStats.Upstream {
		s.UpstreamSpeed = (s.Upstream - prevStats.Upstream) / seconds
	}

	if s.Downstream >= prevStats.Downstream {
		s.DownstreamSpeed = (s.Downstream - prevStats.Downstream) / seconds
	}
}

type Manager struct {
	runtime           *runtime.TunnelRuntime
	lock              sync.RWMutex
	storage           *storage.Storage
	wireguard         *wireguard.Wireguard
	ip4am             *ipam.IPAM
	eventLog          eventlog.EventManager
	statsService      peerStatsService
	peerTrafficSender *peerTrafficUpdateEventSender
	running           atomic.Value
	stop              chan struct{}
	done              chan struct{}

	statistic atomic.Value // *CachedStatistics
}

func New(runtime *runtime.TunnelRuntime, storage *storage.Storage, wireguard *wireguard.Wireguard, ip4am *ipam.IPAM, eventLog eventlog.EventManager) (*Manager, error) {
	peerTrafficSender := NewPeerTrafficUpdateEventSender(runtime, eventLog, nil)

	manager := &Manager{
		runtime:           runtime,
		storage:           storage,
		wireguard:         wireguard,
		ip4am:             ip4am,
		eventLog:          eventLog,
		peerTrafficSender: peerTrafficSender,
		stop:              make(chan struct{}),
		done:              make(chan struct{}),
	}

	manager.restorePeers()
	manager.running.Store(true)
	manager.statistic.Store(&CachedStatistics{
		Upstream:   storage.GetUpstreamMetric(),
		Downstream: storage.GetDownstreamMetric(),
		Collected:  time.Now().Unix(),
	})

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

	// Stop sending all events
	manager.peerTrafficSender.Stop()

	return nil
}

func (manager *Manager) Running() bool {
	return manager.running.Load().(bool)
}

func (manager *Manager) GetCachedStatistics() *CachedStatistics {
	return manager.statistic.Load().(*CachedStatistics)
}

func (manager *Manager) GetPeerSpeeds(peer *types.PeerInfo) (int64, int64) {
	return manager.statsService.GetPeerSpeeds(manager.runtime.Settings.PeerStatistics.UpdateStatisticsInterval, peer)
}
