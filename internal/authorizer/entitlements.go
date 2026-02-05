package authorizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/entitlements"
	"github.com/vpnhouse/common-lib-go/xerror"

	"go.uber.org/zap"
)

const (
	Any = ""
)

type EntitelmentValidator func(entitelments entitlements.Entitlements) error

type jwtAuthorizerEntitlement struct {
	JWTAuthorizer
	Validators []EntitelmentValidator
}

func Entitlement(entitlement string) EntitelmentValidator {
	return func(entitelments entitlements.Entitlements) error {
		if entitlement == Any {
			return nil
		}
		v, ok := entitelments[entitlement]
		if !ok || fmt.Sprint(v) != "true" {
			zap.L().Debug("entitlements",
				zap.String("entitlement", entitlement),
				zap.Any("entitlements", entitelments),
			)
			return xerror.ENoLicense(fmt.Sprintf("no entitlement: %s", entitlement))
		}
		return nil
	}
}

func Restriction(tunnelId string) EntitelmentValidator {
	return func(entitelments entitlements.Entitlements) error {
		restrictions, _ := entitelments["restrictions"].(string)
		if restrictions != "" {
			zap.L().Debug("restricted",
				zap.String("restrictions", restrictions),
				zap.String("tunnel_id", tunnelId),
				zap.Any("entitlements", entitelments),
			)

			if !strings.Contains(restrictions, tunnelId) {
				return xerror.ENoLicense("restricted client cannot be served by this tunnel server")
			}
		}
		return nil
	}
}

var _ JWTAuthorizer = (*jwtAuthorizerEntitlement)(nil)

func WithEntitlement(jwtAuthorizer JWTAuthorizer, validators ...EntitelmentValidator) *jwtAuthorizerEntitlement {
	return &jwtAuthorizerEntitlement{
		JWTAuthorizer: jwtAuthorizer,
		Validators:    validators,
	}
}

func (d *jwtAuthorizerEntitlement) Authenticate(ctx context.Context, tokenString string, myAudience string) (*auth.ClientClaims, error) {
	claims, err := d.JWTAuthorizer.Authenticate(ctx, tokenString, myAudience)
	if err != nil {
		return nil, err
	}

	for _, validator := range d.Validators {
		err = validator(claims.Entitlements)
		if err != nil {
			return nil, err
		}
	}

	return claims, nil
}
