// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"io"
	"net/http"

	tunnelAPI "github.com/comradevpn/api/go/server/tunnel"
	adminAPI "github.com/comradevpn/api/go/server/tunnel_admin"
	mgmtAPI "github.com/comradevpn/api/go/server/tunnel_mgmt"
	"github.com/comradevpn/tunnel/internal/authorizer"
	"github.com/comradevpn/tunnel/internal/federation_keys"
	"github.com/comradevpn/tunnel/internal/manager"
	"github.com/comradevpn/tunnel/internal/runtime"
	"github.com/comradevpn/tunnel/internal/storage"
	"github.com/comradevpn/tunnel/pkg/auth"
	"github.com/comradevpn/tunnel/pkg/xerror"
	"github.com/comradevpn/tunnel/pkg/xnet"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type IpPool interface {
	Available() (xnet.IP, error)
	IsAvailable(ip xnet.IP) bool
}

type TunnelAPI struct {
	runtime    *runtime.TunnelRuntime
	manager    *manager.Manager
	adminJWT   *auth.JWTMaster
	authorizer *authorizer.JWTAuthorizer
	storage    *storage.Storage
	keystore   federation_keys.Keystore
	ippool     IpPool
	running    bool
}

func NewTunnelHandlers(
	runtime *runtime.TunnelRuntime,
	manager *manager.Manager,
	adminJWT *auth.JWTMaster,
	authorizer *authorizer.JWTAuthorizer,
	storage *storage.Storage,
	keystore federation_keys.Keystore,
	ippool IpPool,
) *TunnelAPI {
	instance := &TunnelAPI{
		runtime:    runtime,
		manager:    manager,
		adminJWT:   adminJWT,
		authorizer: authorizer,
		storage:    storage,
		keystore:   keystore,
		ippool:     ippool,
		running:    true,
	}

	return instance
}

func (tun *TunnelAPI) RegisterHandlers(r chi.Router) {
	webroot := http.Dir(tun.runtime.Settings.AdminAPI.StaticRoot)
	webfs := http.FileServer(webroot)
	// handle frontend redirects
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		nfrw := &skipNotFoundWriter{ResponseWriter: w}
		webfs.ServeHTTP(nfrw, r)
		// note that skipNotFoundWriter acts as a normal Writer
		//  for all status codes except the NotFound.
		//  So handle it explicitly below.
		if nfrw.status == http.StatusNotFound {
			w.Header().Set("content-type", "text/html")
			// try to load and serve index html instead of non-existing file:
			//  we need that because very likely the client asks for some /path
			//  that is handled by the SPAs router. So we always have to serve
			//  the frontend app instead of the requested path.
			fd, err := webroot.Open("index.html")
			if err != nil {
				zap.L().Warn("no `index.html` found on webFS, telling user that there is no frontend deployed")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error: no frontend files deployed."))
				return
			}

			defer fd.Close()
			if _, err := io.Copy(w, fd); err != nil {
				_ = xerror.WInternalError("xhttp", "failed to write index.html bytes to the http connection", err)
			}
		}
	})

	r.Delete("/_debug/reset-initial", tun.TmpResetSettingsToDefault)

	// admin API
	adminAPI.HandlerWithOptions(tun, adminAPI.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []adminAPI.MiddlewareFunc{
			tun.adminAuthMiddleware,
			tun.initialSetupMiddleware,
			tun.versionRestrictionsMiddleware,
		},
	})

	if tun.runtime.Features.WithPublicAPI() {
		tunnelAPI.HandlerWithOptions(tun, tunnelAPI.ChiServerOptions{
			BaseRouter: r,
		})
	}

	if tun.runtime.Features.WithFederation() {
		mgmtAPI.HandlerWithOptions(tun, mgmtAPI.ChiServerOptions{
			BaseRouter: r,
			Middlewares: []mgmtAPI.MiddlewareFunc{
				tun.federationAuthMiddleware,
			},
		})
	}
}
