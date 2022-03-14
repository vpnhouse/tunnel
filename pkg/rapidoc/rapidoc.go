// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package rapidoc

import (
	"net/http"

	apiDoc "github.com/comradevpn/api"
	"github.com/comradevpn/tunnel/pkg/version"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

func RegisterHandlers(r chi.Router) {
	zap.L().Info("Registering rapidoc handlers")

	docFs := http.FS(apiDoc.Docs)
	r.Handle("/schemas/*", http.FileServer(docFs))

	if version.IsPersonal() {
		// for the personal version - serve the Admin API docs as an index page
		index, err := apiDoc.Docs.ReadFile("rapidoc/tunnel_admin.html")
		if err != nil {
			zap.L().Error("rapidoc seems misconfigured, failed to open `rapidoc/tunnel_admin.html`", zap.Error(err))
			return
		}

		r.HandleFunc("/rapidoc/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(index)
		})
	} else {
		// server all docs for non-personal versions
		r.Handle("/rapidoc/*", http.FileServer(docFs))
	}
}
