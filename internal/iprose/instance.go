//go:build iprose
// +build iprose

package iprose

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/iprose-go/pkg/server"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

type Instance struct {
	iprose     *server.IPRoseServer
	runtime    *runtime.TunnelRuntime
	authorizer *authorizer.JWTAuthorizer
}

func New(runtime *runtime.TunnelRuntime, authorizer *authorizer.JWTAuthorizer) (*Instance, error) {
	instance := &Instance{
		authorizer: authorizer,
		runtime:    runtime,
	}
	var err error
	instance.iprose, err = server.New(
		"iprose0",
		"10.123.76.1/24",
		"",
		[]string{"0.0.0.0/0"},
		128,
		instance.Authenticate,
	)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (instance *Instance) Authenticate(r *http.Request) (*string, error) {
	// Extract JWT
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		if instance.runtime.Settings.IPRoseNoAuth {
			return nil, nil
		}
		return nil, xerror.EAuthenticationFailed("no auth token", nil)
	}

	// Verify JWT, get JWT claims
	claims, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return nil, err
	}

	return &claims.Subject, nil
}

func (instance *Instance) RegisterHandlers(r chi.Router) {
	for _, hndlr := range instance.iprose.Handlers() {
		r.HandleFunc(hndlr.Pattern, hndlr.Func)
	}
}

func (instance *Instance) Shutdown() error {
	instance.iprose.Shutdown()
	return nil
}

func (instance *Instance) Running() bool {
	return instance.iprose.Running()
}
