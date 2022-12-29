package manager

import (
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/statutils"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type peerChangeType int
type peerChangeSummary int

type speedValue struct {
	Upstream   int64
	Downstream int64
}

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
type runtimePeerStat struct {
	Updated         int64 // timestamp in seconds
	Upstream        int64 // bytes
	UpstreamSpeed   int64 // bytes per second
	Downstream      int64 // bytes
	DownstreamSpeed int64 // bytes per second

	upstreamSpeedAvg   *statutils.AvgValue
	downstreamSpeedAvg *statutils.AvgValue
}

func newRuntimePeerStat() *runtimePeerStat {
	return &runtimePeerStat{
		upstreamSpeedAvg:   statutils.NewAvgValue(10),
		downstreamSpeedAvg: statutils.NewAvgValue(10),
	}
}

func (s *runtimePeerStat) Update(now time.Time, upstream int64, downstream int64) {
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
	if seconds == 0 {
		return
	}

	if upstream >= s.Upstream {
		s.UpstreamSpeed = s.upstreamSpeedAvg.Push(((upstream - s.Upstream) / seconds))
	}

	if downstream >= s.Downstream {
		s.DownstreamSpeed = s.downstreamSpeedAvg.Push((downstream - s.Downstream) / seconds)
	}
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

type runtimePeerStatsService struct {
	lock sync.Mutex
	// {peer public key} -> peerStats
	stats map[string]*runtimePeerStat
	once  sync.Once
}

func (s *runtimePeerStatsService) init() {
	s.stats = make(map[string]*runtimePeerStat, 1000)
}

func (s *runtimePeerStatsService) GetRuntimePeerStat(peer *types.PeerInfo) runtimePeerStat {
	s.lock.Lock()
	defer s.lock.Unlock()
	stat, ok := s.stats[*peer.WireguardPublicKey]
	if !ok || stat == nil {
		return runtimePeerStat{}
	}
	return *stat
}

func (s *runtimePeerStatsService) UpdatePeersStats(peers []*types.PeerInfo, wireguardPeers map[string]wgtypes.Peer) updatePeerStatsResults {
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
		changes := s.updateRuntimePeerStatFromWireguardPeer(now, wgPeer, peer)
		if changes.HasAnyChanges() {
			results.UpdatedPeers = append(results.UpdatedPeers, peer)
		}

		if changes.Has(peerChangeFirstActivity) {
			results.FirstConnectedPeers = append(results.FirstConnectedPeers, peer)
		}

		if changes.Has(peerChangeTraffic) {
			results.TrafficUpdatedPeers = append(results.TrafficUpdatedPeers, peer)
		}

		if peer.Activity != nil {
			results.NumPeersWithHadshakes++
			lastActiveDeltaHours := now.Sub(peer.Activity.Time).Hours()
			zap.L().Debug(
				"peer data",
				zap.Stringp("label", peer.Label),
				zap.String("activity", peer.Activity.Time.Format(time.RFC3339)),
				zap.Bool("is_active_last_hour", lastActiveDeltaHours < 1),
				zap.Bool("is_active_last_day", lastActiveDeltaHours < 24),
			)
			if lastActiveDeltaHours < 1 {
				results.NumPeersActiveLastHour++
			}
			if lastActiveDeltaHours < 24 {
				results.NumPeersActiveLastDay++
			}
		}

		// Peer is expired - add it to the output list for later processing
		if peer.Expires != nil && peer.Expires.Time.Before(now) {
			results.ExpiredPeers = append(results.ExpiredPeers, peer)
		}
	}

	// Finally snap the current number of available peers
	results.NumPeers = len(s.stats)

	return results
}

func (s *runtimePeerStatsService) updateRuntimePeerStatFromWireguardPeer(now time.Time, wgPeer wgtypes.Peer, peer *types.PeerInfo) peerChangeSummary {
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

	stat, ok := s.stats[*peer.WireguardPublicKey]
	if !ok {
		stat = newRuntimePeerStat()
		s.stats[*peer.WireguardPublicKey] = stat
	}

	if wgPeer.ReceiveBytes > stat.Upstream {
		// Upstream never be nil
		*peer.Upstream += wgPeer.ReceiveBytes - stat.Upstream
		changeSum.Set(peerChangeTraffic)
	}

	if wgPeer.TransmitBytes > stat.Downstream {
		// Downstream never be nil
		*peer.Downstream += wgPeer.TransmitBytes - stat.Downstream
		changeSum.Set(peerChangeTraffic)
	}

	zap.L().Debug(
		"update",
		zap.Stringp("label", peer.Label),
		zap.Int64("wg_upstream", wgPeer.ReceiveBytes),
		zap.Int64("stats_upstream", stat.Upstream),
		zap.Int64("peer_upstream", *peer.Upstream),
		zap.Int64("change_upstream", wgPeer.ReceiveBytes-stat.Upstream),
		zap.Int64("wg_downstream", wgPeer.TransmitBytes),
		zap.Int64("stats_downstream", stat.Downstream),
		zap.Int64("peer_downstream", *peer.Downstream),
		zap.Int64("change_downstream", wgPeer.TransmitBytes-stat.Downstream),
	)

	stat.Update(now, wgPeer.ReceiveBytes, wgPeer.TransmitBytes)

	return changeSum
}
