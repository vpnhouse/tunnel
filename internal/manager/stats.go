package manager

import (
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Keeps accumulated peer counters updated for certian wireguard peer
type wireguardStats struct {
	Upstream   int64
	Downstream int64
}

type updatePeerStatsResults struct {
	UpdatedPeers   []*types.PeerInfo
	ExpiredPeers   []*types.PeerInfo
	NumPeersWithHadshakes int
	NumPeersActiveLastHour int
	NumPeersActiveLastDay int
	NumPeers int
}

type peerStatsService struct {
	lock sync.Mutex
	// {peer public key} -> peerStats
	stats map[string]*wireguardStats
	once  sync.Once
}

func (s *peerStatsService) init() {
	s.stats = make(map[string]*wireguardStats, 1000)
}

func (s *peerStatsService) UpdatePeerStats(peers []types.PeerInfo, wireguardPeers map[string]wgtypes.Peer) updatePeerStatsResults {
	s.once.Do(s.init)

	s.lock.Lock()
	defer s.lock.Unlock()

	results := updatePeerStatsResults{
		UpdatedPeers: make([]*types.PeerInfo, 0, len(peers)),
		ExpiredPeers: make([]*types.PeerInfo, 0, len(peers)),
	}

	now := time.Now()

	for _, peer := range peers {
		if peer.Expires != nil && peer.Expires.Time.Before(now) {
			results.ExpiredPeers = append(results.ExpiredPeers, &peer)
			continue
		}

		if peer.WireguardPublicKey == nil {
			continue
		}

		wgPeer, ok := wireguardPeers[*peer.WireguardPublicKey]
		if !ok {
			if _, ok := s.stats[*peer.WireguardPublicKey]; ok {
				delete(s.stats, *peer.WireguardPublicKey)
			}
			continue
		}

		if s.updatePeerStatsFromWgPeer(wgPeer, &peer) {
			results.UpdatedPeers = append(results.UpdatedPeers, &peer)
		}

		if peer.Activity != nil {
			results.NumPeersWithHadshakes++
			lastActiveDeltaHours := now.Sub(peer.Activity.Time)
			if lastActiveDeltaHours < 1 {
				results.NumPeersActiveLastHour++
			}
			if lastActiveDeltaHours < 24 {
				results.NumPeersActiveLastDay++
			}
		}
	}

	results.NumPeers = len(s.stats)

	return results
}

func (s *peerStatsService) updatePeerStatsFromWgPeer(wgPeer wgtypes.Peer, peer *types.PeerInfo) bool {
	if peer.WireguardPublicKey == nil {
		return false
	}

	isUpdated := false
	if !wgPeer.LastHandshakeTime.IsZero() {
		if peer.Activity == nil || peer.Activity.Time.Unix() < wgPeer.LastHandshakeTime.Unix() {
			peer.Activity = xtime.FromTimePtr(&wgPeer.LastHandshakeTime)
			isUpdated = true
		}
	}

	stats, ok := s.stats[*peer.WireguardPublicKey]
	if !ok {
		stats = &wireguardStats{}
		s.stats[*peer.WireguardPublicKey] = stats
	}

	if wgPeer.ReceiveBytes > stats.Upstream {
		peer.Upstream += (wgPeer.ReceiveBytes - stats.Upstream)
		isUpdated = true
	}

	if wgPeer.TransmitBytes > stats.Downstream {
		peer.Downstream += (wgPeer.TransmitBytes - stats.Downstream)
		isUpdated = true
	}

	stats.Upstream = wgPeer.ReceiveBytes
	stats.Downstream = wgPeer.TransmitBytes

	return isUpdated
}
