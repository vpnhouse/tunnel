// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package wireguard

import (
	"github.com/comradevpn/tunnel/internal/types"
	"github.com/comradevpn/tunnel/pkg/xerror"
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

func New(config Config) (*Wireguard, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, xerror.ETunnelError("can't create wireguard controller", err)
	}

	key := config.GetPrivateKey().Unwrap()
	wgConfig := wgtypes.Config{
		PrivateKey: &key,
		ListenPort: &config.ListenPort,
	}

	linkAttrs := wireguardLink{name: config.Interface}
	wg := &Wireguard{
		client: client,
		config: wgConfig,
		link:   &linkAttrs,
	}

	if err := netlink.LinkAdd(wg.link); err != nil {
		// TODO(nikonov): maybe try to takeover the existing interface?
		//  actual for the host-mode only, in docker we'll always have an empty "machine" on restart.
		return nil, xerror.ETunnelError("can't add link", err, zap.Any("iface", wg.link.name))
	}

	defer func() {
		if !wg.running {
			zap.L().Error("removing link due to unsuccessful start", zap.String("iface", wg.link.name))
			_ = netlink.LinkDel(wg.link)
		}
	}()

	addr, err := netlink.ParseAddr(config.ServerAddr())
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse wireguard subnet", err, zap.String("subnet", string(config.Subnet)))
	}

	if err := netlink.AddrAdd(wg.link, addr); err != nil {
		return nil, xerror.ETunnelError("can't add address", err, zap.Any("addr", addr))
	}

	if err := wg.client.ConfigureDevice(config.Interface, wg.config); err != nil {
		return nil, xerror.ETunnelError("can't configure wireguard interface", err, zap.Any("config", wg.config))
	}

	if err := netlink.LinkSetUp(wg.link); err != nil {
		return nil, xerror.ETunnelError("can't set link up", err, zap.Any("iface", wg.link.name), zap.Stringer("addr", addr))
	}

	wg.running = true
	return wg, nil
}

func (wg *Wireguard) Shutdown() error {
	zap.L().Info("removing wireguard interface")
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
func (wg *Wireguard) SetPeer(info types.PeerInfo) error {
	zap.L().Debug("set peer", zap.Any("peer", info))

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
func (wg *Wireguard) UnsetPeer(info types.PeerInfo) error {
	zap.L().Debug("unset peer", zap.Any("peer", info))

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
// Map's key is a peer's public key string.
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
