// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/Codename-Uranium/tunnel/pkg/xap"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	AudienceAuth     = "auth"
	AudienceDiscover = "discover"
	AudienceTunnel   = "tunnel"

	jwtSigningMethod = "RS256"
	jwtKeyID         = "kid"
)

type StringList []string

func (l StringList) Has(entry string) bool {
	for _, e := range l {
		if e == entry {
			return true
		}
	}

	return false
}

type ClientClaims struct {
	Audience StringList `json:"aud,omitempty"`
	jwt.StandardClaims
}

// KeyStoreWrapper wraps any type into its closure func `fn`
// and provides the KeyStore interface.
type KeyStoreWrapper struct {
	Fn func(keyUUID uuid.UUID) (*rsa.PublicKey, error)
}

func (w *KeyStoreWrapper) GetKey(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
	return w.Fn(keyUUID)
}

type KeyStore interface {
	GetKey(keyUUID uuid.UUID) (*rsa.PublicKey, error)
}

type JWTChecker struct {
	keys   KeyStore
	method jwt.SigningMethod
}

// NewJWTChecker creates new JWT validator that uses keys from a given keystore
func NewJWTChecker(keyKeeper KeyStore) (*JWTChecker, error) {
	method := jwt.GetSigningMethod(jwtSigningMethod)
	if method == nil {
		return nil, xerror.EInvalidArgument("signing method is not supported", nil, zap.String("method", jwtSigningMethod))
	}

	return &JWTChecker{
		keys:   keyKeeper,
		method: method,
	}, nil
}

func (instance *JWTChecker) keyHelper(token *jwt.Token) (interface{}, error) {
	keyIdValue, ok := token.Header[jwtKeyID]
	if !ok {
		return nil, xerror.EAuthenticationFailed("invalid token", nil)
	}
	zap.L().Debug("Got key id", zap.Any("keyIdValue", keyIdValue))

	keyID, ok := keyIdValue.(string)
	if !ok {
		return nil, xerror.EAuthenticationFailed("got unexpected key value instead of string",
			nil, xap.ZapType(keyIdValue))
	}

	keyUUID, err := uuid.Parse(keyID)
	if err != nil {
		return nil, xerror.EAuthenticationFailed("invalid token", err)
	}

	key, err := instance.keys.GetKey(keyUUID)
	if err != nil {
		return nil, err
	}

	return key, nil

}

func (instance *JWTChecker) Parse(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, instance.keyHelper)
	if err != nil {
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	if !token.Valid {
		return xerror.EAuthenticationFailed("invalid token", nil)
	}

	method := token.Method.Alg()
	if method != instance.method.Alg() {
		return xerror.EAuthenticationFailed(
			"invalid token",
			fmt.Errorf("invalid signing method"),
			zap.String("method", method),
			zap.Any("token", token),
		)
	}

	return nil
}
