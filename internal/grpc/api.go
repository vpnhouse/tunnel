package grpc

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	federationAuthHeader = "X-VPNHOUSE-FEDERATION-KEY"
	tunnelAuthHeader     = "X-VPNHOUSE-TUNNEL-KEY"
)

func (g *grpcServer) RegisterHandlers(r chi.Router) {
	r.HandleFunc("/grpc/ca", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(federationAuthHeader)
		_, ok := g.keystore.Authorize(authHeader)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var caList []string
		if g.ca != "" {
			caList = append(caList, g.ca)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(tunnelAuthHeader, g.tunnelKey)
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(map[string]interface{}{
			"ca": caList,
		})
	})
}
