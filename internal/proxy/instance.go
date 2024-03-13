package proxy

import (
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

type Instance struct {
	terminated atomic.Bool
	runtime    *runtime.TunnelRuntime
	authorizer authorizer.JWTAuthorizer
}

func New(runtime *runtime.TunnelRuntime, jwtAuthorizer authorizer.JWTAuthorizer) *Instance {
	return &Instance{
		authorizer: jwtAuthorizer, //authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		runtime:    runtime,
	}
}

func (instance *Instance) RegisterHandlers(r chi.Router) {
	r.MethodFunc("CONNECT", "/*", instance.handler)
}

func (instance *Instance) Shutdown() error {
	if instance.terminated.Swap(true) {
		return xerror.EInternalError("Double proxy shutdown", nil)
	}

	return nil
}

func (instance *Instance) Running() bool {
	return !instance.terminated.Load()
}
