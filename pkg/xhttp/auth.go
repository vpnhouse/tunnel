// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xhttp

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

const (
	headerAuthorization = "Authorization"
	authTypeBearer      = "bearer"
)

func AuthIsBearer(authType string) bool {
	return strings.ToLower(authType) == authTypeBearer
}

func ExtractAuthorizationInfo(r *http.Request, header string) (authType string, authInfo string) {
	authHeader := r.Header.Get(header)
	if authHeader == "" {
		return "", ""
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 {
		return "", ""
	}

	return authHeaderParts[0], authHeaderParts[1]
}

func ExtractTokenFromRequest(r *http.Request) (string, bool) {
	authType, authToken := ExtractAuthorizationInfo(r, headerAuthorization)
	if authToken == "" {
		zap.L().Debug("Authentication token was not found")
		return "", false
	}

	if !AuthIsBearer(authType) {
		zap.L().Debug("Invalid authentication type")
		return "", false
	}

	return authToken, true
}
