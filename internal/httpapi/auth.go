// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"net/http"
	"time"

	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

// AdminDoAuth implements handler for GET /api/tunnel/admin/auth
func (tun *TunnelAPI) AdminDoAuth(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		authOK := false

		// Check if basic authentication is successful
		username, password, haveBasic := r.BasicAuth()
		if haveBasic {
			zap.L().Debug("found basic authentication")
			err := tun.adminCheckBasicAuth(username, password)
			if err != nil {
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
