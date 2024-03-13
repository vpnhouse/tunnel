package proxy

import (
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"syscall"

	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

var (
	hasPort = regexp.MustCompile(`:\d+$`)
	isURL   = regexp.MustCompile(`^(https?):\/\/([^\/]+)(\/(.*)$)?`)
)

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

func extractAnyToken(r *http.Request) (userToken string, ok bool) {
	userToken, ok = xhttp.ExtractProxyTokenFromRequest(r)
	if ok {
		return
	}

	return xhttp.ExtractTokenFromRequest(r)
}
