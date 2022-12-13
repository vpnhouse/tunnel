// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package manager

import (
	"database/sql"
	"errors"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xtime"
)

func (manager *Manager) SetPeer(info *types.PeerInfo) error {
	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// note: manager.setPeer changes given struct
	err := manager.setPeer(info)
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}

func (manager *Manager) UpdatePeer(info *types.PeerInfo) error {
	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()
	err := manager.updatePeer(info)
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}

func (manager *Manager) GetPeer(id int64) (*types.PeerInfo, error) {
	if !manager.running.Load().(bool) {
		return nil, xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	return manager.storage.GetPeer(id)
}

func (manager *Manager) UnsetPeer(id int64) error {
	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	info, err := manager.storage.GetPeer(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	err = manager.unsetPeer(info)
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}

func (manager *Manager) UnsetPeerByIdentifiers(identifiers *types.PeerIdentifiers) error {
	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	info, err := manager.findPeerByIdentifiers(identifiers)
	if err != nil {
		return err
	}

	err = manager.unsetPeer(info)
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}

func (manager *Manager) ListPeers() ([]*types.PeerInfo, error) {
	if !manager.running.Load().(bool) {
		return nil, xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	return manager.storage.SearchPeers(nil)
}

func (manager *Manager) ConnectPeer(info *types.PeerInfo) error {
	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	oldPeerShadow := types.PeerInfo{
		PeerIdentifiers: types.PeerIdentifiers{
			UserId:         info.UserId,
			InstallationId: info.InstallationId,
		},
	}

	oldPeers, err := manager.storage.SearchPeers(&oldPeerShadow)
	if err != nil {
		return err
	}

	if len(oldPeers) == 0 {
		err = manager.setPeer(info)
		if err != nil {
			return err
		}
		manager.syncPeerStats()
		return nil
	}

	if len(oldPeers) > 1 {
		return xerror.EInternalError("too many peers for identifiers", nil)
	}

	info.ID = oldPeers[0].ID
	info.Ipv4 = oldPeers[0].Ipv4

	err = manager.updatePeer(info)
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}

func (manager *Manager) UpdatePeerExpiration(identifiers *types.PeerIdentifiers, expires *time.Time) error {
	if identifiers == nil {
		return xerror.EInvalidArgument("no identifiers", nil)
	}

	if !manager.running.Load().(bool) {
		return xerror.EUnavailable("server is shutting down", nil)
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()

	peerQuery := types.PeerInfo{
		PeerIdentifiers: *identifiers,
	}

	peers, err := manager.storage.SearchPeers(&peerQuery)
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		return xerror.EEntryNotFound("peer not found", nil)
	}

	if len(peers) > 1 {
		return xerror.EInvalidArgument("not enough identifiers to update peer", nil)
	}

	peers[0].Expires = xtime.FromTimePtr(expires)
	err = manager.updatePeer(peers[0])
	if err != nil {
		return err
	}
	manager.syncPeerStats()
	return nil
}
