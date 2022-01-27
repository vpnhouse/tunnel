package httpapi

import (
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
)

// AdminGetStatus returns current server status
func (tun *TunnelAPI) AdminGetStatus(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		flags := tun.runtime.Flags
		status := adminAPI.ServiceStatusResponse{
			RestartRequired: flags.RestartRequired,
		}
		return status, nil
	})
}
