// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package authorizer

import (
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

type JWTAuthorizer struct {
	checker *auth.JWTChecker
	running bool
}

func NewJWT(keyKeeper auth.KeyStore) (*JWTAuthorizer, error) {
	checker, err := auth.NewJWTChecker(keyKeeper)

	if err != nil {
		return nil, err
	}

	return &JWTAuthorizer{
		checker: checker,
		running: true,
	}, nil
}

func (d *JWTAuthorizer) Shutdown() error {
	d.running = false
	return nil
}

func (d *JWTAuthorizer) Running() bool {
	return d.running
}

func (d *JWTAuthorizer) Authenticate(tokenString string, myAudience string) (*auth.ClientClaims, error) {
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
