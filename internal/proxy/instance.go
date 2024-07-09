package proxy

import (
	"math/rand"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type Config struct {
	ConnLimit              int           `yaml:"conn_limit"`
	ConnTimeout            time.Duration `yaml:"conn_timeout"`
	MarkHeaderPrefix       string        `yaml:"mark_header_prefix"`
	MarkHeaderRandomLength uint          `yaml:"mark_header_random_length"`
}

type Instance struct {
	config          *Config
	authorizer      authorizer.JWTAuthorizer
	users           *userStorage
	terminated      atomic.Bool
	myDomains       map[string]struct{}
	proxyMarkHeader string
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func markHeader(prefix string, n uint) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return prefix + string(b)
}

func New(config *Config, jwtAuthorizer authorizer.JWTAuthorizer, myDomains []string) (*Instance, error) {
	if config == nil {
		zap.L().Warn("Not starting proxy - no configuration")
		return nil, nil
	}

	domains := make(map[string]struct{})
	for _, domain := range myDomains {
		domains[domain] = struct{}{}
	}

	markHeaderLength := config.MarkHeaderRandomLength
	if markHeaderLength == 0 {
		markHeaderLength = 8
	}

	return &Instance{
		authorizer:      authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		config:          config,
		users:           newUserStorage(config.ConnLimit),
		myDomains:       domains,
		proxyMarkHeader: markHeader(config.MarkHeaderPrefix, markHeaderLength),
	}, nil
}

func (instance *Instance) Shutdown() error {
	if instance.terminated.Swap(true) {
		return xerror.EInternalError("Double proxy shutdown", nil)
	}

	return nil
}

func (instance *Instance) Running() bool {
	return instance.terminated.Load()
}

func (instance *Instance) isMyRequest(r *http.Request) bool {
	hostParts := strings.Split(r.Host, ":")
	_, myDomain := instance.myDomains[hostParts[0]]
	return myDomain
}

func (instance *Instance) cycledProxy(r *http.Request) bool {
	_, cycled := r.Header[instance.proxyMarkHeader]
	return cycled
}

func (instance *Instance) ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (instance.isMyRequest(r) || instance.cycledProxy(r)) && (r.Method != http.MethodConnect) {
			next.ServeHTTP(w, r)
		} else {
			instance.doProxy(w, r)
		}
	})
}
