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
	"github.com/go-chi/chi/v5"
)

type TunnelAPI struct {
	runtime    *runtime.TunnelRuntime
	manager    *manager.Manager
	adminJWT   *auth.JWTMaster
	authorizer *authorizer.JWTAuthorizer
	storage    *storage.Storage
	keystore   federation_keys.Keystore
	running    bool
}

func NewTunnelHandlers(
	runtime *runtime.TunnelRuntime,
	manager *manager.Manager,
	adminJWT *auth.JWTMaster,
	authorizer *authorizer.JWTAuthorizer,
	storage *storage.Storage,
	keystore federation_keys.Keystore,
) *TunnelAPI {
	instance := &TunnelAPI{
		runtime:    runtime,
		manager:    manager,
		adminJWT:   adminJWT,
		authorizer: authorizer,
		storage:    storage,
		keystore:   keystore,
		running:    true,
	}

	return instance
}

func (instance *TunnelAPI) RegisterHandlers(r chi.Router) {
	// handle frontend redirects
	root := instance.runtime.Settings.AdminAPI.StaticRoot
	r.HandleFunc("/", wrap404ToIndex(http.FileServer(http.Dir(root))))

	// admin API
	adminAPI.HandlerWithOptions(instance, adminAPI.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []adminAPI.MiddlewareFunc{
			instance.adminAuthMiddleware,
		},
	})

	if instance.runtime.Features.WithPublicAPI() {
		tunnelAPI.HandlerWithOptions(instance, tunnelAPI.ChiServerOptions{
			BaseRouter: r,
		})
	}

	if instance.runtime.Features.WithFederation() {
		mgmtAPI.HandlerWithOptions(instance, mgmtAPI.ChiServerOptions{
			BaseRouter: r,
			Middlewares: []mgmtAPI.MiddlewareFunc{
				instance.federationAuthMiddleware,
			},
		})
	}
}
