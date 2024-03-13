package proxy

import (
	"context"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

type Config struct {
	ConnLimit int `yaml:"connlimit"`
}

type Instance struct {
	ctx        context.Context
	cancel     context.CancelFunc
	terminated atomic.Bool
	config     *Config
	authorizer authorizer.JWTAuthorizer
	users      *userStorage
}

func New(config *Config, jwtAuthorizer authorizer.JWTAuthorizer) *Instance {
	if config == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Instance{
		ctx:        ctx,
		cancel:     cancel,
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		config:     config,
		users:      newUserStorage(ctx, config.ConnLimit),
	}
}

func (instance *Instance) RegisterHandlers(r chi.Router) {
	r.MethodFunc("CONNECT", "/*", instance.handler)
}

func (instance *Instance) Shutdown() error {
	if instance.terminated.Swap(true) {
		return xerror.EInternalError("Double proxy shutdown", nil)
	}

	instance.cancel()
	return nil
}

func (instance *Instance) Running() bool {
	return !instance.terminated.Load()
}
