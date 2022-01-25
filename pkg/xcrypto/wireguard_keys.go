package xcrypto

import (
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func WgGeneratePrivateKey() (*wgtypes.Key, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, xerror.ETunnelError("can't generate private key", err)
	}

	return &key, nil
}
