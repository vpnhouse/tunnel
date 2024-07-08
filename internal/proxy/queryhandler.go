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

	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"go.uber.org/zap"
)

var (
	hasPort     = regexp.MustCompile(`:\d+$`)
	connCounter atomic.Int64
)

type ProxyQuery struct {
	id            int64
	userId        string
	userInfo      *userInfo
	proxyInstance *Instance
}

func remoteEndpoint(r *http.Request) string {
	host := r.URL.Host
	if !hasPort.MatchString(host) {
		host += ":80"
	}

	return host
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

func (query *ProxyQuery) doPairedForward(wg *sync.WaitGroup, src, dst io.ReadWriteCloser, tag string) {
	defer wg.Done()
	defer dst.Close()

	for {
		buffer := make([]byte, 4096)
		len, err := src.Read(buffer)
		if err != nil {
			if !isConnectionClosed(err) {
				zap.L().Error("Cant read from source", zap.Error(err), zap.String("tag", tag), zap.Int64("id", query.id))
			}
			return
		}

		// TODO: Handle length
		_, err = dst.Write(buffer[:len])
		if err != nil {
			if !isConnectionClosed(err) {
				zap.L().Error("Cant write to destination", zap.Error(err), zap.String("tag", tag), zap.Int64("id", query.id))
			}
			return
		}

	}
}

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
		http.Error(w, "Limit exceeded", http.StatusTooManyRequests)
		xhttp.WriteJsonError(w, err)
		return
	}
	defer instance.users.release(userId, user)

	conn := &ProxyQuery{
		userId:        userId,
		userInfo:      user,
		id:            connCounter.Add(1),
		proxyInstance: instance,
	}

	if r.Method == "CONNECT" {
		if r.ProtoMajor == 1 {
			conn.handleV1Connect(w, r)
			return
		}

		if r.ProtoMajor == 2 {
			conn.handleV2Connect(w, r)
			return
		}

		http.Error(w, "Unsupported protocol version", http.StatusHTTPVersionNotSupported)
		return
	} else {
		conn.handleProxy(w, r)
	}
}
