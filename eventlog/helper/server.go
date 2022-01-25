package helper

import (
	"context"
	"errors"

	"github.com/Codename-Uranium/tunnel/eventlog"
	"github.com/Codename-Uranium/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

// FetchEvents is a gRPC server helper that subscribes caller to a log and streams events back
func FetchEvents(em eventlog.EventSubscriber, req *proto.FetchEventsRequest, stream proto.EventLogService_FetchEventsServer) error {
	sub, err := em.Subscribe(stream.Context(), eventlog.SubscriptionOpts{
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
