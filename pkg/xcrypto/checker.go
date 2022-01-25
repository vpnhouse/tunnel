package xcrypto

import (
	"crypto/rsa"
	"fmt"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// KeyStoreWrapper wraps any type into its closure func `fn`
// and provides the KeyStore interface.
type KeyStoreWrapper struct {
	Fn func(keyUUID uuid.UUID) (*rsa.PublicKey, error)
}

func (w *KeyStoreWrapper) GetKey(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
	return w.Fn(keyUUID)
}

type KeyStore interface {
	GetKey(keyUUID uuid.UUID) (*rsa.PublicKey, error)
}

type JWTChecker struct {
	keys   KeyStore
	method jwt.SigningMethod
}

func NewJWTChecker(keyKeeper KeyStore) (*JWTChecker, error) {
	method := jwt.GetSigningMethod(jwtSigningMethod)
	if method == nil {
		return nil, xerror.EInvalidArgument("signing method is not supported", nil, zap.String("method", jwtSigningMethod))
	}

	return &JWTChecker{
		keys:   keyKeeper,
		method: method,
	}, nil
}

func (instance *JWTChecker) keyHelper(token *jwt.Token) (interface{}, error) {
	keyIdValue, ok := token.Header["kid"]
	if !ok {
		return nil, xerror.EAuthenticationFailed("invalid token", nil)
	}
	zap.L().Debug("Got key id", zap.Any("keyIdValue", keyIdValue))

	var keyIdStr string
	switch v := keyIdValue.(type) {
	case string:
		keyIdStr = v
	default:
		return nil, xerror.EAuthenticationFailed("invalid token", fmt.Errorf("unsupported key id type"))
	}

	keyId, err := uuid.Parse(keyIdStr)
	if err != nil {
		return nil, xerror.EAuthenticationFailed("invalid token", err)
	}

	key, err := instance.keys.GetKey(keyId)
	if err != nil {
		return nil, err
	}

	return key, nil

}

func (instance *JWTChecker) Parse(tokenString string, claims jwt.Claims) error {
	// Check if we have private key

	token, err := jwt.ParseWithClaims(tokenString, claims, instance.keyHelper)
	if err != nil {
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	if !token.Valid {
		return xerror.EAuthenticationFailed("invalid token", nil)
	}

	method := token.Method.Alg()
	if method != instance.method.Alg() {
		return xerror.EAuthenticationFailed(
			"invalid token",
			fmt.Errorf("invalid signing method"),
			zap.String("method", method),
			zap.Any("token", token),
		)
	}

	return nil
}
