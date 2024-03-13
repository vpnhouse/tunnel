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

func (instance *Instance) doConnect(w http.ResponseWriter, r *http.Request) error {
	hij, ok := w.(http.Hijacker)
	if !ok {
		return xerror.EInternalError("Hijack not supported", nil)
	}
	hijConn, _, err := hij.Hijack()
	if err != nil {
		return xerror.EInternalError("Connection hijacking failed", nil)

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
		return nil
	}
	if err != nil {
		return xerror.EEntryNotFound("Connection failed", err)
	}

	remoteConn := conn.(*net.TCPConn)
	if _, err := hijConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		hijConn.Close()
		remoteConn.Close()
		if !isConnectionClosed(err) {
			zap.L().Debug("Failed writing status 200 OK", zap.Error(err))
		}
		return nil
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

	return nil
}

func (instance *Instance) handler(w http.ResponseWriter, r *http.Request) {
	userId, err := instance.doAuth(r)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}

	user, err := instance.users.acquire(userId)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}
	defer instance.users.release(userId, user)

	err = instance.doConnect(w, r)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}
}
