package proxy

import (
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xlimits"
	"github.com/vpnhouse/common-lib-go/xproxy"
	"github.com/vpnhouse/common-lib-go/xrand"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/eventlog"
)

type Config struct {
	ConnLimit              int           `yaml:"conn_limit"`
	ConnTimeout            time.Duration `yaml:"conn_timeout"`
	MarkHeaderPrefix       string        `yaml:"mark_header_prefix"`
	MarkHeaderRandomLength uint          `yaml:"mark_header_random_length"`
}

type Instance struct {
	config         *Config
	authorizer     authorizer.JWTAuthorizer
	fetcher        xproxy.Instance
	users          *xlimits.Blocker
	myDomains      map[string]struct{}
	markHeaderName string
	terminated     atomic.Bool
	statsService   *stats.Service
	geoipResolver  *geoip.Resolver
	eventlog       eventlog.EventManager
}

type query struct {
	sessionID      uuid.UUID
	installationID string
	userID         string
	country        string

	releaser   func()
	reporterTx func()
	reporterRx func()
}

type transport struct {
	httpClient http.Client
}

func (transport *transport) Dial(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

func (transport *transport) HttpClient() *http.Client {
	return &transport.httpClient
}

func New(
	config *Config,
	jwtAuthorizer authorizer.JWTAuthorizer,
	myDomains []string,
	eventlog eventlog.EventManager,
	geoipResolver *geoip.Resolver,
) (*Instance, error) {
	if config == nil {
		return nil, xerror.EInternalError("No configuration", nil)
	}

	statsService, err := stats.New(
		runtime.Settings.Statistics.FlushInterval.Value(),
		eventLog,
		"proxy",
	)

	domains := make(map[string]struct{})
	for _, domain := range myDomains {
		domains[domain] = struct{}{}
	}

	markHeaderLength := config.MarkHeaderRandomLength
	if markHeaderLength == 0 {
		markHeaderLength = 8
	}

	markHeaderName := config.MarkHeaderPrefix + xrand.RandomString(markHeaderLength)
	transport := &transport{
		httpClient: *http.DefaultClient,
	}

	instance := &Instance{
		config:         config,
		authorizer:     authorizer.WithEntitlement(jwtAuthorizer, authorizer.Proxy),
		users:          xlimits.NewBlocker(config.ConnLimit),
		myDomains:      domains,
		markHeaderName: markHeaderName,
		statsService:   statsService,
		geoipResolver:  geoipResolver,
	}

	instance.fetcher = xproxy.Instance{
		MarkHeaderName: markHeaderName,
		Transport:      transport,
		AuthCallback: func(r *http.Request) (description any, err error) {
			return instance.doAuth(r)
		},
		ReleaseCallback: func(description any) {
			instance.doRelease(description.(*query))
		},
		StatsReportRx: instance.doReportRx,
		StatsReportTx: instance.doReportTx,
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
	_, cycled := r.Header[instance.markHeaderName]
	return cycled
}

func (instance *Instance) ProxyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (instance.isMyRequest(r) || instance.cycledProxy(r)) && (r.Method != http.MethodConnect) {
			next.ServeHTTP(w, r)
		} else {
			instance.fetcher.ServeHTTP(w, r)
		}
	})
}

func (instance *Instance) doAuth(r *http.Request) (*query, error) {
	userToken, ok := extractProxyAuthToken(r)
	if !ok {
		return nil, xerror.WAuthenticationFailed("proxy", "no auth token", nil)
	}

	token, err := instance.authorizer.Authenticate(r.Context(), userToken, auth.AudienceTunnel)
	if err != nil {
		return nil, err
	}

	clientInfo := instance.geoipResolver.GetInfo(r)
	user, err := instance.users.Acquire(r.Context(), token.Subject)
	if err != nil {
		return nil, err
	}

	return &query{
		sessionID:      uuid.New(),
		installationID: token.InstallationId,
		userID:         token.Subject,
		country:        clientInfo.Country,
		releaser: func() {
			instance.users.Release(token.Subject, user)
		},
		reporterRx: func() {

		},
		reporterTx: func() {

		},
	}, nil
}

func (instance *Instance) doRelease(authInfo *query) {
	authInfo.releaser()
}

func (instance *Instance) doReportRx(description any, n uint64) {
	if instance.statsService == nil {
		return
	}

	query := description.(*query)
	instance.statsService.ReportStats(query.sessionID, n, 0, func(sessionID uuid.UUID, sessionData *stats.SessionData) {
		sessionData.InstallationID = query.installationID
		sessionData.UserID = query.userID
		sessionData.Country = query.country
	})
}

func (instance *Instance) doReportTx(description any, n uint64) {
	if instance.statsService == nil {
		return
	}

	query := description.(*query)
	instance.statsService.ReportStats(query.sessionID, 0, n, func(sessionID uuid.UUID, sessionData *stats.SessionData) {
		sessionData.InstallationID = query.installationID
		sessionData.UserID = query.userID
		sessionData.Country = query.country
	})
}

// admin.Handler implementation
func (instance *Instance) KillActiveUserSessions(userId string) {
	// TODO: add implementation
}
