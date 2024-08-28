package proxy

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

const (
	headerProxyAuthorization = "Proxy-Authorization"
	authTypeBasic            = "basic"
)

func extractProxyAuthToken(r *http.Request) (string, bool) {
	authToken, ok := xhttp.ExtractTokenFromRequest(r)
	if ok {
		return authToken, ok
	}

	authType, authInfo := xhttp.ExtractAuthorizationInfo(r, headerProxyAuthorization)
	if authInfo == "" {
		return "", false
	}

	if strings.ToLower(authType) != authTypeBasic {
		zap.L().Debug("Invalid authentication type")
		return "", false
	}

	userpass, err := base64.StdEncoding.DecodeString(authInfo)
	if err != nil {
		zap.L().Debug("Failed to extract authentication token", zap.Error(err))
		return "", false
	}

	return string(userpass[:len(userpass)-1]), true
}
