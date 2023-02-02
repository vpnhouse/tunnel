package eventlog

import (
	"github.com/vpnhouse/tunnel/proto"
)

type Event struct {
	*proto.PeerInfo
	Err error
}
