package authorizer

import (
	"context"
	"fmt"

	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/xerror"

	"go.uber.org/zap"
)

const (
	Any = ""
)

type jwtAuthorizerEntitlement struct {
	JWTAuthorizer
	Entitlement string
}

var _ JWTAuthorizer = (*jwtAuthorizerEntitlement)(nil)

func WithEntitlement(jwtAuthorizer JWTAuthorizer, entitlement string) *jwtAuthorizerEntitlement {
	return &jwtAuthorizerEntitlement{
		JWTAuthorizer: jwtAuthorizer,
		Entitlement:   entitlement,
	}
}

func (d *jwtAuthorizerEntitlement) Authenticate(ctx context.Context, tokenString string, myAudience string) (*auth.ClientClaims, error) {
	claims, err := d.JWTAuthorizer.Authenticate(ctx, tokenString, myAudience)
	if err != nil {
		return nil, err
	}

	if d.Entitlement == Any {
		return claims, nil
	}

	//TODO: Also check platform_type
	v, ok := claims.Entitlements[d.Entitlement]
	if !ok || fmt.Sprint(v) != "true" {
		zap.L().Debug("entitlements",
			zap.String("entitlement", d.Entitlement),
			zap.Any("entitlements", claims.Entitlements),
		)
		return nil, xerror.ENoLicense(fmt.Sprintf("no entitlement: %s", d.Entitlement))
	}

	return claims, nil
}
