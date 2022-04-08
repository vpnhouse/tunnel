// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package wireguard

import (
	"fmt"
	"net"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/validator"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Interface  string           `yaml:"interface" valid:"alphanum,required"`
	ServerIPv4 string           `yaml:"server_ipv4" valid:"ipv4"`
	Keepalive  int              `yaml:"keepalive" valid:"natural,required"`
	Subnet     validator.Subnet `yaml:"subnet" valid:"subnet,required"`
	DNS        []string         `yaml:"dns" valid:"ipv4list"`

	// Listen port for wireguard connections.
	ListenPort int `yaml:"server_port" valid:"port,required"`
	// NAT'ed port to access the Listen one, this one announced to the client
	// as part of its configuration. If not specified - `ListenPort` is used.
	// e.g container starts with the -p 3333:3000 option, 3000 here is ListenPort value,
	// so NATedPort must be set to `3333` to push the valid configuration to the client.
	NATedPort int `yaml:"nated_port,omitempty" valid:"port"`

	// PrivateKey of WireGuard, serialized to the string.
	// Generated automatically on the startup.
	PrivateKey string `yaml:"private_key"`

	// parsed version of the field above
	privateKey types.WGPrivateKey
}

func (c *Config) OnLoad() error {
	k, err := wgtypes.ParseKey(c.PrivateKey)
	if err != nil {
		return xerror.EInternalError("failed to parse wireguard's private key", err)
	}

	c.privateKey = (types.WGPrivateKey)(k)
	return nil
}

// ClientPort  returns the port to announce to a client.
// See Config.NATedPort for details.
func (c Config) ClientPort() int {
	if c.NATedPort > 0 {
		return c.NATedPort
	}
	return c.ListenPort
}

// ServerAddr returns IPAddr/mask to use as a wireguard interface address.
func (c Config) ServerAddr() string {
	a := c.Subnet.Unwrap()
	ones, _ := a.Mask().Size()
	return fmt.Sprintf("%s/%d", a.FirstUsable().String(), ones)
}

func (c Config) GetPrivateKey() types.WGPrivateKey {
	return c.privateKey
}

func DefaultConfig() Config {
	privKey, _ := wgtypes.GeneratePrivateKey()
	return Config{
		Interface:  "uwg0",
		ServerIPv4: "",
		ListenPort: 3000,
		Keepalive:  60,
		Subnet:     "10.235.0.0/16",
		DNS:        []string{"8.8.8.8", "8.8.4.4"},

		PrivateKey: privKey.String(),
		privateKey: (types.WGPrivateKey)(privKey),
	}
}

// getPeerConfig generates wireguard configuration for a peer.
// Note: it's caller responsibility to provide fully valid peer
func (wg *Wireguard) getPeerConfig(info types.PeerInfo, remove bool) (*wgtypes.Config, error) {
	key, err := wgtypes.ParseKey(*info.WireguardPublicKey)
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse client public key", err, zap.String("key", *info.WireguardPublicKey))
	}

	ipv4net := net.IPNet{
		IP:   info.Ipv4.IP,
		Mask: net.CIDRMask(32, 32),
	}

	peer := wgtypes.PeerConfig{
		PublicKey:  key,
		Remove:     remove,
		AllowedIPs: []net.IPNet{ipv4net},
	}

	config := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peer},
	}

	return &config, nil
}
