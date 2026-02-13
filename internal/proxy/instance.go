package proxy

import (
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/entitlements"
	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/common-lib-go/keycounter"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xlimits"
	"github.com/vpnhouse/common-lib-go/xproxy"
	"github.com/vpnhouse/common-lib-go/xrand"
	"github.com/vpnhouse/common-lib-go/xstats"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/stats"
)

const ProtoName = "proxy"

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
	statsService   *xstats.Service
	geoipResolver  *geoip.Resolver
	sessionCounter *keycounter.KeyCounter[string]
}

type query struct {
	sessionID      uuid.UUID
	installationID uuid.UUID
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
	instanceID string,
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

	markHeaderName := config.MarkHeaderPrefix + xrand.RandomString(markHeaderLength)
	transport := &transport{
		httpClient: *http.DefaultClient,
	}

	sessionCounter := keycounter.New[string]()

	var statsReporter *xstats.Service
	if statsService != nil {
		var err error
		statsReporter, err = statsService.Register(ProtoName, func() stats.ExtraStats {
			peers := sessionCounter.Count()
			return stats.ExtraStats{
				PeersTotal:  peers,
				PeersActive: peers,
			}
		})
		if err != nil {
			return nil, xerror.EInternalError("Can't register proxy on stat service", err)
		}
	}

	instance := &Instance{
		config: config,
		authorizer: authorizer.WithEntitlement(
			jwtAuthorizer,
			authorizer.Entitlement(entitlements.Proxy),
			authorizer.Restriction(instanceID),
		),
		users:          xlimits.NewBlocker(config.ConnLimit),
		myDomains:      domains,
		markHeaderName: markHeaderName,
		statsService:   statsReporter,
		geoipResolver:  geoipResolver,
		sessionCounter: sessionCounter,
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
	return !instance.terminated.Load()
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

	userID := token.Subject
	instance.sessionCounter.Inc(userID)

	return &query{
		sessionID:      uuid.MustParse(token.Id),
		installationID: uuid.MustParse(token.InstallationId),
		userID:         userID,
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
	instance.sessionCounter.Dec(authInfo.userID)
}

func (q *query) onStatsSessionData(sessionID uuid.UUID, sessionData *xstats.SessionData) {
	sessionData.InstallationID = q.installationID.String()
	sessionData.UserID = q.userID
	sessionData.Country = q.country
}

func (instance *Instance) doReportRx(description any, n uint64) {
	if instance.statsService == nil {
		return
	}

	query := description.(*query)
	instance.statsService.ReportStats(query.sessionID, n, 0, query.onStatsSessionData)
}

func (instance *Instance) doReportTx(description any, n uint64) {
	if instance.statsService == nil {
		return
	}

	query := description.(*query)
	instance.statsService.ReportStats(query.sessionID, 0, n, query.onStatsSessionData)
}

// admin.Handler implementation
func (instance *Instance) KillActiveUserSessions(userId string) {
	// TODO: add implementation
}
