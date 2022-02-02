package wireguard

import (
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Wireguard struct {
	client  *wgctrl.Client
	config  wgtypes.Config
	link    *wireguardLink
	running bool
}

type wireguardLink struct {
	name string
}

func (wl *wireguardLink) Attrs() *netlink.LinkAttrs {
	return &netlink.LinkAttrs{
		Name: wl.name,
	}
}

func (wl *wireguardLink) Type() string {
	return "wireguard"
}

func New(config Config, privateKey types.WGPrivateKey) (*Wireguard, error) {
	linkAttrs := wireguardLink{name: config.Interface}
	client, err := wgctrl.New()
	if err != nil {
		return nil, xerror.ETunnelError("can't create wireguard controller", err)
	}

	key := privateKey.Unwrap()
	wgConfig := wgtypes.Config{
		PrivateKey: &key,
		ListenPort: &config.ServerPort,
	}

	wg := &Wireguard{
		client: client,
		config: wgConfig,
		link:   &linkAttrs,
	}

	err = netlink.LinkAdd(wg.link)
	if err != nil {
		return nil, xerror.ETunnelError("can't add link", err, zap.Any("iface", wg.link.name))
	}

	defer func() {
		if !wg.running {
			zap.L().Error("removing link due to unsuccessful start", zap.String("iface", wg.link.name))
			_ = netlink.LinkDel(wg.link)
		}
	}()

	addr, err := netlink.ParseAddr(config.Subnet.Unwrap().FirstUsable().String())
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse wireguard subnet", err, zap.String("subnet", string(config.Subnet)))
	}

	err = netlink.AddrAdd(wg.link, addr)
	if err != nil {
		return nil, xerror.ETunnelError("can't add address", err, zap.Any("addr", addr))
	}

	err = wg.client.ConfigureDevice(config.Interface, wg.config)
	if err != nil {
		return nil, xerror.ETunnelError("can't configure wireguard interface", err, zap.Any("config", wg.config))
	}

	err = netlink.LinkSetUp(wg.link)
	if err != nil {
		return nil, xerror.ETunnelError("can't set link up", err, zap.Any("iface", wg.link.name), zap.Stringer("addr", addr))
	}

	wg.running = true
	return wg, nil
}

func (wg *Wireguard) Shutdown() error {
	zap.L().Info("Removing wireguard interface")
	err := netlink.LinkDel(wg.link)
	if err != nil {
		return xerror.ETunnelError("can't remove wireguard interface", err)
	}

	wg.running = false
	return nil
}

func (wg *Wireguard) Running() bool {
	return wg.running
}

// SetPeer sets peer on wireguard interface
// Note: it's caller responsibility to provide fully valid peer
func (wg *Wireguard) SetPeer(info *types.PeerInfo) error {
	zap.L().Debug("Set peer", zap.Any("peer", info))

	config, err := wg.getPeerConfig(info, false)
	if err != nil {
		return err
	}

	err = wg.client.ConfigureDevice(wg.link.name, *config)
	if err != nil {
		return xerror.ETunnelError("can't set peer", err, zap.Any("peer", info), zap.Any("config", config))
	}

	return nil
}

// UnsetPeer removes peer from wireguard interface
// Note: it's caller responsibility to provide fully valid peer
func (wg *Wireguard) UnsetPeer(info *types.PeerInfo) error {
	zap.L().Debug("Unset peer", zap.Any("peer", info))

	config, err := wg.getPeerConfig(info, true)
	if err != nil {
		return err
	}

	err = wg.client.ConfigureDevice(wg.link.name, *config)
	if err != nil {
		return xerror.ETunnelError("can't unset peer", err, zap.Any("peer", info), zap.Any("config", config))
	}

	return nil
}

// GetPeers returns peers configured for the underlying device.
// Map key is a peer's public key string.
func (wg *Wireguard) GetPeers() (map[string]wgtypes.Peer, error) {
	dev, err := wg.client.Device(wg.link.name)
	if err != nil {
		return nil, xerror.ETunnelError("failed to get wireguard device", err, zap.String("iface", wg.link.name))
	}

	peers := make(map[string]wgtypes.Peer, len(dev.Peers))
	for _, p := range dev.Peers {
		peers[p.PublicKey.String()] = p
	}

	return peers, nil
}

func (wg *Wireguard) GetLinkStatistic() (*netlink.LinkStatistics, error) {
	link, err := netlink.LinkByName(wg.link.name)
	if err != nil {
		return nil, xerror.EInternalError("failed to get wg link by name", err, zap.String("iface", wg.link.name))
	}

	return link.Attrs().Statistics, nil
}
