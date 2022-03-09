// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"sync"

	"github.com/comradevpn/tunnel/internal/eventlog"
	"github.com/comradevpn/tunnel/internal/runtime"
	"github.com/comradevpn/tunnel/internal/storage"
	"github.com/comradevpn/tunnel/internal/wireguard"
	"github.com/comradevpn/tunnel/pkg/ippool"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type CachedStatistics struct {
	// PeersTotal is a number of peers
	// being authorized to connect to this node
	PeersTotal int
	// PeersWithTraffic is a number of peers
	// being actually connected to this node
	PeersWithTraffic int
	// Wireguard link statistic, may be nil
	LinkStat *netlink.LinkStatistics
}

type Manager struct {
	runtime       *runtime.TunnelRuntime
	mutex         sync.RWMutex
	storage       *storage.Storage
	wireguard     *wireguard.Wireguard
	ipv4pool      *ippool.IPv4pool
	eventLog      eventlog.EventManager
	running       bool
	bgStopChannel chan bool
	bgWaitGroup   sync.WaitGroup

	// statistic guarded by mutex and
	// updated by the backgroundOnce routine.
	statistic CachedStatistics
}

func New(runtime *runtime.TunnelRuntime, storage *storage.Storage, wireguard *wireguard.Wireguard, ipv4pool *ippool.IPv4pool, eventLog eventlog.EventManager) (*Manager, error) {
	manager := &Manager{
		runtime:       runtime,
		storage:       storage,
		wireguard:     wireguard,
		ipv4pool:      ipv4pool,
		eventLog:      eventLog,
		running:       true,
		bgStopChannel: make(chan bool),
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
