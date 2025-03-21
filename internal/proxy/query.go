package proxy

import (
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/posener/h2conn"
	"github.com/vpnhouse/common-lib-go/stats"
	"go.uber.org/zap"
)

var (
	queryCounter    atomic.Int64
	customTransport = http.DefaultTransport
)

type reporter func(n uint64)

type ProxyQuery struct {
	id            int64
	auth          *authInfo
	proxyInstance *Instance
}

func (query *ProxyQuery) doPairedForward(wg *sync.WaitGroup, src, dst io.ReadWriteCloser, rep reporter) {
	defer wg.Done()
	defer dst.Close()

	for {
		buffer := make([]byte, 4096)
		n, err := src.Read(buffer)
		if err != nil {
			return
		}

		// TODO: Handle length
		n, err = dst.Write(buffer[:n])
		if err != nil {
			return
		}
		rep(uint64(n))
	}
}

func (query *ProxyQuery) handleV1Connect(w http.ResponseWriter, r *http.Request) {
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
	sessionID := uuid.New()
	go query.doPairedForward(&wg, clientConn, remoteConn, query.reporterRx(sessionID))
	go query.doPairedForward(&wg, remoteConn, clientConn, query.reporterTx(sessionID))
	wg.Wait()
}

func (query *ProxyQuery) handleV2Connect(w http.ResponseWriter, r *http.Request) {
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
	sessionID := uuid.New()
	go query.doPairedForward(&wg, clientConn, remoteConn, query.reporterTx(sessionID))
	go query.doPairedForward(&wg, remoteConn, clientConn, query.reporterRx(sessionID))
	wg.Wait()
}

func (query *ProxyQuery) reporterRx(sessionID uuid.UUID) reporter {
	if query.proxyInstance.statsService == nil {
		return func(uint64) {}
	}
	return func(drx uint64) {
		query.proxyInstance.statsService.ReportStats(sessionID, drx, 0, func(sessionID uuid.UUID, sessionData *stats.SessionData) {
			sessionData.InstallationID = query.auth.InstallationID
			sessionData.UserID = query.auth.UserID
			sessionData.Country = query.auth.Country
		})
	}
}

func (query *ProxyQuery) reporterTx(sessionID uuid.UUID) reporter {
	if query.proxyInstance.statsService == nil {
		return func(uint64) {}
	}
	return func(dtx uint64) {
		query.proxyInstance.statsService.ReportStats(sessionID, 0, dtx, func(sessionID uuid.UUID, sessionData *stats.SessionData) {
			sessionData.InstallationID = query.auth.InstallationID
			sessionData.UserID = query.auth.UserID
			sessionData.Country = query.auth.Country
		})
	}
}

func (query *ProxyQuery) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 {
		http.Error(w, "Bad request", http.StatusHTTPVersionNotSupported)
	}

	proxyReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		zap.L().Error("Error creating proxy request", zap.Error(err), zap.Int64("id", query.id))
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	r.Header.Add(query.proxyInstance.proxyMarkHeader, randomString(8))

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Send the proxy request using the custom transport
	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
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
