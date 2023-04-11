package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/human"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
)

type trafficState struct {
	UpstreamBytesChange   int64
	DownstreamBytesChange int64
}

func (s *trafficState) Reset() {
	s.UpstreamBytesChange = 0
	s.DownstreamBytesChange = 0
}

type PeerTraffic struct {
	Downstream int64
	Upstream   int64
}

type peerTrafficUpdateEventSender struct {
	eventLog           eventlog.EventManager
	maxUpstreamBytes   int64
	maxDownstreamBytes int64
	sendInterval       time.Duration
	stop               chan struct{}
	done               chan struct{}
	statsService       *runtimePeerStatsService

	needSendChan chan struct{}
	lock         sync.Mutex
	state        trafficState
	// All peers (prev)
	peerTraffic map[string]*PeerTraffic
	// peers candidates for sending
	updatedPeers map[string]*types.PeerInfo
}

func NewPeerTrafficUpdateEventSender(runtime *runtime.TunnelRuntime, eventLog eventlog.EventManager, statsService *runtimePeerStatsService, peers []*types.PeerInfo) *peerTrafficUpdateEventSender {
	maxUpstreamBytes := int64(0)
	maxDownstreamBytes := int64(0)
	sendInterval := runtime.Settings.GetSentEventInterval().Value()
	if runtime.Settings != nil && runtime.Settings.PeerStatistics != nil {
		maxUpstreamBytes = runtime.Settings.PeerStatistics.MaxUpstreamTrafficChange.Value()
		maxDownstreamBytes = runtime.Settings.PeerStatistics.MaxDownstreamTrafficChange.Value()
	}

	peerTraffic := make(map[string]*PeerTraffic, len(peers))
	for _, peer := range peers {
		if peer.WireguardPublicKey == nil {
			continue
		}
		var downstream int64
		if peer.Downstream != nil {
			downstream = *peer.Downstream
		}
		var upstream int64
		if peer.Upstream != nil {
			upstream = *peer.Upstream
		}
		peerTraffic[*peer.WireguardPublicKey] = &PeerTraffic{
			Downstream: downstream,
			Upstream:   upstream,
		}
	}

	sender := &peerTrafficUpdateEventSender{
		maxUpstreamBytes:   maxUpstreamBytes,
		maxDownstreamBytes: maxDownstreamBytes,
		sendInterval:       sendInterval,
		eventLog:           eventLog,
		peerTraffic:        peerTraffic,
		updatedPeers:       make(map[string]*types.PeerInfo, len(peers)),
		statsService:       statsService,
		needSendChan:       make(chan struct{}, 1),
	}

	go sender.run()

	return sender
}

func (s *peerTrafficUpdateEventSender) Add(peer *types.PeerInfo) {
	if peer.WireguardPublicKey == nil {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	var downstream int64
	if peer.Downstream != nil {
		downstream = *peer.Downstream
	}
	var upstream int64
	if peer.Upstream != nil {
		upstream = *peer.Upstream
	}
	s.peerTraffic[*peer.WireguardPublicKey] = &PeerTraffic{
		Downstream: downstream,
		Upstream:   upstream,
	}
}

func (s *peerTrafficUpdateEventSender) Remove(peer *types.PeerInfo) {
	if peer.WireguardPublicKey == nil {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.peerTraffic[*peer.WireguardPublicKey]; ok {
		delete(s.peerTraffic, *peer.WireguardPublicKey)
	}
}

func (s *peerTrafficUpdateEventSender) Send(peers []*types.PeerInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, peer := range peers {
		oldPeerTraffic, ok := s.peerTraffic[*peer.WireguardPublicKey]
		if !ok {
			// We should never be here but for the sake of simplitity
			// assume the peer gone and simply do nothing
			continue
		}
		if peer.Upstream != nil {
			s.state.UpstreamBytesChange += *peer.Upstream - oldPeerTraffic.Upstream
			oldPeerTraffic.Upstream = *peer.Upstream
		}
		if peer.Downstream != nil {
			s.state.DownstreamBytesChange += *peer.Downstream - oldPeerTraffic.Downstream
			oldPeerTraffic.Downstream = *peer.Downstream
		}
		s.updatedPeers[*peer.WireguardPublicKey] = peer
	}

	// Check upstream or downstream exceeds the limits
	needSend := false
	if s.maxUpstreamBytes > 0 && s.state.UpstreamBytesChange > s.maxUpstreamBytes {
		needSend = true
	} else if s.maxDownstreamBytes > 0 && s.state.DownstreamBytesChange > s.maxDownstreamBytes {
		needSend = true
	}

	if needSend {
		select {
		case s.needSendChan <- struct{}{}:
		default:
		}
	}
}

func (s *peerTrafficUpdateEventSender) sendUpdates() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.updatedPeers) == 0 {
		return
	}
	for _, peer := range s.updatedPeers {
		for _, sess := range s.statsService.GetSessions(peer) {
			err := s.eventLog.Push(eventlog.PeerTraffic, intoProto(peer, &sess))
			if err != nil {
				zap.L().Error("failed to push event", zap.Error(err), zap.Uint32("type", uint32(proto.EventType_PeerTraffic)))
			}
		}
	}
	zap.L().Info(
		"send peer traffic updates",
		zap.Int("peers", len(s.updatedPeers)),
		zap.String("upstream", human.FormatSizeToHuman(uint64(s.state.UpstreamBytesChange))),
		zap.String("downstream", human.FormatSizeToHuman(uint64(s.state.DownstreamBytesChange))),
	)
	s.updatedPeers = make(map[string]*types.PeerInfo, len(s.updatedPeers))
	s.state.Reset()
}

func intoProto(peer *types.PeerInfo, sess *Session) *proto.PeerInfo {
	p := peer.IntoProto()
	p.BytesRx = uint64(sess.Upstream)
	p.BytesDeltaRx = uint64(sess.UpstreamDelta)
	p.BytesTx = uint64(sess.Downstream)
	p.BytesDeltaTx = uint64(sess.DownstreamDelta)
	p.Seconds = uint64(sess.Seconds)
	p.ActivityID = sess.ActivityID.String()
	p.Country = sess.Country
	peerStr := fmt.Sprintf("id=%s, sec=%d, rx=%d(+%d)[%d], tx=%d(+%d)[%d]",
		p.ActivityID, p.Seconds, p.BytesRx, p.BytesDeltaRx, *peer.Upstream, p.BytesTx, p.BytesDeltaTx, *peer.Downstream)
	zap.L().Debug("traffic change", zap.String("peer", peerStr))
	return p
}

func (s *peerTrafficUpdateEventSender) run() {
	sendPeerTicker := time.NewTicker(s.sendInterval)
	zap.L().Debug("Start sending peer traffic updates", zap.String("interval", fmt.Sprint(s.sendInterval)))

	defer func() {
		sendPeerTicker.Stop()
		close(s.done)
		zap.L().Debug("Stop sending peer traffic updates")
	}()

	for {
		select {
		case <-s.stop:
			zap.L().Info("Shutting down sending peer traffic updates")
			return
		case <-sendPeerTicker.C:
		case <-s.needSendChan:
		}
		s.sendUpdates()
	}
}

func (s *peerTrafficUpdateEventSender) Stop() {
	close(s.stop)
	<-s.done
}
