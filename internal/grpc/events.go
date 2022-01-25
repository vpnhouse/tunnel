package grpc

import (
	"github.com/Codename-Uranium/common/eventlog"
	"github.com/Codename-Uranium/common/eventlog/helper"
	"github.com/Codename-Uranium/common/proto"
)

type eventServer struct {
	proto.EventLogServiceServer
	events eventlog.EventManager
}

func newEventServer(events eventlog.EventManager) proto.EventLogServiceServer {
	return &eventServer{events: events}
}

func (m *eventServer) FetchEvents(req *proto.FetchEventsRequest, stream proto.EventLogService_FetchEventsServer) error {
	return helper.FetchEvents(m.events, req, stream)
}
