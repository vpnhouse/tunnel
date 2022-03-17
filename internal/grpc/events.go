// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package grpc

import (
	"context"
	"errors"

	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eventServer struct {
	proto.EventLogServiceServer
	events eventlog.EventManager
}

func newEventServer(events eventlog.EventManager) proto.EventLogServiceServer {
	return &eventServer{events: events}
}

func (m *eventServer) FetchEvents(req *proto.FetchEventsRequest, stream proto.EventLogService_FetchEventsServer) error {
	sub, err := m.events.Subscribe(stream.Context(), eventlog.SubscriptionOpts{
		LogID:  req.GetLogID(),
		Offset: req.GetOffset(),
		Labels: req.GetLabels(),
	})
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	types := newEventTypeSet(req.GetEventTypes())
	for {
		select {
		// note that we dont handle the request's context cancellation explicitly,
		//  it will be done by the subscription's internals,
		//  thus we care only about the errors chan.
		case event := <-sub.Events():
			if types.has(event.Type) {
				if err := stream.Send(event.IntoProto()); err != nil {
					zap.L().Warn("failed to send an event",
						zap.Error(err), zap.Any("caller_labels", req.GetLabels()))
				}
			}
		case err := <-sub.Errors():
			if err == nil || errors.Is(err, context.Canceled) {
				// peer gone and closed the subscription itself
				return status.Error(codes.Canceled, err.Error())
			}

			return status.Error(codes.Aborted, err.Error())
		}
	}
}

type eventTypesSet map[uint32]struct{}

func newEventTypeSet(vs []uint32) eventTypesSet {
	m := eventTypesSet{}
	for _, v := range vs {
		m[v] = struct{}{}
	}
	return m
}

func (e eventTypesSet) has(v uint32) bool {
	// empty set mean any events
	if len(e) == 0 {
		return true
	}
	_, ok := e[v]
	return ok
}
