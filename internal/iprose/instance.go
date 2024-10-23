//go:build iprose
// +build iprose

package iprose

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vpnhouse/iprose-go/pkg/server"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
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
	iprose     *server.IPRoseServer
	authorizer authorizer.JWTAuthorizer
	config     Config
}

func New(config Config, jwtAuthorizer authorizer.JWTAuthorizer) (*Instance, error) {
	zap.L().Info("Starting iprose service",
		zap.Int("trusted tokens", len(config.PersistentTokens)),
		zap.Int("queue size", config.QueueSize),
		zap.Duration("session timeout", config.SessionTimeout))

	instance := &Instance{
		authorizer: authorizer.WithEntitlement(jwtAuthorizer, authorizer.IPRose),
		config:     config,
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
	)
	if err != nil {
		zap.L().Error("Can't start iprose service", zap.Error(err))
		return nil, err
	}
	return instance, nil
}

func (instance *Instance) authenticate(r *http.Request) (string, error) {
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return "", xerror.EAuthenticationFailed("no auth token", nil)
	}

	for _, t := range instance.config.PersistentTokens {
		if userToken == t {
			zap.L().Debug("Authenticated with fixed trusted token")
			return "", nil
		}
	}

	claims, err := instance.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return "", err
	}

	var userId string
	if claims != nil {
		userId = claims.UserId
	} else {
		userId = uuid.New().String()
	}

	return userId, nil
}

func (instance *Instance) Authenticate(r *http.Request) (error, int, []byte, string) {
	userId, err := instance.authenticate(r)
	if err != nil {
		code, body := xerror.ErrorToHttpResponse(err)
		return err, code, body, ""
	}

	return nil, 0, nil, userId
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
