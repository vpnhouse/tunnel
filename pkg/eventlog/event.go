package eventlog

import (
	"encoding/json"

	"github.com/vpnhouse/tunnel/proto"
)

type Event struct {
	*proto.PeerInfo
	Err error
}

func ToEvent(resp proto.FetchEventsResponse) *Event {
	if resp.Data == nil {
		return &Event{}
	}

	var peerInfo proto.PeerInfo
	err := json.Unmarshal(resp.Data, &peerInfo)
	if err != nil {
		return &Event{Err: err}
	}

	return &Event{PeerInfo: &peerInfo}
}
