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

func New(flushInterval time.Duration, eventlogService eventlog.EventManager, protocol string) *stats.Service {
	if flushInterval == 0 {
		flushInterval = human.MustParseInterval(settings.DefaultFlushStatisticsInterval).Value()
		zap.L().Info("stats flush interval is not defined use default one")
		return nil
	}

	stats.New(flushInterval, func(report *stats.Report) {
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
			ActivityID:     report.SessionID,
		})
		if err != nil {
			// Avoid error stack trace details - simply pur error description
			zap.L().Error("error sending event traffic", zap.String("error", err.Error()))
		}
	})
}
