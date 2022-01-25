package xhttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	middlewarestd "github.com/slok/go-http-metrics/middleware/std"
	"go.uber.org/zap"
)

type Handlers map[string]http.Handler

type Service struct {
	server      *http.Server
	muxer       *http.ServeMux
	httpMetrics middleware.Middleware
	running     bool
}

func New(listenAddr string, eventManager *control.EventManager, handlers ...Handlers) (*Service, error) {
	muxer := http.NewServeMux()

	// Create instance
	instance := &Service{
		server: &http.Server{
			Handler: muxer,
		},
		muxer: muxer,
		httpMetrics: middleware.New(middleware.Config{
			Recorder:      metrics.NewRecorder(metrics.Config{}),
			GroupedStatus: true,
		}),
	}

	// Listen to port
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, xerror.EInternalError("can't listen port", err, zap.String("address", listenAddr))
	}

	// Register metrics handler
	muxer.Handle("/metrics", promhttp.Handler())

	// Register initial handlers
	for _, handlerList := range handlers {
		for path, handler := range handlerList {
			zap.L().Debug("registering handler", zap.String("path", path))
			instance.Handle(path, handler)
		}
	}

	// Start HTTP API
	go func() {
		instance.running = true
		if err := instance.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			eventManager.EmitEventWithInfo(control.EventCriticalError,
				xerror.EInternalError("can't server HTTP", err),
			)
		}
	}()

	return instance, nil
}

func (m *Service) Handle(pattern string, handler http.Handler) {
	h := middlewarestd.Handler("", m.httpMetrics, handler)
	m.muxer.Handle(pattern, h)
}

func (m *Service) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := m.server.Shutdown(ctx); err != nil {
		return xerror.EInternalError("HTTP server shutdown failed", err)
	}

	m.running = false
	zap.L().Debug("HTTP Server Exited Properly")
	return nil
}

func (m *Service) Running() bool {
	return m.running
}
