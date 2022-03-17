// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	tunnelAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

// AdminDoAuth implements handler for GET /api/tunnel/admin/auth
func (tun *TunnelAPI) AdminDoAuth(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		authOK := false

		// Check if basic authentication is successful
		if _, password, ok := r.BasicAuth(); ok {
			zap.L().Debug("found basic authentication")
			if err := tun.runtime.Settings.VerifyAdminPassword(password); err != nil {
				return nil, err
			}
			authOK = true
		}

		if !authOK {
			// Check if bearer authentication is successful
			tokenStr, haveBearer := xhttp.ExtractTokenFromRequest(r)
			if haveBearer {
				zap.L().Debug("found bearer authentication")
				err := tun.adminCheckBearerAuth(tokenStr)
				if err != nil {
					return nil, err
				}
				authOK = true
			}
		}

		if !authOK {
			return nil, xerror.EUnauthorized("basic or bearer authentication expected", nil)
		}

		// Create claims
		issued := time.Now().Unix()
		expires := issued + int64(tun.runtime.Settings.AdminAPI.TokenLifetime)
		claims := jwt.StandardClaims{
			IssuedAt:  issued,
			ExpiresAt: expires,
		}

		signedToken, err := tun.adminJWT.Token(&claims)
		if err != nil {
			return nil, err
		}

		response := &tunnelAPI.AdminAuthResponse{
			AccessToken: *signedToken,
		}
		return response, nil
	})
}
