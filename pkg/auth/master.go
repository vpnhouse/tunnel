// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package auth

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/vpnhouse/tunnel/pkg/xcrypto"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type JWTMaster struct {
	keyID   *uuid.UUID
	private *rsa.PrivateKey
	method  jwt.SigningMethod
}

func NewJWTMaster(private *rsa.PrivateKey, privateId *uuid.UUID) (*JWTMaster, error) {
	// Generate new private key if it's not given by caller
	if private == nil {
		if privateId != nil {
			return nil, xerror.EInternalError("privateId must be nil when private is nil", nil)
		}

		vPrivateId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		privateId = &vPrivateId

		zap.L().Info("generating keys for JWT")
		private, err = xcrypto.GenerateKey()
		if err != nil {
			return nil, xerror.EInternalError("can't generate JWT key pair", err)
		}
	} else {
		if privateId == nil {
			return nil, xerror.EInvalidArgument("privateId must be set when private is set", nil)
		}
	}

	method := jwt.GetSigningMethod(jwtSigningMethod)
	if method == nil {
		return nil, xerror.EInvalidArgument("signing method is not supported", nil, zap.String("method", jwtSigningMethod))
	}

	return &JWTMaster{
		private: private,
		keyID:   privateId,
		method:  method,
	}, nil
}

func (instance *JWTMaster) Token(claims jwt.Claims) (*string, error) {
	// Create token
	token := jwt.NewWithClaims(instance.method, claims)
	token.Header["kid"] = instance.keyID

	// Sign token
	signedToken, err := token.SignedString(instance.private)
	if err != nil {
		zap.L().Error("Can't sign auth token", zap.Error(err))
		return nil, xerror.EInternalError("can't sign token", err)
	}

	return &signedToken, nil
}

func (instance *JWTMaster) Parse(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return instance.private.Public(), nil
	})

	if err != nil || token == nil {
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	if !token.Valid {
		return xerror.EAuthenticationFailed("invalid token", nil)
	}

	method := token.Method.Alg()
	if method != instance.method.Alg() {
		zap.L().Error("Invalid signing method", zap.String("method", method), zap.Any("token", token))
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	return nil
}
