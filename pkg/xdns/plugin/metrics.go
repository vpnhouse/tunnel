package plugin

import (
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var blockedCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "blocklist",
	Name:      "request_blocked_total",
	Help:      "Counter of requests blocked.",
}, []string{"server"})

var lookupDurationHist = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: plugin.Namespace,
	Subsystem: "blocklist",
	Name:      "lookup_duration",
	Help:      "Distribution of the lookup duration",
	// lets use nanoseconds, the native duration values
	Buckets: []float64{
		float64(500 * time.Microsecond),
		float64(1_000 * time.Microsecond),     // 1ms
		float64(5_000 * time.Microsecond),     // 5ms
		float64(10_000 * time.Microsecond),    // 10ms
		float64(50_000 * time.Microsecond),    // 50ms
		float64(100_000 * time.Microsecond),   // 100ms
		float64(250_000 * time.Microsecond),   // 1/4s
		float64(500_000 * time.Microsecond),   // 1/2s
		float64(1_000_000 * time.Microsecond), // 1s
	},
}, []string{"server"})
