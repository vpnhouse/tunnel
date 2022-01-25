package httpapi

import (
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

// AdminGetStatus returns current server status
func (instance *TunnelAPI) AdminGetStatus(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("GetStatus", zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		flags := instance.runtime.Flags
		status := adminAPI.ServiceStatusResponse{
			RestartRequired: flags.RestartRequired,
		}
		return status, nil
	})
}
