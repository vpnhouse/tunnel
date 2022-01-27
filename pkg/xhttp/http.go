package xhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	openapi "github.com/Codename-Uranium/api/go/server/common"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	middlewarestd "github.com/slok/go-http-metrics/middleware/std"
	"go.uber.org/zap"
)

type Middleware = func(http.Handler) http.Handler

type Option func(w *wrapper)

func WithMiddleware(mw Middleware) Option {
	return func(w *wrapper) {
		w.router.Middlewares()
		w.router.Use(mw)
	}
}

func WithMetrics() Option {
	measureMW := middleware.New(middleware.Config{
		Recorder:      metrics.NewRecorder(metrics.Config{}),
		GroupedStatus: true,
	})
	return func(w *wrapper) {
		// the measurement middleware
		w.router.Use(func(handler http.Handler) http.Handler {
			return middlewarestd.Handler("", measureMW, handler)
		})
		// route to obtain metrics
		w.router.Handle("/metrics", promhttp.Handler())
	}
}

func WithLogger() Option {
	return func(w *wrapper) {
		w.router.Use(requestLogger)
	}
}

type wrapper struct {
	srv    *http.Server
	router chi.Router
}

// Run starts the http server and blocks
func (w *wrapper) Run(addr string) error {
	w.srv = &http.Server{
		Handler: w.router,
		Addr:    addr,
	}

	zap.L().Info("starting HTTP server", zap.String("addr", addr))
	return w.srv.ListenAndServe()
}

// Router exposes chi.Router for the external registration of handlers.
// usage:
// 		h.Router().Get("/apt/path", myHandler)
// 		h.Router().Post("/apt/verb", myOtherHandler)
func (w *wrapper) Router() chi.Router {
	return w.router
}

func New(opts ...Option) *wrapper {
	r := chi.NewRouter()
	// always respond with JSON by using the custom error handlers
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		txt := "Not found"
		err := openapi.Error{
			Result: "404",
			Error:  &txt,
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(err)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		txt := "Method not allowed"
		err := openapi.Error{
			Result: "405",
			Error:  &txt,
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(err)
	})

	h := &wrapper{router: r}
	for _, o := range opts {
		o(h)
	}

	return h
}

func NewDefault() *wrapper {
	return New(
		WithLogger(),
		WithMetrics(),
	)
}

func (w *wrapper) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := w.srv.Shutdown(ctx)
	w.srv = nil

	return err
}

func (w *wrapper) Running() bool {
	return w.srv != nil
}
