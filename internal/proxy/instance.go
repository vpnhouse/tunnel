package proxy

import (
	"sync/atomic"

	"github.com/go-httpproxy/httpproxy"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type Config struct {
	ConnLimit int `yaml:"connlimit"`
}

type Instance struct {
	config     *Config
	authorizer authorizer.JWTAuthorizer
	proxy      *httpproxy.Proxy
	users      *userStorage
	terminated atomic.Bool
}

func New(config *Config, jwtAuthorizer authorizer.JWTAuthorizer) (*Instance, error) {
	if config == nil {
		zap.L().Warn("Not starting proxy - no configuration")
		return nil, nil
	}

	proxy, err := httpproxy.NewProxy()
	if err != nil {
		return nil, xerror.EInternalError("Can't create proxy", err)
	}

	return &Instance{
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		config:     config,
		proxy:      proxy,
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
