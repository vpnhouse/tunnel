package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func RegisterHandlers(r chi.Router, paths []string, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		zap.L().Error("SKipping invalid target URL", zap.Error(err), zap.String("target", target))
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = targetURL.Host
	}
	for _, path := range paths {
		r.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			proxy.ServeHTTP(w, req)
		})
	}
}
