package stats

import (
	"time"

	"github.com/vpnhouse/common-lib-go/human"
	"github.com/vpnhouse/common-lib-go/stats"
	"go.uber.org/zap"

	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/settings"
	"github.com/vpnhouse/tunnel/proto"
)

type OnFlush func(protocol string, report *stats.Report)

func New(flushInterval time.Duration, eventlogService eventlog.EventManager, protocol string, onFlush OnFlush) (*stats.Service, error) {
	if flushInterval == 0 {
		flushInterval = human.MustParseInterval(settings.DefaultFlushStatisticsInterval).Value()
		zap.L().Info("stats flush interval is not defined use default one")
	}

	zap.L().Info("stats flush interval",
		zap.String("protocol", protocol), zap.Duration("flush_interval", flushInterval))

	return stats.New(flushInterval, func(report *stats.Report) {
		if onFlush != nil {
			onFlush(protocol, report)
		}

		err := eventlogService.Push(eventlog.PeerTraffic, &proto.PeerInfo{
			SessionID:      report.SessionID,
			UserID:         report.UserID,
			InstallationID: report.InstallationID,
			Country:        report.Country,
			Protocol:       protocol,
			BytesDeltaTx:   report.DeltaTx,
			BytesRx:        report.DeltaRx,
			Seconds:        report.DeltaT,
			Created:        &proto.Timestamp{Sec: int64(report.Created)},
			Updated:        &proto.Timestamp{Sec: int64(report.Created + report.DeltaT)},
			ActivityID:     report.SessionID,
		})
		if err != nil {
			// Avoid error stack trace details - simply pur error description
			zap.L().Error("error sending event traffic", zap.String("error", err.Error()))
		}
	})
}
