package authorizer

import (
	"github.com/Codename-Uranium/tunnel/pkg/xcrypto"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
)

type InternalAuthorizer struct {
	checker *xcrypto.JWTChecker
	running bool
}

func NewInternalAuthorizer(keyKeeper xcrypto.KeyStore) (*InternalAuthorizer, error) {
	checker, err := xcrypto.NewJWTChecker(keyKeeper)

	if err != nil {
		return nil, err
	}

	return &InternalAuthorizer{
		checker: checker,
		running: true,
	}, nil
}

func (d *InternalAuthorizer) Shutdown() error {
	d.running = false
	return nil
}

func (d *InternalAuthorizer) Running() bool {
	return d.running
}

func (d *InternalAuthorizer) Authenticate(tokenString string, myAudience string) (*xcrypto.ClientClaims, error) {
	var claims xcrypto.ClientClaims

	err := d.checker.Parse(tokenString, &claims)
	if err != nil {
		return nil, err
	}

	if !claims.Audience.Has(myAudience) {
		return nil, xerror.EAuthenticationFailed("invalid audience", nil)
	}

	return &claims, nil
}
