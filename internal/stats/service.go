package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xstats"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/storage"

	"go.uber.org/zap"
)

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
	service *xstats.Service
}

type Service struct {
	lock     sync.RWMutex
	settings *Settings
	storage  *storage.Storage
	eventlog eventlog.EventManager
	shutdown chan struct{}
	records  map[string]*protoRecord
}

func NewService(settings *Settings, eventLog eventlog.EventManager, storage *storage.Storage) *Service {
	s := &Service{
		storage:  storage,
		eventlog: eventLog,
		shutdown: make(chan struct{}),
		records:  make(map[string]*protoRecord),
	}

	go s.worker()

	return s
}

func (s *Service) Stats() (Stats, map[string]Stats) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	global := Stats{}

	result := map[string]Stats{}
	for k, v := range s.records {
		result[k] = v.export
		global.Add(&v.export)

	}

	return global, result
}

func (s *Service) Register(proto string, extraCb ExtraStatsCb) (*xstats.Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	record, ok := s.records[proto]
	if ok && record.service != nil {
		return nil, xerror.EInternalError("Protocol already registered", nil, zap.String("proto", proto))
	}

	service, err := xstats.New(s.settings.flushInterval(),
		func(report *xstats.Report) {
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
	s.loadMetrics(proto)
	s.export()
	return record.service, nil
}

func (s *Service) Running() bool {
	select {
	case <-s.shutdown:
		return false
	default:
		return true
	}
}

func (s *Service) Shutdown() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	close(s.shutdown)
	s.save()

	return nil
}
