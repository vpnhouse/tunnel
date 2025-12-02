package manager

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/xstats"
	"github.com/vpnhouse/common-lib-go/xtime"
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

	updated       time.Time
	wgReceived    int64
	wgTransmitted int64
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
	if oldStats == nil {
		oldStats = &wgStats{}
	}
	newStats := make(wgStats)
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

		oldPeerStats := (*oldStats)[*peer.WireguardPublicKey]
		newPeerStats := manager.handlePeerStats(oldPeerStats, peer, wgPeer, now)
		newStats[*peer.WireguardPublicKey] = newPeerStats
	}

	manager.wgStats.Store(&newStats)

	// Delete expired peers
	for _, peer := range expiredPeers {
		err = manager.unsetPeer(peer)
		if err != nil {
			zap.L().Error("failed to unset expired peer", zap.Error(err))
		}
	}
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

func (manager *Manager) handlePeerStats(oldPeerStats PeerStats, peer *types.PeerInfo, wgPeer *wgtypes.Peer, now time.Time) (peerStats PeerStats) {
	peerStats = PeerStats{
		updated: now,
		Country: manager.peerCountry(peer, wgPeer),
	}

	dRx := wgPeer.ReceiveBytes - oldPeerStats.wgReceived
	dTx := wgPeer.TransmitBytes - oldPeerStats.wgTransmitted

	peerStats.Upstream += dTx
	peerStats.Downstream += dRx
	peerStats.wgReceived = wgPeer.ReceiveBytes
	peerStats.wgTransmitted = wgPeer.TransmitBytes

	if !oldPeerStats.updated.IsZero() {
		diffTimeMilli := now.Sub(oldPeerStats.updated).Milliseconds()
		if diffTimeMilli > 0 {
			peerStats.UpstreamSpeed = (dRx * 1000) / diffTimeMilli
			peerStats.DownstreamSpeed = (dRx * 1000) / diffTimeMilli
		} else {
			zap.L().Error("Negative time delta", zap.String("label", *peer.Label), zap.Time("updated", oldPeerStats.updated), zap.Time("now", now))
		}
	}

	if dRx == 0 && dTx == 0 {
		return
	}

	getOrZero := func(v *uuid.UUID) uuid.UUID {
		if v == nil {
			return uuid.UUID{}
		} else {
			return *v
		}
	}

	manager.statsReporter.ReportStats(getOrZero(peer.SessionId), uint64(dTx), uint64(dRx), func(_ uuid.UUID, out *xstats.SessionData) {
		out.Country = peerStats.Country
		if peer.InstallationId != nil {
			out.InstallationID = getOrZero(peer.InstallationId).String()
		}
		if peer.UserId != nil {
			out.UserID = *peer.UserId
		}
	})

	if peer.Upstream == nil {
		peer.Upstream = &dRx
	} else {
		*peer.Upstream += dRx
	}

	if peer.Downstream == nil {
		peer.Downstream = &dTx
	} else {
		*peer.Downstream += dTx
	}

	peer.Activity = xtime.FromTimePtr(&now)
	manager.storage.UpdatePeerStats(now, peer)

	return
}
