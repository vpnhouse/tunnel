package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Target struct {
	URL      string   `yaml:"url"`
	Patterns []string `yaml:"patterns"`
}

type Config struct {
	Targets []*Target `yaml:"targets"`
}

func RegisterHandlers(r chi.Router, config *Config) {
	for _, target := range config.Targets {
		targetURL, err := url.Parse(target.URL)
		if err != nil {
			zap.L().Error("SKipping invalid target URL", zap.Error(err), zap.String("target", target.URL))
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = targetURL.Host
		}
		for _, path := range target.Patterns {
			r.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
				proxy.ServeHTTP(w, req)
			})
		}
	}
}
