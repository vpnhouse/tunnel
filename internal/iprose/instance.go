//go:build iprose
// +build iprose

package iprose

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/iprose-go/pkg/server"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

type Config struct {
	QueueSize int `yaml:"queue_size"`
}

type Instance struct {
	iprose     *server.IPRoseServer
	authorizer authorizer.JWTAuthorizer
}

func New(config *Config, jwtAuthorizer authorizer.JWTAuthorizer) (*Instance, error) {
	if config == nil {
		zap.L().Warn("Not starting iprose - no configuration")
		return nil, nil
	}

	instance := &Instance{
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.IPRose),
	}
	var err error
	instance.iprose, err = server.New(
		"iprose0",
		"10.123.76.1/24",
		"",
		[]string{"0.0.0.0/0"},
		config.QueueSize,
		instance.Authenticate,
	)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (instance *Instance) authenticate(r *http.Request) error {
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return xerror.EAuthenticationFailed("no auth token", nil)
	}

	_, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return err
	}

	return nil
}

func (instance *Instance) Authenticate(r *http.Request) (error, int, []byte) {
	err := instance.authenticate(r)
	if err == nil {
		return nil, 0, nil
	} else {
		code, body := xerror.ErrorToHttpResponse(err)
		return err, code, body
	}
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
