package manager

import (
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type peerChangeType int
type peerChangeSumary int

const (
	peerChangeNone          peerChangeType = 0
	peerChangeFirstActivity peerChangeType = 1
	peerChangeActivity      peerChangeType = 2
	peerChangeTraffic       peerChangeType = 4
)

func (s peerChangeSumary) HasChanges() bool {
	return int(s) != int(peerChangeNone)
}

func (s peerChangeSumary) Has(t peerChangeType) bool {
	return (int(s) & int(t)) == int(t)
}

func (s *peerChangeSumary) Set(t peerChangeType) {
	*s = peerChangeSumary(int(*s) | int(t))
}

// Keeps accumulated peer counters updated for certian wireguard peer
type wireguardStats struct {
	Upstream   int64
	Downstream int64
}

type updatePeerStatsResults struct {
	UpdatedPeers           []*types.PeerInfo
	ExpiredPeers           []*types.PeerInfo
	FirstConnectedPeers    []*types.PeerInfo
	TrafficUpdatedPeers    []*types.PeerInfo
	NumPeersWithHadshakes  int
	NumPeersActiveLastHour int
	NumPeersActiveLastDay  int
	NumPeers               int
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
		UpdatedPeers:        make([]*types.PeerInfo, 0, len(peers)),
		ExpiredPeers:        make([]*types.PeerInfo, 0, len(peers)),
		FirstConnectedPeers: make([]*types.PeerInfo, 0, len(peers)),
	}

	now := time.Now()

	for _, peer := range peers {
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
			// Remove peer stats in case it's gone
			if _, ok := s.stats[*peer.WireguardPublicKey]; ok {
				delete(s.stats, *peer.WireguardPublicKey)
			}
			continue
		}

		// Update peer stats and add peer to the update peers list for futher processing
		changes := s.updatePeerStatsFromWgPeer(wgPeer, &peer)
		if changes.HasChanges() {
			results.UpdatedPeers = append(results.UpdatedPeers, &peer)
		}

		if changes.Has(peerChangeFirstActivity) {
			results.FirstConnectedPeers = append(results.FirstConnectedPeers, &peer)
		}

		if changes.Has(peerChangeTraffic) {
			results.TrafficUpdatedPeers = append(results.TrafficUpdatedPeers, &peer)
		}

		if peer.Activity != nil {
			results.NumPeersWithHadshakes++
			lastActiveDeltaHours := now.Sub(peer.Activity.Time).Hours()
			if lastActiveDeltaHours < 1 {
				results.NumPeersActiveLastHour++
			}
			if lastActiveDeltaHours < 24 {
				results.NumPeersActiveLastDay++
			}
		}

		// Peer is expired - add it to the output list for later processing
		if peer.Expires != nil && peer.Expires.Time.Before(now) {
			results.ExpiredPeers = append(results.ExpiredPeers, &peer)
		}
	}

	// Finally snap the current number of available peers
	results.NumPeers = len(s.stats)

	return results
}

func (s *peerStatsService) updatePeerStatsFromWgPeer(wgPeer wgtypes.Peer, peer *types.PeerInfo) peerChangeSumary {
	var changeSum peerChangeSumary

	if peer.WireguardPublicKey == nil {
		return changeSum
	}

	if !wgPeer.LastHandshakeTime.IsZero() {
		if peer.Activity == nil || peer.Activity.Time.Unix() < wgPeer.LastHandshakeTime.Unix() {
			if peer.Activity == nil {
				changeSum.Set(peerChangeFirstActivity)
			}
			changeSum.Set(peerChangeActivity)
			peer.Activity = xtime.FromTimePtr(&wgPeer.LastHandshakeTime)
		}
	}

	stats, ok := s.stats[*peer.WireguardPublicKey]
	if !ok {
		stats = &wireguardStats{}
		s.stats[*peer.WireguardPublicKey] = stats
	}

	if wgPeer.ReceiveBytes > stats.Upstream {
		// Upstream never be nil
		*peer.Upstream += (wgPeer.ReceiveBytes - stats.Upstream)
		changeSum.Set(peerChangeTraffic)
	}

	if wgPeer.TransmitBytes > stats.Downstream {
		// Downstream never be nil
		*peer.Downstream += (wgPeer.TransmitBytes - stats.Downstream)
		changeSum.Set(peerChangeTraffic)
	}

	stats.Upstream = wgPeer.ReceiveBytes
	stats.Downstream = wgPeer.TransmitBytes

	return changeSum
}
