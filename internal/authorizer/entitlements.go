package authorizer

import (
	"fmt"

	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

type EntitlementType string

const (
	Any       EntitlementType = ""
	Wireguard EntitlementType = "wireguard"
	IPRose    EntitlementType = "iprose"
	Proxy     EntitlementType = "proxy"
)

type jwtAuthorizerEntitlement struct {
	JWTAuthorizer
	Entitlement EntitlementType
}

var _ JWTAuthorizer = (*jwtAuthorizerEntitlement)(nil)

func WithEntitlement(jwtAuthorizer JWTAuthorizer, entitlement EntitlementType) *jwtAuthorizerEntitlement {
	return &jwtAuthorizerEntitlement{
		JWTAuthorizer: jwtAuthorizer,
		Entitlement:   entitlement,
	}
}

func (d *jwtAuthorizerEntitlement) Authenticate(tokenString string, myAudience string) (*auth.ClientClaims, error) {
	claims, err := d.JWTAuthorizer.Authenticate(tokenString, myAudience)
	if err != nil {
		return nil, err
	}

	if d.Entitlement == Any {
		return claims, nil
	}

	// Note
	// Probably need check entitlement + platform_type
	v, ok := claims.Entitlements[string(d.Entitlement)]
	if !ok || fmt.Sprint(v) != "true" {
		return nil, xerror.ENoLicense(fmt.Sprintf("no entitlement: %s", d.Entitlement))
	}

	return claims, nil
}
