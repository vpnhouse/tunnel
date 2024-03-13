package proxy

import (
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

type proxyHandlerFunc func(w http.ResponseWriter, r *http.Request) (int, string)

func (instance *Instance) doAuth(r *http.Request) (string, error) {
	userToken, ok := extractAnyToken(r)
	if !ok {
		return "", xerror.EAuthenticationFailed("no auth token", nil)
	}

	token, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return "", err
	}

	return token.UserId, nil
}

func (instance *Instance) doProxyConnect(w http.ResponseWriter, r *http.Request) (int, string) {
	hij, ok := w.(http.Hijacker)
	if !ok {
		zap.L().Error("Hijack is not supported")
		return http.StatusInternalServerError, "Hijack is not supported"
	}
	hijConn, _, err := hij.Hijack()
	if err != nil {
		zap.L().Error("Connection hijack failed")
		return http.StatusInternalServerError, "Connection hijack failed"
	}

	host := r.URL.Host
	if !hasPort.MatchString(host) {
		host += ":80"
	}

	conn, err := net.Dial("tcp", host)
	if err != nil {
		hijConn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		hijConn.Close()
		zap.L().Debug("Failed dialing to remote", zap.Error(err))
		return 0, ""
	}

	remoteConn := conn.(*net.TCPConn)
	if _, err := hijConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		hijConn.Close()
		remoteConn.Close()
		if !isConnectionClosed(err) {
			zap.L().Debug("Failed writing status 200 OK", zap.Error(err))
		}
		return 0, ""
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer func() {
			e := recover()
			err, ok := e.(error)
			if !ok {
				return
			}
			hijConn.Close()
			remoteConn.Close()
			if !isConnectionClosed(err) {
				zap.L().Debug("Request read error", zap.Error(err))
			}
		}()
		_, err := io.Copy(remoteConn, hijConn)
		if err != nil {
			panic(err)
		}
		remoteConn.CloseWrite()
		if c, ok := hijConn.(*net.TCPConn); ok {
			c.CloseRead()
		}
	}()
	go func() {
		defer wg.Done()
		defer func() {
			e := recover()
			err, ok := e.(error)
			if !ok {
				return
			}
			hijConn.Close()
			remoteConn.Close()
			if !isConnectionClosed(err) {
				zap.L().Debug("response write error", zap.Error(err))
			}
		}()
		_, err := io.Copy(hijConn, remoteConn)
		if err != nil {
			panic(err)
		}
		remoteConn.CloseRead()
		if c, ok := hijConn.(*net.TCPConn); ok {
			c.CloseWrite()
		}
	}()
	wg.Wait()
	hijConn.Close()
	remoteConn.Close()

	return 0, ""
}

func (instance *Instance) doProxy(w http.ResponseWriter, r *http.Request, handler proxyHandlerFunc) {
	userId, err := instance.doAuth(r)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}

	code, status := func() (int, string) {
		user, err := instance.users.acquire(userId)
		if err != nil {
			return http.StatusGone, "Gone"
		}
		defer instance.users.release(userId, user)

		return handler(w, r)
	}()

	if code != 0 {
		w.WriteHeader(code)
		w.Write([]byte(status))
	}
}

func (instance *Instance) doProxyHttp(w http.ResponseWriter, r *http.Request) (int, string) {
	// http.StatusAccepted
	// return xerror.EUnavailable("Not implemented yet", nil)
	return http.StatusTeapot, "I'm a teapot"
}

func (instance *Instance) ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Debug("Query", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		if r.Method == http.MethodConnect {
			instance.doProxy(w, r, instance.doProxyConnect)
			return
		}
		if isURL.MatchString(r.RequestURI) {
			instance.doProxy(w, r, instance.doProxyHttp)
			return
		}
		next.ServeHTTP(w, r)
	})
}
