package wireguard

import (
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Wireguard struct {
	running bool
}

func New(config Config, privateKey wgtypes.Key) (*Wireguard, error) {
	return &Wireguard{running: true}, nil
}

func (w *Wireguard) Shutdown() error {
	w.running = false
	return nil
}

func (w *Wireguard) Running() bool { return w.running }

func (*Wireguard) SetPeer(info *types.PeerInfo) error {
	zap.L().Debug("wg: set peer")
	return nil
}

func (*Wireguard) UnsetPeer(info *types.PeerInfo) error {
	zap.L().Debug("wg: unset peer")
	return nil
}

func (*Wireguard) GetPeers() (map[string]wgtypes.Peer, error) {
	zap.L().Debug("wg: get peers")
	return map[string]wgtypes.Peer{}, nil
}

func (*Wireguard) GetLinkStatistic() (*netlink.LinkStatistics, error) {
	zap.L().Debug("wg: get link stats")
	return &netlink.LinkStatistics{
		RxPackets: 111,
		TxPackets: 222,
		RxBytes:   333,
		TxBytes:   444,
		RxErrors:  42,
		TxErrors:  91,
	}, nil
}
