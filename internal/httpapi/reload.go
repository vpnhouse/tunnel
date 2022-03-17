// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"net/http"

	"github.com/vpnhouse/tunnel/pkg/control"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

// AdminReloadService reloads server with new configuration
func (tun *TunnelAPI) AdminReloadService(w http.ResponseWriter, r *http.Request) {
	// ask the default wrapper to write OK string to the client conn
	xhttp.JSONResponse(w, func() (interface{}, error) { return nil, nil })
	w.(http.Flusher).Flush()
	tun.runtime.Events.EmitEvent(control.EventRestart)
}
