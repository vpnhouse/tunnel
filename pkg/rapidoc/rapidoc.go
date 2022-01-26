package rapidoc

import (
	_ "embed"
	"net/http"

	apiDoc "github.com/Codename-Uranium/api"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var rapidocEnabled = false

func Switch(enabled bool) {
	rapidocEnabled = enabled
}

func RegisterHandlers(r chi.Router) {
	if !rapidocEnabled {
		return
	}

	zap.L().Info("Registering rapidoc handler")

	fs := http.FS(apiDoc.Docs)
	r.Handle("/rapidoc/", http.FileServer(fs))
	r.Handle("/schemas/", http.FileServer(fs))
}
