package xcrypto

import (
	"github.com/dgrijalva/jwt-go"
)

const (
	AudienceAuth     = "auth"
	AudienceDiscover = "discover"
	AudienceTunnel   = "tunnel"
)

type StringList []string

func (l StringList) Has(entry string) bool {
	for _, e := range l {
		if e == entry {
			return true
		}
	}

	return false
}

type ClientClaims struct {
	Audience StringList `json:"aud,omitempty"`
	jwt.StandardClaims
}
