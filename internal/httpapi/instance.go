package httpapi

import (
	"net/http"

	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	mgmtAPI "github.com/Codename-Uranium/api/go/server/tunnel_mgmt"
	libAuthorizer "github.com/Codename-Uranium/common/authorizer"
	libToken "github.com/Codename-Uranium/common/token"
	"github.com/Codename-Uranium/common/xhttp"
	"github.com/Codename-Uranium/tunnel/internal/federation_keys"
	"github.com/Codename-Uranium/tunnel/internal/manager"
	"github.com/Codename-Uranium/tunnel/internal/runtime"
	"github.com/Codename-Uranium/tunnel/internal/storage"
	"github.com/Codename-Uranium/tunnel/internal/wireguard"
)

type TunnelAPI struct {
	runtime    *runtime.TunnelRuntime
	manager    *manager.Manager
	adminJWT   *libToken.JWTMaster
	wireguard  *wireguard.Wireguard
	authorizer *libAuthorizer.InternalAuthorizer
	storage    *storage.Storage
	keystore   federation_keys.Keystore
	handlers   xhttp.Handlers
	running    bool
}

func NewTunnelHandlers(
	runtime *runtime.TunnelRuntime,
	manager *manager.Manager,
	adminJWT *libToken.JWTMaster,
	wireguard *wireguard.Wireguard,
	authorizer *libAuthorizer.InternalAuthorizer,
	storage *storage.Storage,
	keystore federation_keys.Keystore,
) (*TunnelAPI, error) {
	instance := &TunnelAPI{
		runtime:    runtime,
		manager:    manager,
		adminJWT:   adminJWT,
		wireguard:  wireguard,
		authorizer: authorizer,
		storage:    storage,
		keystore:   keystore,
	}

	publicRoutes := tunnelAPI.Handler(instance)
	adminRoutes := adminAPI.HandlerWithOptions(instance, adminAPI.ChiServerOptions{
		Middlewares: []adminAPI.MiddlewareFunc{
			instance.adminAuthMiddleware,
		},
	})

	federationRoutes := mgmtAPI.HandlerWithOptions(instance, mgmtAPI.ChiServerOptions{
		Middlewares: []mgmtAPI.MiddlewareFunc{
			instance.federationAuthMiddleware,
		},
	})

	instance.handlers = xhttp.Handlers{
		"/":                       wrap404ToIndex(http.FileServer(http.Dir(runtime.Settings.AdminAPI.StaticRoot))),
		"/api/client/":            publicRoutes,
		"/api/tunnel/admin/":      adminRoutes,
		"/api/tunnel/federation/": federationRoutes,
	}

	instance.running = true
	return instance, nil
}

func (instance *TunnelAPI) Handlers() xhttp.Handlers {
	return instance.handlers
}

func (instance *TunnelAPI) Shutdown() error {
	instance.running = false
	return nil
}

func (instance *TunnelAPI) Running() bool {
	return instance.running
}
