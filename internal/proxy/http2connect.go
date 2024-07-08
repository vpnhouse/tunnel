package proxy

import (
	"net"
	"net/http"
	"sync"

	"github.com/posener/h2conn"
	"go.uber.org/zap"
)

func (query *ProxyQuery) handleV2Connect(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Processing v2 CONNECT", zap.Int64("id", query.id))

	remoteConn, err := net.DialTimeout("tcp", remoteEndpoint(r), query.proxyInstance.config.ConnTimeout)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	clientConn, err := h2conn.Accept(w, r)
	if err != nil {
		zap.L().Error("h2conn error", zap.Error(err), zap.Int64("id", query.id))
		remoteConn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go query.doPairedForward(&wg, clientConn, remoteConn, "c2r")
	go query.doPairedForward(&wg, remoteConn, clientConn, "r2c")
	wg.Wait()
}
