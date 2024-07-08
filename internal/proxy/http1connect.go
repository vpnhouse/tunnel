package proxy

import (
	"net"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

func (query *ProxyQuery) handleV1Connect(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Processing v1 CONNECT", zap.Int64("id", query.id))

	remoteConn, err := net.DialTimeout("tcp", remoteEndpoint(r), query.proxyInstance.config.ConnTimeout)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijack not supported", http.StatusServiceUnavailable)
		zap.L().Error("Hijacking is not supported", zap.Int64("id", query.id))
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		if r.Body != nil {
			defer r.Body.Close()
		}
		zap.L().Error("Hijack error", zap.Error(err), zap.Int64("id", query.id))
		return
	}

	if _, err := clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		clientConn.Close()
		remoteConn.Close()
		if !isConnectionClosed(err) {
			zap.L().Error("Can't write 200 OK response", zap.Error(err), zap.Int64("id", query.id))
		}
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go query.doPairedForward(&wg, clientConn, remoteConn, "c2r")
	go query.doPairedForward(&wg, remoteConn, clientConn, "r2c")
	wg.Wait()
}
