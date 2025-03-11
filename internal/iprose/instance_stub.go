//go:build !iprose
// +build !iprose

package iprose

import (
	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/tunnel/internal/authorizer"
)

type (
	Config   struct{}
	Instance struct {
		stopped bool
	}
)

var DefaultConfig = Config{}

func New(config Config, jwtAuthorizer authorizer.JWTAuthorizer, statsService *stats.Service) (*Instance, error) {
	return &Instance{}, nil
}

func (instance *Instance) RegisterHandlers(r chi.Router) {}

func (instance *Instance) Shutdown() error {
	instance.stopped = true
	return nil
}

func (instance *Instance) Running() bool {
	return !instance.stopped
}
