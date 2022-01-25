package authorizer

import (
	"fmt"
	"strings"
	"time"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type FirebaseClaims struct {
	ProviderId string `json:"provider_id,omitempty"`
	AuthTime   int64  `json:"auth_time"`
	UserId     string `json:"user_id"`
	jwt.StandardClaims
}

const (
	defaultFirebaseRefreshInterval = 10 * time.Second
)

type FirebaseAuthorizer struct {
	projectIDs map[string]bool
	keyKeeper  *FirebaseKeyKeeper
	running    bool
}

func NewFirebaseAuthorizer(projectIDs []string, keyKeeper *FirebaseKeyKeeper) (*FirebaseAuthorizer, error) {
	projectIDsMap := make(map[string]bool)
	for _, id := range projectIDs {
		projectIDsMap[id] = true
	}

	return &FirebaseAuthorizer{
		projectIDs: projectIDsMap,
		keyKeeper:  keyKeeper,
		running:    true,
	}, nil
}

func (a *FirebaseAuthorizer) Shutdown() error {
	a.running = false
	return nil
}

func (a *FirebaseAuthorizer) Running() bool {
	return a.running
}

func (a *FirebaseAuthorizer) Authenticate(authInfo string) (string, int64, error) {
	if !a.running {
		return "", 0, xerror.EInternalError("Authenticate attempt on stopped firebase authorizer", nil)
	}

	claims := FirebaseClaims{}
	err := a.parse(authInfo, &claims)
	if err != nil {
		return "", 0, err
	}

	if !a.hasProject(claims.Audience) {
		return "", 0, xerror.EAuthenticationFailed("Project is not known", nil, zap.String("Audience", claims.Audience))
	}

	return claims.UserId, claims.ExpiresAt, nil
}

func (a *FirebaseAuthorizer) hasProject(id string) bool {
	_, ok := a.projectIDs[id]
	return ok
}

func (a *FirebaseAuthorizer) parse(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, a.keyKeeper.KeyHelper)
	if err != nil {
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	if !token.Valid {
		return xerror.EAuthenticationFailed("invalid token", nil)
	}

	method := token.Method.Alg()
	if strings.ToUpper(method[:2]) != "RS" {
		return xerror.EAuthenticationFailed(
			"invalid token",
			fmt.Errorf("invalid signing method"),
			zap.String("method", method),
		)
	}

	return nil
}
