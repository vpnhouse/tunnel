package manager

import (
	"strings"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type wgStats map[string]PeerStats

type PeerStats struct {
	Upstream        int64  // bytes
	UpstreamSpeed   int64  // bytes per second
	Downstream      int64  // bytes
	DownstreamSpeed int64  // bytes per second
	Country         string // user country

	updated time.Time
}

func (manager *Manager) syncPeerStats() {
	wireguardPeers, err := manager.wireguard.GetPeers()
	if err != nil {
		return
	}

	peers, err := manager.peers()
	if err != nil {
		return
	}

	oldStats := manager.wgStats.Load()
	newStats := make(wgStats)
	updatedPeers := make([]*types.PeerInfo, 0)
	expiredPeers := make([]*types.PeerInfo, 0)
	now := time.Now()
	numPeersWithHadshakes := 0
	for _, peer := range peers {
		// Peer is expired - add it to the output list for later processing
		if peer.Expires != nil && peer.Expires.Time.Before(now) {
			expiredPeers = append(expiredPeers, peer)
		}

		if peer.Activity != nil {
			numPeersWithHadshakes++
		}

		if peer.WireguardPublicKey == nil {
			// We should never be here so it's added to be in safe
			zap.L().Error(
				"got a peer without the public key",
				zap.Any("id", peer.ID),
				zap.Any("user_id", peer.UserId),
				zap.Any("install_id", peer.InstallationId),
			)
			continue
		}

		wgPeer, ok := wireguardPeers[*peer.WireguardPublicKey]
		if !ok {
			zap.L().Error(
				"peer is presented in the manager's storage but not configured on the interface",
				zap.String("pub_key", *peer.WireguardPublicKey),
				zap.Any("id", peer.ID),
				zap.Any("user_id", peer.UserId),
				zap.Any("install_id", peer.InstallationId),
			)
			continue
		}

		newPeerStats, peerChanged := manager.handlePeerStats((*oldStats)[*peer.WireguardPublicKey], peer, wgPeer, now)
		if peerChanged {
			updatedPeers = append(updatedPeers, peer)
		}

		newStats[*peer.WireguardPublicKey] = newPeerStats
	}

	manager.wgStats.Store(&newStats)

	// Save stats of the updated peers
	for _, peer := range updatedPeers {
		// Store updated peers
		err = manager.storage.UpdatePeerStats(now, peer)
		if err != nil {
			zap.L().Error("failed to update peer stats", zap.Error(err))
			continue
		}
	}

	// Delete expired peers
	for _, peer := range expiredPeers {
		err = manager.unsetPeer(peer)
		if err != nil {
			zap.L().Error("failed to unset expired peer", zap.Error(err))
		}
	}

	peersWithHandshakesGauge.Set(float64(numPeersWithHadshakes))
}

func (manager *Manager) peerCountry(peer *types.PeerInfo, wgPeer *wgtypes.Peer) string {
	if manager.geoipService == nil || wgPeer.Endpoint == nil {
		return ""
	}

	country, err := manager.geoipService.GetCountry(wgPeer.Endpoint.IP)
	if err != nil {
		zap.L().Error("failed to detect country", zap.Stringp("peer", peer.Label))
	}
	country = strings.ToLower(country)

	return country
}

func (manager *Manager) handlePeerStats(oldPeerStats PeerStats, peer *types.PeerInfo, wgPeer *wgtypes.Peer, now time.Time) (PeerStats, bool) {
	peerStats := PeerStats{
		updated: now,
		Country: manager.peerCountry(peer, wgPeer),
	}

	diffUpstream := wgPeer.ReceiveBytes - oldPeerStats.Upstream
	diffDownstream := wgPeer.TransmitBytes - oldPeerStats.Downstream

	peerStats.Upstream += diffUpstream
	peerStats.Downstream += diffDownstream

	if !oldPeerStats.updated.IsZero() {
		diffTimeMilli := now.Sub(oldPeerStats.updated).Milliseconds()
		if diffTimeMilli > 0 {
			peerStats.UpstreamSpeed = (diffUpstream * 1000) / diffTimeMilli
			peerStats.DownstreamSpeed = (diffUpstream * 1000) / diffTimeMilli
		} else {
			zap.L().Error("Negative time delta", zap.String("label", *peer.Label), zap.Time("updated", oldPeerStats.updated), zap.Time("now", now))
		}
	}

	return peerStats, diffUpstream == 0 && diffDownstream == 0
}
