package wireguard

import (
	"net"

	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/validator"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Interface  string `yaml:"interface" valid:"alphanum,required"`
	ServerIPv4 string `yaml:"server_ipv4" valid:"ipv4"`
	ServerPort int    `yaml:"server_port" valid:"port,required"`
	Keepalive  int    `yaml:"keepalive" valid:"natural,required"`
	// FIXME(nikonov): it's not a subnet, it is ip/mask,
	//  where IP is a server IP and a mask represents
	//  the address range for the WG clients.
	// Subnet string   `yaml:"subnet" valid:"cidr"`
	Subnet validator.Subnet `yaml:"subnet" valid:"subnet,required"`
	DNS    []string         `yaml:"dns" valid:"ipv4list"`
}

// getPeerConfig generates wireguard configuration for a peer.
// Note: it's caller responsibility to provide fully valid peer
func (wg *Wireguard) getPeerConfig(info types.PeerInfo, remove bool) (*wgtypes.Config, error) {
	if *info.Type != types.TunnelWireguard {
		return nil, xerror.ETunnelError("can't configure non-wireguard peer in wireguard module", nil, zap.Int("type", *info.Type))
	}

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
