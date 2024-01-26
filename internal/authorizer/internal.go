// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package authorizer

import (
	"sync/atomic"

	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

type JWTAuthorizer interface {
	Authenticate(tokenString string, myAudience string) (*auth.ClientClaims, error)
}

var _ JWTAuthorizer = (*jwtAuthorizer)(nil)

type jwtAuthorizer struct {
	checker *auth.JWTChecker
	running atomic.Bool
}

func NewJWT(keyKeeper auth.KeyStore) (*jwtAuthorizer, error) {
	checker, err := auth.NewJWTChecker(keyKeeper)

	if err != nil {
		return nil, err
	}

	jwtAuth := &jwtAuthorizer{
		checker: checker,
	}
	jwtAuth.running.Store(true)

	return jwtAuth, nil
}

func (d *jwtAuthorizer) Shutdown() error {
	d.running.Store(false)
	return nil
}

func (d *jwtAuthorizer) Running() bool {
	return d.running.Load()
}

func (d *jwtAuthorizer) Authenticate(tokenString string, myAudience string) (*auth.ClientClaims, error) {
	var claims auth.ClientClaims

	err := d.checker.Parse(tokenString, &claims)
	if err != nil {
		return nil, err
	}

	if !claims.Audience.Has(myAudience) {
		return nil, xerror.EAuthenticationFailed("invalid audience", nil)
	}

	return &claims, nil
}
