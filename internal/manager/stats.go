package manager

import (
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/human"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type peerChangeType int
type peerChangeSummary int

const (
	peerChangeNone          peerChangeType = 0
	peerChangeFirstActivity peerChangeType = 1
	peerChangeActivity      peerChangeType = 2
	peerChangeTraffic       peerChangeType = 4
)

func (s peerChangeSummary) HasAnyChanges() bool {
	return int(s) != int(peerChangeNone)
}

func (s peerChangeSummary) Has(t peerChangeType) bool {
	return (int(s) & int(t)) == int(t)
}

func (s *peerChangeSummary) Set(t peerChangeType) {
	*s = peerChangeSummary(int(*s) | int(t))
}

// Keeps accumulated peer counters updated for certian wireguard peer
type wireguardStats struct {
	Updated    int64 // timestamp in seconds
	Upstream   int64 // bytes
	Downstream int64 // bytes

	// Recalculated on update
	upstreamSpeed   int64 // bytes per second
	downstreamSpeed int64 // bytes per second
}

func (s *wireguardStats) Update(now time.Time, upstream int64, downstream int64) {
	ts := now.Unix()
	defer func() {
		s.Upstream = upstream
		s.Downstream = downstream
		s.Updated = ts
	}()

	if ts <= s.Updated || s.Updated == 0 {
		return
	}

	seconds := ts - s.Updated

	if upstream >= s.Upstream {
		s.upstreamSpeed = (upstream - s.Upstream) / seconds
	}

	if downstream >= s.Downstream {
		s.downstreamSpeed = (downstream - s.Downstream) / seconds
	}
}

func (s *wireguardStats) LastUpstreamSpeed(updateInterval human.Interval) int64 {
	// Return speed only if stats was initialized and update time is out of given
	if s.Updated != 0 || s.Updated+int64(updateInterval.Value().Seconds())*2 < time.Now().Unix() {
		return 0
	}
	return s.upstreamSpeed
}

func (s *wireguardStats) LastDownstreamSpeed(updateInterval human.Interval) int64 {
	// Return speed only if stats was initialized and update time is out of given
	if s.Updated != 0 || s.Updated+int64(updateInterval.Value().Seconds())*2 < time.Now().Unix() {
		return 0
	}
	return s.downstreamSpeed
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

func (s *peerStatsService) GetPeerSpeeds(updateInterval human.Interval, peer *types.PeerInfo) (int64, int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	peerStats, ok := s.stats[*peer.WireguardPublicKey]
	if !ok {
		return 0, 0
	}
	return peerStats.LastUpstreamSpeed(updateInterval), peerStats.LastDownstreamSpeed(updateInterval)
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
		if changes.HasAnyChanges() {
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

func (s *peerStatsService) updatePeerStatsFromWgPeer(wgPeer wgtypes.Peer, peer *types.PeerInfo) peerChangeSummary {
	var changeSum peerChangeSummary

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
