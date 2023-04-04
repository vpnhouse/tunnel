package eventlog

import (
	"github.com/vpnhouse/tunnel/proto"
)

type Event struct {
	Timestamp int64
	EventType proto.EventType
	PeerInfo  *proto.PeerInfo
	Error     error
}
