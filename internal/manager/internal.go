// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"errors"
	"time"

	"github.com/vpnhouse/common-lib-go/ippool"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xtime"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func (manager *Manager) peers() ([]*types.PeerInfo, error) {
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
	}
}

func (manager *Manager) unsetPeer(peer *types.PeerInfo) error {
	err := manager.storage.DeletePeer(peer.ID)
	errs := multierr.Append(nil, err)

	err = manager.wireguard.UnsetPeer(peer)
	errs = multierr.Append(errs, err)

	err = manager.ip4am.Unset(*peer.Ipv4)
	errs = multierr.Append(errs, err)
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
		if err := manager.wireguard.SetPeer(peer); err != nil {
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

	return nil
}

// updatePeer changes given newPeer,
// fields: ID, IPv4
func (manager *Manager) updatePeer(newPeer *types.PeerInfo) error {
	if newPeer.Expired() {
		return manager.unsetPeer(newPeer)
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
		id, err := manager.storage.UpdatePeer(newPeer)
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

		if err := manager.wireguard.SetPeer(newPeer); err != nil {
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
			_ = manager.wireguard.UnsetPeer(newPeer)
			_ = manager.wireguard.SetPeer(oldPeer)
		}

		return err
	}

	return nil
}

func (manager *Manager) findPeerByIdentifiers(identifiers *types.PeerIdentifiers) (*types.PeerInfo, error) {
	if identifiers == nil {
		return nil, xerror.EInvalidArgument("no identifiers", nil)
	}

	peerQuery := types.PeerInfo{
		PeerIdentifiers: *identifiers,
	}

	peers, err := manager.storage.SearchPeers(&peerQuery)
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, xerror.EEntryNotFound("peer not found", nil)
	}

	if len(peers) > 1 {
		return nil, xerror.EInvalidArgument("not enough identifiers to update peer", nil)
	}

	return peers[0], nil
}

func (manager *Manager) sync() {
	manager.syncPeerStats()
}

func (manager *Manager) background() {
	syncPeerTicker := time.NewTicker(manager.runtime.Settings.GetUpdateStatisticsInterval().Value())
	zap.L().Debug("Start update peer stats", zap.Stringer("interval", manager.runtime.Settings.GetUpdateStatisticsInterval()))

	defer func() {
		syncPeerTicker.Stop()
		close(manager.done)
	}()

	manager.lock.Lock()
	manager.sync()
	manager.lock.Unlock()

	for {
		select {
		case <-manager.stop:
			zap.L().Info("Shutting down manager background process")
			return
		case <-syncPeerTicker.C:
			manager.lock.Lock()
			manager.sync()
			manager.lock.Unlock()
		}
	}
}
