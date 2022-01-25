package xcrypto

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const jwtSigningMethod = "RS256"

type JWTMaster struct {
	private   *rsa.PrivateKey
	privateId *uuid.UUID
	public    *rsa.PublicKey
	method    jwt.SigningMethod
}

func NewJWTMaster(private *rsa.PrivateKey, privateId *uuid.UUID) (*JWTMaster, error) {
	// Generate new private key if it's not given by caller
	if private == nil {
		if privateId != nil {
			return nil, xerror.EInternalError("privateId must be nil when private is nil", nil)
		}

		vPrivateId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		privateId = &vPrivateId

		zap.L().Info("generating keys for JWT")
		private, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, xerror.EInternalError("can't generate JWT key pair", err)
		}
	} else {
		if privateId == nil {
			return nil, xerror.EInvalidArgument("privateId must be set when private is set", nil)
		}
	}

	public := &private.PublicKey
	method := jwt.GetSigningMethod(jwtSigningMethod)
	if method == nil {
		return nil, xerror.EInvalidArgument("signing method is not supported", nil, zap.String("method", jwtSigningMethod))
	}

	return &JWTMaster{
		private:   private,
		privateId: privateId,
		public:    public,
		method:    method,
	}, nil
}

func (instance *JWTMaster) Token(claims jwt.Claims) (*string, error) {
	// Check if we have private key
	if instance.private == nil {
		zap.L().Fatal("can't produce token, private key is not set")
	}

	// Create token
	token := jwt.NewWithClaims(instance.method, claims)
	token.Header["kid"] = instance.privateId

	// Sign token
	signedToken, err := token.SignedString(instance.private)
	if err != nil {
		zap.L().Error("Can't sign auth token", zap.Error(err))
		return nil, xerror.EInternalError("can't sign token", err)
	}

	return &signedToken, nil
}

func (instance *JWTMaster) Parse(tokenString string, claims jwt.Claims) error {
	// Check if we have private key
	if instance.public == nil {
		zap.L().Fatal("can't verify token, public key is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return instance.public, nil
	})

	if err != nil || token == nil {
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	if !token.Valid {
		return xerror.EAuthenticationFailed("invalid token", nil)
	}

	method := token.Method.Alg()
	if method != instance.method.Alg() {
		zap.L().Error("Invalid signing method", zap.String("method", method), zap.Any("token", token))
		return xerror.EAuthenticationFailed("invalid token", err)
	}

	return nil
}
