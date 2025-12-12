package stats

import "github.com/prometheus/client_golang/prometheus"

const Namespace = "tunnel"

type prometheusStats struct {
	peers           prometheus.Gauge
	active          prometheus.Gauge
	upstreamBytes   prometheus.Counter
	downstreamBytes prometheus.Counter
}

func newPrometheusStats(proto string) *prometheusStats {
	stats := prometheusStats{
		peers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   Namespace,
			Name:        "peers_total",
			Help:        "number of peers",
			ConstLabels: prometheus.Labels{"proto": proto},
		}),

		active: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   Namespace,
			Name:        "peers_active",
			Help:        "number of active peers",
			ConstLabels: prometheus.Labels{"proto": proto},
		}),

		upstreamBytes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   Namespace,
			Name:        "upstream_bytes",
			Help:        "upstream bytes count",
			ConstLabels: prometheus.Labels{"proto": proto},
		}),

		downstreamBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   Namespace,
			Name:        "downstream_bytes",
			Help:        "downstream bytes count",
			ConstLabels: prometheus.Labels{"proto": proto},
		}),
	}

	prometheus.MustRegister(stats.peers)
	prometheus.MustRegister(stats.active)
	prometheus.MustRegister(stats.upstreamBytes)
	prometheus.MustRegister(stats.downstreamBytes)

	return &stats
}
