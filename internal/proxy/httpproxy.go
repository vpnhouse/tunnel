package proxy

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

var (
	customTransport = http.DefaultTransport
)

func (query *ProxyQuery) handleProxy(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Processing proxy request", zap.Int64("id", query.id))

	proxyReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		zap.L().Error("Error creating proxy request", zap.Error(err), zap.Int64("id", query.id))
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Send the proxy request using the custom transport
	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
		zap.L().Debug("Error sending proxy request", zap.Error(err), zap.Int64("id", query.id))
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	io.Copy(w, resp.Body)
}
