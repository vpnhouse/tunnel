package stats

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/proto"

	"go.uber.org/zap"
)

type ExtraStats struct {
	Peers int
}

type ExtraStatsCb func() ExtraStats

type Stats struct {
	ExtraStats
	UpstreamBytes   uint64
	DownstreamBytes uint64
	UpstreamSpeed   uint64
	DownstreamSpeed uint64
}

type protoRecord struct {
	stats   Stats
	extraCb ExtraStatsCb
	service *stats.Service
}

type protoRecordMap map[string]*protoRecord

type StatsService struct {
	lock          sync.Mutex
	flushInterval time.Duration
	storage       *storage.Storage
	eventlog      eventlog.EventManager
	shutdown      chan struct{}
	records       protoRecordMap
}

func NewService(flushInterval time.Duration, eventLog eventlog.EventManager, storage *storage.Storage) *StatsService {
	s := &StatsService{
		flushInterval: flushInterval,
		storage:       storage,
		eventlog:      eventLog,
		shutdown:      make(chan struct{}),
		records:       make(protoRecordMap),
	}

	s.load()
	go s.worker()

	return s
}

func (s *StatsService) Stats() map[string]Stats {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := map[string]Stats{}
	for k, v := range s.records {
		result[k] = v.stats
	}

	return result
}

func (s *StatsService) Register(proto string, extraCb ExtraStatsCb) (*stats.Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	record, ok := s.records[proto]
	if ok && record.service != nil {
		return nil, xerror.EInternalError("Protocol already registered", nil, zap.String("proto", proto))
	}

	service, err := stats.New(s.flushInterval,
		func(report *stats.Report) {
			s.onFlush(proto, report)
		},
	)
	if err != nil {
		delete(s.records, proto)
		return nil, xerror.EInternalError("Failed to create statistics service protocol", err, zap.String("proto", proto))
	}

	record = &protoRecord{
		service: service,
		extraCb: extraCb,
	}
	s.records[proto] = record
	return record.service, nil
}

func (s *StatsService) pushLog(protocol string, report *stats.Report) {
	err := s.eventlog.Push(eventlog.PeerTraffic, &proto.PeerInfo{
		SessionID:      report.SessionID,
		UserID:         report.UserID,
		InstallationID: report.InstallationID,
		Country:        report.Country,
		Protocol:       protocol,
		BytesDeltaTx:   report.DeltaTx,
		BytesRx:        report.DeltaRx,
		Seconds:        report.DeltaTNano / 1e9,
		Created:        &proto.Timestamp{Sec: int64(report.CreatedNano / 1e9)},
		Updated:        &proto.Timestamp{Sec: int64(report.CreatedNano+report.DeltaTNano) / 1e9},
		ActivityID:     report.SessionID,
	})
	if err != nil {
		// Avoid error stack trace details - simply pur error description
		zap.L().Error("error sending event traffic", zap.String("error", err.Error()))
	}

}

func (s *StatsService) onFlush(proto string, report *stats.Report) {
	// We must go to background to avoid double lock when onFlush is combined with Report call
	go func() {
		s.lock.Lock()
		defer s.lock.Unlock()

		record, ok := s.records[proto]
		if !ok || record.service == nil {
			zap.L().Error("Ignoring onFlush event: protocol is not registered", zap.String("proto", proto))
			return
		}

		record.stats.UpstreamBytes += report.DeltaRx
		record.stats.DownstreamBytes += report.DeltaTx
		record.stats.ExtraStats = record.extraCb()
		record.stats.UpstreamSpeed = (report.DeltaRx * 1e3) / (report.DeltaTNano / 1e6)
		record.stats.DownstreamSpeed = (report.DeltaTx * 1e3) / (report.DeltaTNano / 1e6)
	}()
}

func (s *StatsService) Report(proto string, sessionID uuid.UUID, drx, dtx uint64, onSessionDataRequired stats.OnData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	record, ok := s.records[proto]
	if !ok {
		zap.L().Error("Failed to lookup record for protocol", zap.String("proto", proto))
		return
	}

	if record.service == nil {
		zap.L().Error("Reporting on unregistered protocol is ignored", zap.String("proto", proto))
		return
	}

	record.service.ReportStats(sessionID, drx, dtx, onSessionDataRequired)
}

func (s *StatsService) Running() bool {
	select {
	case <-s.shutdown:
		return false
	default:
		return true
	}
}

func (s *StatsService) Shutdown() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.save()
	close(s.shutdown)
	return nil
}

func (s *StatsService) worker() {
	for {
		select {
		case <-s.shutdown:
			return
		case <-time.NewTicker(time.Minute).C:
			s.lock.Lock()
			s.save()
			s.lock.Unlock()
		}
	}
}

func (s *StatsService) load() {
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

		if direction == "upstream" {
			s.records[proto].stats.DownstreamBytes = uint64(value)
		} else if direction == "downstream" {
			s.records[proto].stats.DownstreamBytes = uint64(value)
		} else {
			zap.L().Error("Ignoring unknown metrics record", zap.String("name", name), zap.Int64("value", value))
			continue
		}
	}
}

func (s *StatsService) save() {
	metrics := map[string]int64{}
	for k, v := range s.records {
		metrics["upstream_"+k] = int64(v.stats.UpstreamBytes)
		metrics["downstream_"+k] = int64(v.stats.UpstreamBytes)
	}
	s.storage.SetMetrics(metrics)
}
