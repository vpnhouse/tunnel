package stats

import (
	"strings"
	"time"

	"github.com/vpnhouse/common-lib-go/xstats"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
)

func (s *Service) pushLog(protocol string, report *xstats.Report) {
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
		// Avoid error stack trace details - simply put error description
		zap.L().Error("error sending event traffic", zap.String("error", err.Error()))
	}

}

func (s *Service) onFlush(proto string, report *xstats.Report) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	record, ok := s.records[proto]
	if !ok || record.service == nil {
		zap.L().Error("Ignoring onFlush event: protocol is not registered", zap.String("proto", proto))
		return
	}

	record.pending.upstream.Add(report.DeltaTx)
	record.pending.downstream.Add(report.DeltaRx)

	s.pushLog(proto, report)
}

func (s *Service) worker() {
	for {
		select {
		case <-s.shutdown:
			return
		case <-time.NewTicker(s.settings.exportInterval()).C:
			s.lock.RLock()
			s.export()
			s.save()
			s.lock.RUnlock()
		}
	}
}

func (s *Service) export() {
	now := time.Now()
	for _, record := range s.records {
		deltaUp := record.pending.upstream.Swap(0)
		deltaDown := record.pending.downstream.Swap(0)

		record.total.upstream += deltaUp
		record.total.downstream += deltaDown

		record.export = Stats{
			ExtraStats:      record.extraCb(),
			UpstreamBytes:   int64(record.total.upstream),
			DownstreamBytes: int64(record.total.downstream),
			UpstreamSpeed:   int64(speed(deltaUp, now, record.total.at)),
			DownstreamSpeed: int64(speed(deltaDown, now, record.total.at)),
		}

		record.total.at = now
	}
}

func (s *Service) loadMetrics(proto string) {
	metrics, err := s.storage.GetMetrics([]string{"upstream_" + proto, "downstream_" + proto})
	if err != nil {
		zap.L().Error("Can't load statistics", zap.Error(err))
		return
	}

	now := time.Now()
	for name, value := range metrics {
		direction, proto, found := strings.Cut(name, "_")
		if !found {
			zap.L().Error("Invalid metrics record", zap.String("name", name), zap.Int64("value", value))
			continue
		}

		s.records[proto].total.at = now
		if direction == "upstream" {
			s.records[proto].total.upstream = uint64(value)
		} else if direction == "downstream" {
			s.records[proto].total.downstream = uint64(value)
		} else {
			zap.L().Error("Ignoring unknown metrics record", zap.String("name", name), zap.Int64("value", value))
			continue
		}
	}
}

func (s *Service) save() {
	metrics := map[string]int64{}
	for k, v := range s.records {
		metrics["upstream_"+k] = int64(v.total.upstream)
		metrics["downstream_"+k] = int64(v.total.downstream)
	}
	s.storage.SetMetrics(metrics)
}

func speed(delta uint64, now, since time.Time) uint64 {
	deltaTMilli := now.Sub(since).Milliseconds()
	if deltaTMilli == 0 {
		return 0
	}
	return (delta * 1000) / uint64(deltaTMilli)
}
