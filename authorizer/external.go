package authorizer

import (
	"github.com/Codename-Uranium/tunnel/pkg/control"
)

type ExternalAuthorizer interface {
	control.ServiceController
	Authenticate(authInfo string) (userId string, expire int64, err error)
}
