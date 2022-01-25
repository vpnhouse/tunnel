package httpapi

import (
	"net/http"

	libControl "github.com/Codename-Uranium/common/control"
	"github.com/Codename-Uranium/common/xhttp"
	"go.uber.org/zap"
)

// AdminReloadService reloads server with new configuration
func (instance *TunnelAPI) AdminReloadService(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Reload", zap.Any("info", xhttp.RequestInfo(r)))

	// ask the default wrapper to write OK string to the client conn
	xhttp.JSONResponse(w, func() (interface{}, error) { return nil, nil })
	w.(http.Flusher).Flush()
	instance.runtime.Events.EmitEvent(libControl.EventRestart)
}
