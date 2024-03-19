package proxy

import (
	"net/http"

	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

func (instance *Instance) doAuth(r *http.Request) (string, error) {
	userToken, ok := extractProxyAuthToken(r)
	if !ok {
		return "", xerror.EAuthenticationFailed("no auth token", nil)
	}

	token, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return "", err
	}

	return token.UserId, nil
}

func (instance *Instance) doProxy(w http.ResponseWriter, r *http.Request) {
	userId, err := instance.doAuth(r)
	if err != nil {
		w.Header()["Proxy-Authenticate"] = []string{"Basic realm=\"proxy\""}
		w.WriteHeader(http.StatusProxyAuthRequired)
		w.Write([]byte("Proxy authentication required"))
		return
	}

	user, err := instance.users.acquire(r.Context(), userId)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}
	defer instance.users.release(userId, user)

	instance.proxy.ServeHTTP(w, r)
}

func (instance *Instance) ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Debug("Query", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		if r.Method == http.MethodConnect || r.URL.IsAbs() {
			instance.doProxy(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
