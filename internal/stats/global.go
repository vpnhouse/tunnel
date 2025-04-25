package stats

import (
	"strings"
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/internal/storage"
	"go.uber.org/zap"
)

type Pair struct {
	Upstream   int64
	Downstream int64
}

type ProtoRecord struct {
	Pair
	Peers int
	At    time.Time
}

type ProtoRecordMap map[string]*ProtoRecord
type ProtoSpeedMap map[string]*Pair

type TrafficStats struct {
	lock sync.Mutex

	storage  *storage.Storage
	shutdown chan struct{}

	current, previous ProtoRecordMap
}

func NewTrafficStats(storage *storage.Storage) *TrafficStats {
	s := &TrafficStats{
		storage:  storage,
		shutdown: make(chan struct{}),
		current:  make(ProtoRecordMap),
	}

	s.load()
	go s.worker()

	return s
}

func (s *TrafficStats) Stats() (ProtoRecordMap, ProtoSpeedMap) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.traffic(), s.speed()
}

func (s *TrafficStats) GlobalStats() (ProtoRecord, Pair) {
	s.lock.Lock()
	defer s.lock.Unlock()

	current := s.current.sum()
	previous := s.previous.sum()

	return current, *current.speed(&previous)
}

func (s *TrafficStats) Report(proto string, diffUpstream, diffDownstream int64, peersCount int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	record := s.current[proto]
	record.Upstream += diffUpstream
	record.Downstream += diffDownstream
	record.Peers = peersCount
	record.At = time.Now()
}

func (s *TrafficStats) Shutdown() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.save()
	close(s.shutdown)
}

func (s *TrafficStats) traffic() ProtoRecordMap {
	return s.current.copy()
}

func (s *TrafficStats) speed() ProtoSpeedMap {
	speeds := ProtoSpeedMap{}
	for proto, record := range s.current {
		speeds[proto] = record.speed(s.previous[proto])
	}

	return speeds
}

func (s *TrafficStats) worker() {
	for {
		select {
		case <-s.shutdown:
			return
		case <-time.NewTicker(time.Minute).C:
			s.lock.Lock()
			s.save()
			s.previous = s.current.copy()
			s.lock.Unlock()
		}
	}
}

func (s *TrafficStats) load() {
	now := time.Now()
	metrics, err := s.storage.GetMetricsLike([]string{"upstream_%", "downstream_%"})
	if err != nil {
		zap.L().Error("Can't load statistics", zap.Error(err))
	}

	for name, value := range metrics {
		sp := strings.SplitN(name, "_", 2)
		if len(sp) != 2 {
			zap.L().Error("Invalid metrics record", zap.String("name", name), zap.Int64("value", value))
			continue
		}

		direction := sp[0]
		proto := sp[1]

		s.current[proto].At = now
		if direction == "upstream" {
			s.current[proto].Upstream = value
		} else if direction == "downstream" {
			s.current[proto].Downstream = value
		} else {
			zap.L().Error("Ignoring unknown metrics record", zap.String("name", name), zap.Int64("value", value))
			continue
		}
	}
}

func (s *TrafficStats) save() {
	metrics := map[string]int64{}
	for k, v := range s.current {
		metrics["upstream_"+k] = v.Upstream
		metrics["downstream_"+k] = v.Downstream
	}
	s.storage.SetMetrics(metrics)
}

func (m ProtoRecordMap) copy() ProtoRecordMap {
	copied := ProtoRecordMap{}

	for k, v := range m {
		vCopied := *v
		copied[k] = &vCopied
	}

	return copied
}

func (m ProtoRecordMap) sum() ProtoRecord {
	sum := ProtoRecord{}

	for _, v := range m {
		if sum.At.Before(v.At) {
			sum.At = v.At
		}

		sum.Upstream += v.Upstream
		sum.Downstream += v.Downstream
		sum.Peers += v.Peers
	}

	return sum
}

func (p *ProtoRecord) speed(since *ProtoRecord) *Pair {
	if since == nil || p.At == since.At {
		return &Pair{}
	}

	interval := p.At.Sub(since.At)
	return &Pair{
		Upstream:   (p.Upstream - since.Upstream) / int64(interval.Seconds()),
		Downstream: (p.Downstream - since.Downstream) / int64(interval.Seconds()),
	}
}
