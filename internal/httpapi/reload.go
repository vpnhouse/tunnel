package httpapi

import (
	"net/http"

	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
)

// AdminReloadService reloads server with new configuration
func (instance *TunnelAPI) AdminReloadService(w http.ResponseWriter, r *http.Request) {
	// ask the default wrapper to write OK string to the client conn
	xhttp.JSONResponse(w, func() (interface{}, error) { return nil, nil })
	w.(http.Flusher).Flush()
	instance.runtime.Events.EmitEvent(control.EventRestart)
}
