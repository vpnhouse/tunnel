package manager

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/geoip"
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

type runtimePeerSession struct {
	ActivityID      uuid.UUID // id describing the session
	Seconds         int64     // session seconds
	UpstreamDelta   int64     // delta in bytes between previous and current Upstream
	Upstream        int64     // current Upstream value in bytes
	DownstreamDelta int64     // delta in bytes between previous and current Downstream
	Downstream      int64     // current Downstream value in bytes
	Country         string    // user country
}

type Session struct {
	ActivityID      uuid.UUID // id describing the session
	Seconds         int64     // session seconds
	UpstreamDelta   int64     // delta in bytes between previous and current Upstream
	Upstream        int64     // current Upstream value in bytes
	DownstreamDelta int64     // delta in bytes between previous and current Downstream
	Downstream      int64     // current Downstream value in bytes
	Country         string    // user country
}

// Keeps accumulated peer counters updated for certian wireguard peer
type runtimePeerStat struct {
	Updated         int64  // timestamp in seconds (when session is updated)
	Upstream        int64  // bytes
	UpstreamSpeed   int64  // bytes per second
	Downstream      int64  // bytes
	DownstreamSpeed int64  // bytes per second
	Country         string // user country

	upstreamSpeedAvg   *statutils.AvgValue
	downstreamSpeedAvg *statutils.AvgValue
	startUpstream      int64 // bytes
	startDownstream    int64 // bytes

	lock     sync.Mutex
	sessions []*runtimePeerSession
}

func newRuntimePeerStat(updated int64, startUpstream int64, startDownstream int64, country string) *runtimePeerStat {
	zap.L().Debug("PEER_STAT",
		zap.Int64("updated", updated),
		zap.Int64("start_upstream", startUpstream),
		zap.Int64("start_downstream", startDownstream),
		zap.String("country", country),
	)
	return &runtimePeerStat{
		Updated:            updated,
		startUpstream:      startUpstream,
		startDownstream:    startDownstream,
		Country:            country,
		upstreamSpeedAvg:   statutils.NewAvgValue(10),
		downstreamSpeedAvg: statutils.NewAvgValue(10),
	}
}

func (s *runtimePeerStat) currentSession() *runtimePeerSession {
	if len(s.sessions) == 0 {
		return s.newSession()
	}
	return s.sessions[len(s.sessions)-1]
}

func (s *runtimePeerStat) newSession() *runtimePeerSession {
	if len(s.sessions) > 0 {
		sess := s.sessions[len(s.sessions)-1]
		if sess.DownstreamDelta == 0 && sess.UpstreamDelta == 0 {
			sess.ActivityID = uuid.New()
			sess.Seconds = 0
			return sess
		}
	}
	sess := &runtimePeerSession{
		ActivityID: uuid.New(),
		Upstream:   s.Upstream,
		Downstream: s.Downstream,
		Country:    s.Country,
	}
	s.sessions = append(s.sessions, sess)
	return sess
}

func (s *runtimePeerStat) UpdateSession(upstream int64, downstream int64, seconds int64, country string, resetInterval time.Duration) {
	if s == nil {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	var sess *runtimePeerSession
	if seconds > int64(resetInterval.Seconds()*2) {
		sess = s.newSession()
	} else {
		sess = s.currentSession()
	}
	zap.L().Debug("SESSION",
		zap.Int64("upstream", upstream),
		zap.Int64("session.upstream", sess.Upstream),
		zap.Int64("delta.upstream", upstream-sess.Upstream),
		zap.Int64("downstream", downstream),
		zap.Int64("delta.downstream", downstream-sess.Downstream),
		zap.Int64("seconds", seconds),
	)
	sess.Seconds += seconds
	sess.UpstreamDelta += upstream - sess.Upstream
	sess.Upstream = upstream
	sess.DownstreamDelta += downstream - sess.Downstream
	sess.Downstream = downstream
	sess.Country = country
}

func (s *runtimePeerStat) GetSessions() []Session {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.sessions) == 0 {
		return nil
	}
	sessions := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sessions = append(sessions, Session{
			ActivityID:      sess.ActivityID,
			Seconds:         sess.Seconds,
			UpstreamDelta:   sess.UpstreamDelta,
			Upstream:        s.startUpstream + sess.Upstream,
			DownstreamDelta: sess.DownstreamDelta,
			Downstream:      s.startDownstream + sess.Downstream,
			Country:         sess.Country,
		})
	}
	s.sessions = s.sessions[len(s.sessions)-1:]
	s.sessions[0].Seconds = 0
	s.sessions[0].UpstreamDelta = 0
	s.sessions[0].DownstreamDelta = 0
	return sessions
}

func (s *runtimePeerStat) Update(now time.Time, upstream int64, downstream int64, country string, resetInterval time.Duration) {
	if s == nil {
		return
	}
	ts := now.Unix()
	if upstream < s.Upstream {
		upstream = s.Upstream
	}
	if downstream < s.Downstream {
		downstream = s.Downstream
	}
	defer func() {
		s.Upstream = upstream
		s.Downstream = downstream
		s.Updated = ts
		s.Country = country
	}()

	var seconds int64
	if ts <= s.Updated || s.Updated == 0 {
		seconds = 0
	} else {
		seconds = ts - s.Updated
	}

	s.UpdateSession(upstream, downstream, seconds, country, resetInterval)

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

func (s *runtimePeerStat) UpdateSpeedNoTraffic() {
	if s == nil {
		return
	}
	s.UpstreamSpeed = s.upstreamSpeedAvg.Push(0)
	s.DownstreamSpeed = s.downstreamSpeedAvg.Push(0)
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
	ResetInterval time.Duration
	Geo           *geoip.Instance

	lock sync.Mutex
	// {peer public key} -> peerStats
	stats map[string]*runtimePeerStat
	once  sync.Once
}

func (s *runtimePeerStatsService) init() {
	s.stats = make(map[string]*runtimePeerStat, 1000)
}

func (s *runtimePeerStatsService) GetRuntimePeerStat(peer *types.PeerInfo) *runtimePeerStat {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.stats[*peer.WireguardPublicKey]
}

func (s *runtimePeerStatsService) GetSessions(peer *types.PeerInfo) []Session {
	stats := s.GetRuntimePeerStat(peer)
	// Stats can gone on peer deletion that's detected on UpdatePeersStats
	// In this case we simply return empty sessions list for removed peer
	if stats == nil {
		return nil
	}
	return stats.GetSessions()
}

func (s *runtimePeerStatsService) UpdatePeersStats(now time.Time, peers []*types.PeerInfo, wireguardPeers map[string]wgtypes.Peer) updatePeerStatsResults {
	s.once.Do(s.init)

	s.lock.Lock()
	defer s.lock.Unlock()

	results := updatePeerStatsResults{
		UpdatedPeers:        make([]*types.PeerInfo, 0, len(peers)),
		ExpiredPeers:        make([]*types.PeerInfo, 0, len(peers)),
		FirstConnectedPeers: make([]*types.PeerInfo, 0, len(peers)),
	}

	existedPeers := make(map[string]struct{}, len(peers))

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
			delete(s.stats, *peer.WireguardPublicKey)
			continue
		}

		existedPeers[*peer.WireguardPublicKey] = struct{}{}

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

	for key := range s.stats {
		if _, ok := existedPeers[key]; ok {
			continue
		}
		// Delete peer that was gone or absent
		// Usually the peers may disappear when expire
		delete(s.stats, key)
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

	var country string
	if s.Geo != nil && wgPeer.Endpoint != nil {
		var err error
		country, err = s.Geo.GetCountry(wgPeer.Endpoint.IP)
		if err != nil {
			zap.L().Error("failed to detect country", zap.Stringp("peer", peer.Label))
		}
		country = strings.ToLower(country)
	}

	stat, ok := s.stats[*peer.WireguardPublicKey]
	if !ok {
		var updated int64
		if peer.Updated != nil {
			updated = peer.Updated.Time.Unix()
		}
		// Upstream and Upstream never be nil
		stat = newRuntimePeerStat(updated, *peer.Upstream, *peer.Downstream, country)
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

	if wgPeer.TransmitBytes-stat.Downstream > 0 || wgPeer.ReceiveBytes-stat.Upstream > 0 {
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
	}

	if changeSum.Has(peerChangeTraffic) {
		stat.Update(now, wgPeer.ReceiveBytes, wgPeer.TransmitBytes, country, s.ResetInterval)
	} else {
		stat.UpdateSpeedNoTraffic()
	}

	return changeSum
}
