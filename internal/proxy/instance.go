package proxy

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type Config struct {
	ConnLimit   int           `yaml:"conn_limit"`
	ConnTimeout time.Duration `yaml:"conn_timeout"`
}

type Instance struct {
	config     *Config
	authorizer authorizer.JWTAuthorizer
	users      *userStorage
	terminated atomic.Bool
}

func New(config *Config, jwtAuthorizer authorizer.JWTAuthorizer) (*Instance, error) {
	if config == nil {
		zap.L().Warn("Not starting proxy - no configuration")
		return nil, nil
	}

	return &Instance{
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		config:     config,
		users:      newUserStorage(config.ConnLimit),
	}, nil
}

func (instance *Instance) Shutdown() error {
	if instance.terminated.Swap(true) {
		return xerror.EInternalError("Double proxy shutdown", nil)
	}

	return nil
}

func (instance *Instance) Running() bool {
	return instance.terminated.Load()
}

func (instance *Instance) ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Debug("Query", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		if r.Method == http.MethodConnect || r.URL.IsAbs() {
			instance.doProxy(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
