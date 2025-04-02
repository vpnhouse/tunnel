package proxy

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/common-lib-go/xlimits"
	"github.com/vpnhouse/tunnel/internal/authorizer"
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
	users           *xlimits.Blocker
	myDomains       map[string]struct{}
	proxyMarkHeader string
	terminated      atomic.Bool
	statsService    *stats.Service
	geoipResolver   *geoip.Resolver
}

type authInfo struct {
	InstallationID string
	UserID         string
	Country        string
}

func New(
	config *Config,
	jwtAuthorizer authorizer.JWTAuthorizer,
	myDomains []string,
	statsService *stats.Service,
	geoipResolver *geoip.Resolver,
) (*Instance, error) {
	if config == nil {
		return nil, xerror.EInternalError("No configuration", nil)
	}

	domains := make(map[string]struct{})
	for _, domain := range myDomains {
		domains[domain] = struct{}{}
	}

	markHeaderLength := config.MarkHeaderRandomLength
	if markHeaderLength == 0 {
		markHeaderLength = 8
	}

	instance := &Instance{
		config:          config,
		authorizer:      authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		users:           xlimits.NewBlocker(config.ConnLimit),
		myDomains:       domains,
		proxyMarkHeader: config.MarkHeaderPrefix + randomString(markHeaderLength),
		statsService:    statsService,
		geoipResolver:   geoipResolver,
	}

	return instance, nil
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

func (instance *Instance) doAuth(r *http.Request) (*authInfo, error) {
	userToken, ok := extractProxyAuthToken(r)
	if !ok {
		return nil, xerror.WAuthenticationFailed("proxy", "no auth token", nil)
	}

	token, err := instance.authorizer.Authenticate(r.Context(), userToken, auth.AudienceTunnel)
	if err != nil {
		return nil, err
	}

	clientInfo := instance.geoipResolver.GetInfo(r)
	return &authInfo{
		InstallationID: token.InstallationId,
		UserID:         token.Subject,
		Country:        clientInfo.Country,
	}, nil
}

func (instance *Instance) doProxy(w http.ResponseWriter, r *http.Request) {
	authInfo, err := instance.doAuth(r)
	if err != nil {
		w.Header()["Proxy-Authenticate"] = []string{"Basic realm=\"proxy\""}
		w.WriteHeader(http.StatusProxyAuthRequired)
		w.Write([]byte("Proxy authentication required"))
		return
	}

	user, err := instance.users.Acquire(r.Context(), authInfo.UserID)
	if err != nil {
		http.Error(w, "Limit exceeded", http.StatusTooManyRequests)
		xhttp.WriteJsonError(w, err)
		return
	}
	defer instance.users.Release(authInfo.UserID, user)

	query := &ProxyQuery{
		auth:          authInfo,
		id:            queryCounter.Add(1),
		proxyInstance: instance,
	}

	if r.Method == "CONNECT" {
		if r.ProtoMajor == 1 {
			query.handleV1Connect(w, r)
			return
		}

		if r.ProtoMajor == 2 {
			query.handleV2Connect(w, r)
			return
		}

		http.Error(w, "Unsupported protocol version", http.StatusHTTPVersionNotSupported)
		return
	} else {
		query.handleProxy(w, r)
	}
}

// admin.Handler implementation
func (instance *Instance) KillActiveUserSessions(userId string) {
	// TODO: add implementation
}
