package rapidoc

import (
	_ "embed"
	"net/http"

	apiDoc "github.com/Codename-Uranium/api"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"

	"go.uber.org/zap"
)

var rapidocEnabled = false

func Switch(enabled bool) {
	rapidocEnabled = enabled
}

func Handlers() xhttp.Handlers {

	if !rapidocEnabled {
		return xhttp.Handlers{}
	}

	fs := http.FS(apiDoc.Docs)

	zap.L().Warn("Registering rapidoc handler")
	return xhttp.Handlers{
		"/rapidoc/": http.FileServer(fs),
		"/schemas/": http.FileServer(fs),
	}
}
