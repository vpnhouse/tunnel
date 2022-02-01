package types

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// let the compiler check types for us
type (
	WGPrivateKey wgtypes.Key
	WGPublicKey  wgtypes.Key
)

func (p WGPrivateKey) Public() WGPublicKey {
	pub := (wgtypes.Key)(p).PublicKey()
	return (WGPublicKey)(pub)
}

func (p WGPrivateKey) Unwrap() wgtypes.Key {
	return (wgtypes.Key)(p)
}

func (p WGPublicKey) Unwrap() wgtypes.Key {
	return (wgtypes.Key)(p)
}
