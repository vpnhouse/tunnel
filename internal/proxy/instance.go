package proxy

import (
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

type Instance struct {
	terminated atomic.Bool
	runtime    *runtime.TunnelRuntime
	authorizer authorizer.JWTAuthorizer
}

func New(runtime *runtime.TunnelRuntime, jwtAuthorizer authorizer.JWTAuthorizer) *Instance {
	return &Instance{
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		runtime:    runtime,
	}
}

func (instance *Instance) doAuth(r *http.Request) error {
	// Extract JWT
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return xerror.EAuthenticationFailed("no auth token", nil)
	}

	// Verify JWT, get JWT claims
	_, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return err
	}

	return nil
}

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	i := 0
	var newerr = &err
	for opError, ok := (*newerr).(*net.OpError); ok && i < 10; {
		i++
		newerr = &opError.Err
		if syscallError, ok := (*newerr).(*os.SyscallError); ok {
			if syscallError.Err == syscall.EPIPE || syscallError.Err == syscall.ECONNRESET || syscallError.Err == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}

var hasPort = regexp.MustCompile(`:\d+$`)

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
	err := instance.doAuth(r)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}

	err = instance.doConnect(w, r)
	if err != nil {
		xhttp.WriteJsonError(w, err)
		return
	}
}

func (instance *Instance) RegisterHandlers(r chi.Router) {
	r.MethodFunc("CONNECT", "*", instance.handler)
}

func (instance *Instance) Shutdown() error {
	if instance.terminated.Swap(true) {
		return xerror.EInternalError("Double proxy shutdown", nil)
	}

	return nil
}

func (instance *Instance) Running() bool {
	return !instance.terminated.Load()
}
