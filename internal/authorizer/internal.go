// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package authorizer

import (
	"context"
	"sync/atomic"

	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/xerror"
)

type JWTAuthorizer interface {
	Authenticate(ctx context.Context, tokenString string, myAudience string) (*auth.ClientClaims, error)
}

type AuthClient func(ctx context.Context, clientClaims *auth.ClientClaims) error

type JWTOption func(opts *JWTOptions)

type JWTOptions struct {
	AuthClient AuthClient
}

func WithAuthClient(authClient AuthClient) JWTOption {
	return func(opts *JWTOptions) {
		opts.AuthClient = authClient
	}
}

var _ JWTAuthorizer = (*jwtAuthorizer)(nil)

type jwtAuthorizer struct {
	checker    *auth.JWTChecker
	authClient AuthClient
	running    atomic.Bool
}

func NewJWT(keyKeeper auth.KeyStore, opts ...JWTOption) (*jwtAuthorizer, error) {
	checker, err := auth.NewJWTChecker(keyKeeper)
	if err != nil {
		return nil, err
	}

	var options JWTOptions
	for _, o := range opts {
		o(&options)
	}

	jwtAuth := &jwtAuthorizer{
		checker:    checker,
		authClient: options.AuthClient,
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

func (d *jwtAuthorizer) Authenticate(ctx context.Context, tokenString string, myAudience string) (*auth.ClientClaims, error) {
	var claims auth.ClientClaims

	err := d.checker.Parse(tokenString, &claims)
	if err != nil {
		return nil, err
	}

	if !claims.Audience.Has(myAudience) {
		return nil, xerror.EAuthenticationFailed("invalid audience", nil)
	}

	if d.authClient != nil {
		err := d.authClient(ctx, &claims)
		if err != nil {
			return nil, err
		}
	}

	return &claims, nil
}
