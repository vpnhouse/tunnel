// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"errors"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/ippool"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func (manager *Manager) peers() ([]types.PeerInfo, error) {
	return manager.storage.SearchPeers(nil)
}

// restore peers on startup
func (manager *Manager) restorePeers() {
	peers, err := manager.peers()
	if err != nil {
		// err has already been logged inside
		return
	}

	for _, peer := range peers {
		if peer.Expired() {
			zap.L().Debug("wiping expired peer", zap.Any("peer", peer))
			_ = manager.storage.DeletePeer(peer.ID)
			continue
		}

		if err := manager.ip4am.Set(*peer.Ipv4, peer.GetNetworkPolicy()); err != nil {
			if !errors.Is(err, ippool.ErrNotInRange) {
				continue
			}

			newIP, err := manager.ip4am.Alloc(peer.GetNetworkPolicy())
			if err != nil {
				// TODO(nikonov): remove peer OR mark it as invalid
				//  to allow further migration by hand.
				continue
			}
			peer.Ipv4 = &newIP
			if _, err := manager.storage.UpdatePeer(peer); err != nil {
				continue
			}
		}

		_ = manager.wireguard.SetPeer(peer)
		allPeersGauge.Inc()
		manager.peerTrafficSender.Add(&peer)
	}
}

func (manager *Manager) unsetPeer(peer types.PeerInfo) error {
	err := manager.storage.DeletePeer(peer.ID)
	errs := multierr.Append(nil, err)

	err = manager.wireguard.UnsetPeer(peer)
	errs = multierr.Append(errs, err)

	err = manager.ip4am.Unset(*peer.Ipv4)
	errs = multierr.Append(errs, err)

	allPeersGauge.Dec()
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerRemove), time.Now().Unix(), peer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerRemove)))
	}

	manager.peerTrafficSender.Remove(&peer)

	return errs
}

// setPeer changes the given PeerInfo,
// fields: ID, IPv4
func (manager *Manager) setPeer(peer *types.PeerInfo) error {
	err := func() error {
		if peer.Expired() {
			return xerror.EInvalidArgument("peer already expired", nil)
		}

		if peer.Ipv4 == nil || peer.Ipv4.IP == nil {
			// Allocate IP, if necessary
			ipv4, err := manager.ip4am.Alloc(peer.GetNetworkPolicy())
			if err != nil {
				return err
			}

			peer.Ipv4 = &ipv4
		} else {
			// Check if IP can be used
			err := manager.ip4am.Set(*peer.Ipv4, peer.GetNetworkPolicy())
			if err != nil {
				return err
			}
		}

		// Set counters to zeros to prevent any fails on update stats operation
		if peer.Upstream == nil {
			var zeroVal int64
			peer.Upstream = &zeroVal
		}
		if peer.Downstream == nil {
			var zeroVal int64
			peer.Downstream = &zeroVal
		}

		// Create peer in storage
		id, err := manager.storage.CreatePeer(*peer)
		if err != nil {
			return err
		}
		peer.ID = id

		// Set peer in wireguard
		if err := manager.wireguard.SetPeer(*peer); err != nil {
			return err
		}

		return nil
	}()

	// rollback an action on error
	if err != nil {
		if peer.Ipv4 != nil {
			_ = manager.ip4am.Unset(*peer.Ipv4)
		}

		if peer.ID > 0 {
			_ = manager.storage.DeletePeer(peer.ID)
		}

		return err
	}

	allPeersGauge.Inc()
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerAdd), time.Now().Unix(), peer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerAdd)))
	}
	manager.peerTrafficSender.Add(peer)

	return nil
}

// updatePeer changes given newPeer,
// fields: ID, IPv4
func (manager *Manager) updatePeer(newPeer *types.PeerInfo) error {
	if newPeer.Expired() {
		return manager.unsetPeer(*newPeer)
	}

	// Find old peer to remove it from wireguard interface
	oldPeer, err := manager.storage.GetPeer(newPeer.ID)
	if err != nil {
		return err
	}

	ipOK, dbOK, wgOK, err := func() (bool, bool, bool, error) {
		var ipOK, dbOK, wgOK bool
		// Prepare ipv4 address
		if newPeer.Ipv4 == nil {
			// IP is not set - allocate new one
			ipv4, err := manager.ip4am.Alloc(newPeer.GetNetworkPolicy())
			if err != nil {
				// TODO: Differentiate log level by error type (i.e. no space is debug message, others are errors)
				zap.L().Debug("can't allocate new IP for existing peer", zap.Error(err))

				// Something went wrong - use old IP
				newPeer.Ipv4 = oldPeer.Ipv4
			} else {
				// Hurrah, we have new IP!
				newPeer.Ipv4 = &ipv4
			}
		} else if !newPeer.Ipv4.Equal(*oldPeer.Ipv4) {
			// Try to set up new ip, if it differs from old one
			if err := manager.ip4am.Set(*newPeer.Ipv4, newPeer.GetNetworkPolicy()); err != nil {
				return ipOK, dbOK, wgOK, err
			}
		}

		// We finished IP updating
		ipOK = true

		// Update database
		now := xtime.Now()
		newPeer.Updated = &now
		id, err := manager.storage.UpdatePeer(*newPeer)
		if err != nil {
			return ipOK, dbOK, wgOK, err
		}
		// We finished database updating
		newPeer.ID = id
		dbOK = true

		// Update wireguard peer
		if *oldPeer.WireguardPublicKey != *newPeer.WireguardPublicKey {
			// Key changed - we need remove old peer and set new
			if err := manager.wireguard.UnsetPeer(oldPeer); err != nil {
				return ipOK, dbOK, wgOK, err
			}
		}

		if err := manager.wireguard.SetPeer(*newPeer); err != nil {
			zap.L().Error("failed to set new peer, trying to revert old", zap.Error(err))
			err = manager.wireguard.SetPeer(oldPeer)
			return ipOK, dbOK, wgOK, err
		}

		wgOK = true
		return ipOK, dbOK, wgOK, err
	}()

	// Reverting back
	if err != nil {
		if dbOK {
			// Try to revert peer state
			_, _ = manager.storage.UpdatePeer(oldPeer)
		}

		if ipOK && !newPeer.Ipv4.Equal(*oldPeer.Ipv4) {
			// Try to cleanup new IP
			_ = manager.ip4am.Unset(*newPeer.Ipv4)
		}

		if wgOK {
			// Try to revert wireguard peer
			_ = manager.wireguard.UnsetPeer(*newPeer)
			_ = manager.wireguard.SetPeer(oldPeer)
		}

		return err
	}

	// TODO(nikonov): report an actual traffic on update
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerUpdate), time.Now().Unix(), newPeer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerUpdate)))
	}

	return nil
}

func (manager *Manager) findPeerByIdentifiers(identifiers *types.PeerIdentifiers) (types.PeerInfo, error) {
	if identifiers == nil {
		return types.PeerInfo{}, xerror.EInvalidArgument("no identifiers", nil)
	}

	peerQuery := types.PeerInfo{
		PeerIdentifiers: *identifiers,
	}

	peers, err := manager.storage.SearchPeers(&peerQuery)
	if err != nil {
		return types.PeerInfo{}, err
	}

	if len(peers) == 0 {
		return types.PeerInfo{}, xerror.EEntryNotFound("peer not found", nil)
	}

	if len(peers) > 1 {
		return types.PeerInfo{}, xerror.EInvalidArgument("not enough identifiers to update peer", nil)
	}

	return peers[0], nil
}

func (manager *Manager) syncPeerStats() {
	linkStats, err := manager.wireguard.GetLinkStatistic()
	if err == nil {
		// non-nil error will be logged
		// by the common.Error inside the method.
		updatePrometheusFromLinkStats(linkStats)
	}

	// ignore error because it logged by the common.Error wrapper.
	// it is safe to call reportTrafficByPeer with nil map.
	wireguardPeers, _ := manager.wireguard.GetPeers()

	peers, err := manager.peers()
	if err != nil {
		return
	}

	// Update peer stats according to current metrics in wireguard peers
	results := manager.statsService.UpdatePeerStats(peers, wireguardPeers)

	// Save stats of the updated peers
	for _, peer := range results.UpdatedPeers {
		// Store updated peers
		err = manager.storage.UpdatePeerStats(peer)
		if err != nil {
			zap.L().Error("failed to update peer stats", zap.Error(err))
			continue
		}
	}

	// Send notifications about peers with first connection
	for _, peer := range results.FirstConnectedPeers {
		// Send event containing updated peer
		err := manager.eventLog.Push(uint32(proto.EventType_PeerFirstConnect), time.Now().Unix(), peer.IntoProto())
		if err != nil {
			zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerFirstConnect)))
		}
	}

	// Notify with the peers with traffic updates
	manager.peerTrafficSender.Send(results.TrafficUpdatedPeers)

	// Delete expired peers
	for _, peer := range results.ExpiredPeers {
		err = manager.unsetPeer(*peer)
		if err != nil {
			zap.L().Error("failed to unset expired peer", zap.Error(err))
		}
	}

	oldStats := manager.GetCachedStatistics()

	diffUpstream := linkStats.RxBytes
	diffDownstream := linkStats.TxBytes
	if oldStats.LinkStat != nil {
		diffUpstream -= oldStats.LinkStat.RxBytes
		diffDownstream -= oldStats.LinkStat.TxBytes
	}

	newStats := &CachedStatistics{
		PeersTotal:          results.NumPeers,
		PeersWithTraffic:    results.NumPeersWithHadshakes,
		PeersActiveLastHour: results.NumPeersActiveLastHour,
		PeersActiveLastDay:  results.NumPeersActiveLastDay,
		LinkStat:            linkStats,
		Upstream:            oldStats.Upstream + int64(diffUpstream),
		Downstream:          oldStats.Downstream + int64(diffDownstream),
		Collected:           time.Now().Unix(),
	}

	newStats.UpdateSpeeds(oldStats)

	zap.L().Info("STATS",
		zap.Int("total", results.NumPeers),
		zap.Int("connected", results.NumPeersWithHadshakes),
		zap.Int("active_1h", results.NumPeersActiveLastHour),
		zap.Int("active_1d", results.NumPeersActiveLastDay),
		zap.Int("rx_bytes", int(linkStats.RxBytes)),
		zap.Int("rx_packets", int(linkStats.RxPackets)),
		zap.Int("tx_bytes", int(linkStats.TxBytes)),
		zap.Int("tx_packets", int(linkStats.TxPackets)))

	peersWithHandshakesGauge.Set(float64(results.NumPeersWithHadshakes))
	manager.storage.SetUpstreamMetric(newStats.Upstream)
	manager.storage.SetDownstreamMetric(newStats.Downstream)

	manager.statistic.Store(newStats)
}

func (manager *Manager) background() {
	syncPeerTicker := time.NewTicker(manager.runtime.Settings.GetUpdateStatisticsInterval().Value())
	zap.L().Debug("Start update peer stats", zap.Stringer("interval", manager.runtime.Settings.GetUpdateStatisticsInterval()))

	defer func() {
		syncPeerTicker.Stop()
		close(manager.done)
	}()

	manager.lock.Lock()
	manager.syncPeerStats()
	manager.lock.Unlock()

	for {
		select {
		case <-manager.stop:
			zap.L().Info("Shutting down manager background process")
			return
		case <-syncPeerTicker.C:
			manager.lock.Lock()
			manager.syncPeerStats()
			manager.lock.Unlock()
		}
	}
}
