package grpc

import (
	"github.com/Codename-Uranium/tunnel/eventlog"
	"github.com/Codename-Uranium/tunnel/eventlog/helper"
	"github.com/Codename-Uranium/tunnel/proto"
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
