// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"sync"

	"github.com/vishvananda/netlink"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/storage"
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
	// Downstream traffic totally
	Downstream int64
}

type Manager struct {
	runtime       *runtime.TunnelRuntime
	mutex         sync.RWMutex
	storage       *storage.Storage
	wireguard     *wireguard.Wireguard
	ip4am         *ipam.IPAM
	eventLog      eventlog.EventManager
	running       bool
	bgStopChannel chan bool
	bgWaitGroup   sync.WaitGroup

	// statistic guarded by mutex and
	// updated by the backgroundOnce routine.
	statistic CachedStatistics
}

func New(runtime *runtime.TunnelRuntime, storage *storage.Storage, wireguard *wireguard.Wireguard, ip4am *ipam.IPAM, eventLog eventlog.EventManager) (*Manager, error) {
	manager := &Manager{
		runtime:       runtime,
		storage:       storage,
		wireguard:     wireguard,
		ip4am:         ip4am,
		eventLog:      eventLog,
		running:       true,
		bgStopChannel: make(chan bool),
		statistic: CachedStatistics{
			Upstream:   storage.GetUpstreamMetric(),
			Downstream: storage.GetDownstreamMetric(),
		},
	}

	// Run background goroutine
	manager.bgWaitGroup.Add(1)
	go manager.background()

	manager.restorePeers()

	return manager, nil
}

func (manager *Manager) Shutdown() error {
	// Shutdown background goroutine
	zap.L().Debug("Sending stop signal to manager background goroutine")
	manager.bgStopChannel <- true

	zap.L().Debug("Waiting for shutting down background goroutine")
	manager.bgWaitGroup.Wait()

	// Get lock and forbid all further operations
	zap.L().Debug("Acquiring main manager lock")
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	zap.L().Debug("Marking manager as not accepting any requests anymore")
	manager.running = false

	return nil
}

func (manager *Manager) Running() bool {
	return manager.running
}

func (manager *Manager) GetCachedStatistics() CachedStatistics {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return manager.statistic
}
