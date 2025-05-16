//go:build !iprose
// +build !iprose

package iprose

import (
	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/stats"
)

type (
	Config   struct{}
	Instance struct {
		stopped bool
	}
)

var DefaultConfig = Config{}

func New(
	config Config,
	jwtAuthorizer authorizer.JWTAuthorizer,
	statsService *stats.Service,
	geoipResolver *geoip.Resolver,
) (*Instance, error) {
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

// admin.Handler implementation
func (instance *Instance) KillActiveUserSessions(userId string) {
}
