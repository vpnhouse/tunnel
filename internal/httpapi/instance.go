package httpapi

import (
	"net/http"

	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	mgmtAPI "github.com/Codename-Uranium/api/go/server/tunnel_mgmt"
	"github.com/Codename-Uranium/tunnel/internal/authorizer"
	"github.com/Codename-Uranium/tunnel/internal/federation_keys"
	"github.com/Codename-Uranium/tunnel/internal/manager"
	"github.com/Codename-Uranium/tunnel/internal/runtime"
	"github.com/Codename-Uranium/tunnel/internal/storage"
	"github.com/Codename-Uranium/tunnel/pkg/auth"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"github.com/go-chi/chi/v5"
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
	// handle frontend redirects
	root := tun.runtime.Settings.AdminAPI.StaticRoot
	r.HandleFunc("/", wrap404ToIndex(http.FileServer(http.Dir(root))))

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
