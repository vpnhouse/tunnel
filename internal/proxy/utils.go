package proxy

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"syscall"
)

var (
	hasPort       = regexp.MustCompile(`:\d+$`)
	randomLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func randomString(n uint) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randomLetters[rand.Intn(len(randomLetters))]
	}
	return string(b)
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

	var syscallError *os.SyscallError
	if errors.As(err, &syscallError) {
		if syscallError.Err == syscall.EPIPE || syscallError.Err == syscall.ECONNRESET || syscallError.Err == syscall.EPROTOTYPE {
			return true
		}
	}

	return false
}
