package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/storage"

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

type pendingStats struct {
	upstream, downstream atomic.Uint64
}

type totalStats struct {
	upstream, downstream uint64
	at                   time.Time
}

type protoRecord struct {
	export  Stats
	pending pendingStats
	total   totalStats
	extraCb ExtraStatsCb
	service *stats.Service
}

type StatsService struct {
	lock     sync.RWMutex
	settings *Settings
	storage  *storage.Storage
	eventlog eventlog.EventManager
	shutdown chan struct{}
	records  map[string]*protoRecord
}

func NewService(settings *Settings, eventLog eventlog.EventManager, storage *storage.Storage) *StatsService {
	s := &StatsService{
		storage:  storage,
		eventlog: eventLog,
		shutdown: make(chan struct{}),
		records:  make(map[string]*protoRecord),
	}

	s.load()
	go s.worker()

	return s
}

func (s *StatsService) Stats() map[string]Stats {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := map[string]Stats{}
	for k, v := range s.records {
		result[k] = v.export

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

	service, err := stats.New(s.settings.flushInterval(),
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

func (s *StatsService) Report(proto string, sessionID uuid.UUID, drx, dtx uint64, onSessionDataRequired stats.OnData) {
	s.lock.RLock()
	defer s.lock.RUnlock()

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

	close(s.shutdown)
	s.save()

	return nil
}
