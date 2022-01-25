package authorizer

import (
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/google/uuid"
)

type AnonymousAuthorizer struct {
	tokenLifetime int64
	running       bool
}

func NewAnonymousAuthorizer(tokenLifetime int64) (*AnonymousAuthorizer, error) {
	return &AnonymousAuthorizer{
		tokenLifetime: tokenLifetime,
		running:       true,
	}, nil
}

func (a *AnonymousAuthorizer) Shutdown() error {
	a.running = false
	return nil
}

func (a *AnonymousAuthorizer) Running() bool {
	return a.running
}

func (a *AnonymousAuthorizer) Authenticate(authInfo string) (string, int64, error) {
	if !a.running {
		return "", 0, xerror.EInternalError("Authenticate attempt on stopped anonymous authorizer", nil)
	}

	_, err := uuid.Parse(authInfo)
	if err != nil {
		return "", 0, xerror.EAuthenticationFailed("Invalid token format", err)
	}

	return authInfo, a.tokenLifetime, nil
}
