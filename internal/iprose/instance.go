//go:build iprose
// +build iprose

package iprose

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/geoip"
	"github.com/vpnhouse/common-lib-go/stats"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/iprose-go/pkg/server"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"go.uber.org/zap"
)

const (
	DefaultQueueSize      = 1024
	DefaultSessionTimeout = time.Minute * 5
)

type Config struct {
	QueueSize        int           `yaml:"queue_size"`
	PersistentTokens []string      `yaml:"persistent_tokens"`
	SessionTimeout   time.Duration `yaml:"session_timeout"`
	ProxyConnLimit   int           `yaml:"proxy_conn_limit"`
}

var DefaultConfig = Config{
	QueueSize:      DefaultQueueSize,
	SessionTimeout: DefaultSessionTimeout,
}

type Instance struct {
	iprose        *server.IPRoseServer
	authorizer    authorizer.JWTAuthorizer
	config        Config
	geoipResolver *geoip.Resolver
}

func New(
	config Config,
	jwtAuthorizer authorizer.JWTAuthorizer,
	statsService *stats.Service,
	geoipResolver *geoip.Resolver,
) (*Instance, error) {
	zap.L().Info("Starting iprose service",
		zap.Int("trusted tokens", len(config.PersistentTokens)),
		zap.Int("queue size", config.QueueSize),
		zap.Duration("session timeout", config.SessionTimeout))

	instance := &Instance{
		authorizer:    authorizer.WithEntitlement(jwtAuthorizer, authorizer.IPRose),
		config:        config,
		geoipResolver: geoipResolver,
	}
	var err error
	instance.iprose, err = server.New(
		"iprose0",
		"10.123.0.1/16",
		"fc00:123:76::1/96",
		[]string{"0.0.0.0/0"},
		config.QueueSize,
		instance.Authenticate,
		config.SessionTimeout,
		config.ProxyConnLimit != 0,
		config.ProxyConnLimit,
		statsService, // safe to pass nil
	)
	if err != nil {
		zap.L().Error("Can't start iprose service", zap.Error(err))
		return nil, err
	}

	return instance, nil
}

func (instance *Instance) Authenticate(r *http.Request) (*server.UserInfo, error) {
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return nil, xerror.EAuthenticationFailed("no auth token", nil)
	}

	for _, t := range instance.config.PersistentTokens {
		if userToken == t {
			zap.L().Debug("Authenticated with fixed trusted token")
			return &server.UserInfo{}, nil
		}
	}

	claims, err := instance.authorizer.Authenticate(r.Context(), userToken, auth.AudienceTunnel)
	if err != nil {
		return nil, err
	}

	var userID string
	var installationID string
	if claims != nil {
		userID = claims.Subject
		installationID = claims.InstallationId
	} else {
		// must be in form of subject "<project_id>/<auth_method_id>/<external_user_id>"
		userID = strings.Join([]string{uuid.New().String(), uuid.New().String(), uuid.New().String()}, "/")
		// to indicate it's dummy
		installationID = ""
	}

	clientInfo := instance.geoipResolver.GetInfo(r)
	zap.L().Debug("iprose: client info",
		zap.Stringer("request", xhttp.RequestStringer(r)),
		zap.String("country", clientInfo.Country),
	)

	return &server.UserInfo{
		InstallationID: installationID,
		UserID:         userID,
		Country:        clientInfo.Country,
	}, nil
}

func (instance *Instance) RegisterHandlers(r chi.Router) {
	for _, hndlr := range instance.iprose.Handlers() {
		r.HandleFunc(hndlr.Pattern, hndlr.Func)
	}
}

func (instance *Instance) Shutdown() error {
	instance.iprose.Shutdown()
	return nil
}

func (instance *Instance) Running() bool {
	return instance.iprose.Running()
}

// admin.Handler implementation
func (instance *Instance) KillActiveUserSessions(userId string) {
	// TODO: add implementation
}
