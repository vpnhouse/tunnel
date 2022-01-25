package manager

import (
	"errors"
	"time"

	libCommon "github.com/Codename-Uranium/common/common"
	"github.com/Codename-Uranium/common/proto"
	"github.com/Codename-Uranium/common/xnet"
	"github.com/Codename-Uranium/common/xtime"
	"github.com/Codename-Uranium/tunnel/internal/ippool"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (manager *Manager) peers(criteria *types.PeerInfo) ([]types.PeerInfo, error) {
	if criteria == nil {
		criteria = &types.PeerInfo{}
	}

	return manager.storage.SearchPeers(criteria)
}

// restore peers on startup
func (manager *Manager) restorePeers() {
	peers, err := manager.peers(nil)
	if err != nil {
		// err has already been logged inside
		return
	}

	for _, peer := range peers {
		if peer.Expired() {
			zap.L().Debug("Wiping expired peer", zap.Any("peer", peer))
			_ = manager.storage.DeletePeer(*peer.Id)
			continue
		}

		if err := manager.ipv4pool.Set(*peer.Ipv4); err != nil {
			if !errors.Is(err, ippool.ErrNotInRange) {
				continue
			}

			newIP, err := manager.ipv4pool.Alloc()
			if err != nil {
				// TODO(nikonov): remove peer OR mark it as invalid
				//  to allow further migration by hand.
				continue
			}
			peer.Ipv4 = newIP
			if _, err := manager.storage.UpdatePeer(&peer); err != nil {
				continue
			}
		}

		switch *peer.Type {
		case types.TunnelWireguard:
			_ = manager.wireguard.SetPeer(&peer)
			allPeersGauge.Inc()
		default:
			zap.L().Error("unsupported tunnel type", zap.Int("type", *peer.Type))
			continue
		}
	}
}

func (manager *Manager) unsetPeer(peer *types.PeerInfo) error {
	if peer == nil {
		return libCommon.EInternalError("peer info is nil", nil)
	}

	errManager := manager.storage.DeletePeer(*peer.Id)
	errWireguard := manager.wireguard.UnsetPeer(peer)
	errPool := manager.ipv4pool.Unset(*peer.Ipv4)

	// TODO(nikonov): report an actual traffic on remove
	allPeersGauge.Dec()
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerRemove), time.Now().Unix(), peer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerRemove)))
	}

	return func(errors ...error) error {
		for _, e := range errors {
			if e != nil {
				return e
			}
		}
		return nil
	}(errManager, errPool, errWireguard)
}

func (manager *Manager) setPeer(peer *types.PeerInfo) (*int64, error) {
	id, ipv4, err := func() (*int64, *xnet.IP, error) {
		if peer.Expired() {
			return nil, nil, libCommon.EInvalidArgument("peer already expired", nil)
		}

		if peer.Ipv4 == nil {
			// Allocate IP, if necessary
			ipv4, err := manager.ipv4pool.Alloc()
			if err != nil {
				return nil, nil, err
			}

			peer.Ipv4 = ipv4
		} else {
			// Check if IP can be used
			err := manager.ipv4pool.Set(*peer.Ipv4)
			if err != nil {
				return nil, nil, err
			}
		}

		// Create peer in storage
		id, err := manager.storage.CreatePeer(peer)
		if err != nil {
			return nil, peer.Ipv4, err
		}

		// Set peer in wireguard
		return id, peer.Ipv4, manager.wireguard.SetPeer(peer)
	}()

	if err != nil {
		if ipv4 != nil {
			_ = manager.ipv4pool.Unset(*ipv4)
		}

		if id != nil {
			_ = manager.storage.DeletePeer(*id)
		}

		return nil, err
	}

	allPeersGauge.Inc()
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerAdd), time.Now().Unix(), peer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerRemove)))
	}
	return id, nil
}

func (manager *Manager) updatePeer(newPeer *types.PeerInfo) error {
	if newPeer.Expired() {
		return manager.unsetPeer(newPeer)
	}

	// Find old peer to remove it from wireguard interface
	oldPeer, err := manager.storage.GetPeer(*newPeer.Id)

	if oldPeer == nil {
		if err != nil {
			return err
		} else {
			return libCommon.EEntryNotFound("peer not found", nil, zap.Any("newPeer", newPeer))
		}
	}

	if *oldPeer.Type != *newPeer.Type {
		return libCommon.EInvalidArgument("changing peer type is not allowed", nil, zap.Any("newPeer", newPeer), zap.Any("oldPeer", oldPeer))
	}

	if *newPeer.Type != types.TunnelWireguard {
		return libCommon.EInvalidArgument("updating this tunnel type is not supported yet", nil, zap.Any("newPeer", newPeer))
	}

	ipOK, dbOK, wgOK, err := func() (ipOK bool, dbOK bool, wgOK bool, err error) {
		// Prepare ipv4 address
		if newPeer.Ipv4 == nil {
			// IP is not set - allocate new one
			var ipv4 *xnet.IP
			ipv4, err = manager.ipv4pool.Alloc()
			if err != nil {
				// TODO: Differentiate log level by error type (i.e. no space is debug message, others are errors)
				zap.L().Debug("can't allocate new IP for existing peer", zap.Error(err))

				// Something went wrong - use old IP
				newPeer.Ipv4 = oldPeer.Ipv4
			} else {
				// Hurrah, we have new IP!
				newPeer.Ipv4 = ipv4
			}
		} else if !newPeer.Ipv4.Equal(oldPeer.Ipv4) {
			// Try to set up new ip, if it differs from old one
			err = manager.ipv4pool.Set(*newPeer.Ipv4)
			if err != nil {
				return
			}
		}

		// We finished IP updating
		ipOK = true

		// Update database
		now := xtime.Now()
		newPeer.Updated = &now
		_, err = manager.storage.UpdatePeer(newPeer)
		if err != nil {
			return
		}

		// We finished database updating
		dbOK = true

		// Update wireguard peer
		if *oldPeer.WireguardPublicKey != *newPeer.WireguardPublicKey {
			// Key changed - we need remove old peer and set new
			err = manager.wireguard.UnsetPeer(oldPeer)
			if err != nil {
				return
			}
		}

		err = manager.wireguard.SetPeer(newPeer)
		if err != nil {
			zap.L().Error("failed to set new peer, trying to revert old", zap.Error(err))
			err = manager.wireguard.SetPeer(oldPeer)
			return
		}

		wgOK = true
		return
	}()

	// Reverting back
	if err != nil {
		if dbOK {
			// Try to revert peer state
			_, _ = manager.storage.UpdatePeer(oldPeer)
		}

		if ipOK && !newPeer.Ipv4.Equal(oldPeer.Ipv4) {
			// Try to cleanup new IP
			_ = manager.ipv4pool.Unset(*newPeer.Ipv4)
		}

		if wgOK {
			// Try to revert wireguard peer
			_ = manager.wireguard.UnsetPeer(newPeer)
			_ = manager.wireguard.SetPeer(oldPeer)
		}

		return err
	}

	// TODO(nikonov): report an actual traffic on update
	if err := manager.eventLog.Push(uint32(proto.EventType_PeerUpdate), time.Now().Unix(), newPeer.IntoProto()); err != nil {
		// do not return an error here because it's not related to the method itself.
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerRemove)))
	}
	return nil
}

func (manager *Manager) findPeerByIdentifiers(identifiers *types.PeerIdentifiers) (*types.PeerInfo, error) {
	if identifiers == nil {
		return nil, libCommon.EInvalidArgument("no identifiers", nil)
	}

	peerQuery := types.PeerInfo{
		PeerIdentifiers: *identifiers,
	}

	peers, err := manager.storage.SearchPeers(&peerQuery)
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, libCommon.EEntryNotFound("peer not found", nil)
	}

	if len(peers) > 1 {
		return nil, libCommon.EInvalidArgument("not enough identifiers to update peer", nil)
	}

	return &peers[0], nil
}

func (manager *Manager) lock() error {
	manager.mutex.Lock()
	if !manager.running {
		manager.mutex.Unlock()
		return libCommon.EUnavailable("server is shutting down", nil)
	}

	return nil
}

func (manager *Manager) unlock() {
	manager.mutex.Unlock()
}

func (manager *Manager) backgroundOnce() {
	if err := manager.lock(); err != nil {
		return
	}
	defer manager.unlock()

	linkStats, err := manager.wireguard.GetLinkStatistic()
	if err == nil {
		// non-nil error will be logged
		// by the common.Error inside the method.
		updatePrometheusFromLinkStats(linkStats)
	}

	// ignore error because it logged by the common.Error wrapper.
	// it is safe to call reportTrafficByPeer with nil map.
	wireguardPeers, _ := manager.wireguard.GetPeers()
	peersTotal := 0
	withHandshakes := 0

	peers, err := manager.peers(nil)
	if err != nil {
		return
	}

	for _, peer := range peers {
		if peer.Expired() {
			_ = manager.unsetPeer(&peer)
			continue
		}

		peersTotal++
		wgPeer, ok := findWgPeerByPublicKey(&peer, wireguardPeers)
		if ok {
			// no handshake - no traffic, avoid spamming empty events to the log
			if !wgPeer.LastHandshakeTime.IsZero() {
				manager.reportPeerTraffic(&peer, wgPeer)
				withHandshakes++
			}
		}
	}

	manager.statistic = CachedStatistics{
		PeersTotal:       peersTotal,
		PeersWithTraffic: withHandshakes,
		LinkStat:         linkStats,
	}

	peersWithHandshakesGauge.Set(float64(withHandshakes))
}

func (manager *Manager) background() {
	defer manager.bgWaitGroup.Done()

	// TODO (Sergey Kovalev): Move interval to settings
	expirationTicker := time.NewTicker(time.Second * 60)
	defer expirationTicker.Stop()

	for {
		select {
		case <-manager.bgStopChannel:
			zap.L().Info("Shutting down manager background process")
			return
		case <-expirationTicker.C:
			zap.L().Debug("Running expiration round")
			manager.backgroundOnce()
		}
	}
}

// findWgPeerByPublicKey returns wireguard peer for matching peer public key, if any.
func findWgPeerByPublicKey(peer *types.PeerInfo, wgPeers map[string]wgtypes.Peer) (wgtypes.Peer, bool) {
	// make it safe to call with empty or nil map
	if len(wgPeers) == 0 {
		return wgtypes.Peer{}, false
	}

	if peer.WireguardPublicKey == nil {
		// should this ever happen?
		// why we even have this as string *pointer*?
		zap.L().Error("got a peer without the public key",
			zap.Any("id", peer.Id),
			zap.Any("user_id", peer.UserId),
			zap.Any("install_id", peer.InstallationId))
		return wgtypes.Peer{}, false
	}

	key := *peer.WireguardPublicKey
	wgPeer, ok := wgPeers[key]
	if !ok {
		zap.L().Error("peer is presented in the manager's storage but not configured on the interface",
			zap.String("pub_key", *peer.WireguardPublicKey),
			zap.Any("id", peer.Id),
			zap.Any("user_id", peer.UserId),
			zap.Any("install_id", peer.InstallationId))
		return wgtypes.Peer{}, false
	}

	return wgPeer, true
}

// reportPeerTraffic reports peer's rx/tx traffic to the eventlog.
func (manager *Manager) reportPeerTraffic(peer *types.PeerInfo, wgPeer wgtypes.Peer) {
	info := peer.IntoProto()
	info.BytesTx = uint64(wgPeer.TransmitBytes)
	info.BytesRx = uint64(wgPeer.ReceiveBytes)

	if err := manager.eventLog.Push(uint32(proto.EventType_PeerTraffic), time.Now().Unix(), &info); err != nil {
		zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerRemove)))
	}
}
