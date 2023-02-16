// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/federation_keys"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type eventServer struct {
	proto.EventLogServiceServer
	events    eventlog.EventManager
	keystore  federation_keys.Keystore
	tunnelKey string
	storage   *storage.Storage
}

func newEventServer(events eventlog.EventManager, keystore federation_keys.Keystore, tunnelKey string, storage *storage.Storage) proto.EventLogServiceServer {
	return &eventServer{
		events:    events,
		keystore:  keystore,
		tunnelKey: tunnelKey,
		storage:   storage,
	}
}

func (m *eventServer) FetchEvents(req *proto.FetchEventsRequest, stream proto.EventLogService_FetchEventsServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "failed to get metadata")
	}

	authSecret := md.Get(federationAuthHeader)
	if len(authSecret) == 0 || authSecret[0] == "" {
		return status.Errorf(codes.Unauthenticated, "auth secret is empty or not supplied")
	}

	subscriberId, ok := m.keystore.Authorize(authSecret[0])
	if !ok {
		return status.Errorf(codes.Unauthenticated, "auth secret is not valid")
	}

	eventlogPosition := eventlog.EventlogPosition{
		LogID:  req.GetStartPosition().GetLogId(),
		Offset: req.GetStartPosition().GetOffset(),
	}

	if req.StartPosition == nil {
		eventlogSubscriber, err := m.storage.GetEventlogsSubscriber(subscriberId)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			zap.L().Error("failed to get eventlogs subscriber", zap.String("subscriber_id", subscriberId), zap.Error(err))
		}
		if eventlogSubscriber != nil {
			eventlogPosition.LogID = eventlogSubscriber.LogID
			eventlogPosition.Offset = eventlogSubscriber.Offset
		}
	}

	zap.L().Debug("start reading eventlogs",
		zap.String("subscriber_id", subscriberId),
		zap.String("log_id", eventlogPosition.LogID),
		zap.Int64("offset", eventlogPosition.Offset),
	)

	sub, err := m.events.Subscribe(stream.Context(), subscriberId, eventlog.WithPosition(eventlogPosition))
	if err != nil && errors.Is(err, eventlog.ErrNotFound) {
		// Return not found error in case caller supply the position
		if req.StartPosition != nil {
			zap.L().Error("failed to detect start eventlogs position", zap.Error(err))
			return status.Error(codes.NotFound, err.Error())
		}

		// Otherwise try to subscribe to the active log
		sub, err = m.events.Subscribe(stream.Context(), subscriberId, eventlog.WithActiveLog())
	}

	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	header := metadata.New(map[string]string{tunnelAuthHeader: m.tunnelKey})
	err = grpc.SendHeader(stream.Context(), header)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
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
						zap.Error(err), zap.String("subscriber_id", subscriberId))
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

func (m *eventServer) EventFetched(stream proto.EventLogService_EventFetchedServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "failed to get metadata")
	}

	authSecret := md.Get(federationAuthHeader)
	if len(authSecret) == 0 || authSecret[0] == "" {
		return status.Errorf(codes.Unauthenticated, "auth secret is empty or not supplied")
	}

	subscriberId, ok := m.keystore.Authorize(authSecret[0])
	if !ok {
		return status.Errorf(codes.Unauthenticated, "auth secret is not valid")
	}

	header := metadata.New(map[string]string{tunnelAuthHeader: m.tunnelKey})
	err := grpc.SendHeader(stream.Context(), header)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for {
		sub, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				zap.L().Debug("event fetched notifications closed", zap.String("subscriber_id", subscriberId))
			} else {
				zap.L().Error("failed to recieve event fetched notification", zap.String("subscriber_id", subscriberId))
			}
			break
		}
		err = m.storage.PutEventlogsSubscriber(&types.EventlogSubscriber{
			SubscriberID: subscriberId,
			LogID:        sub.GetPosition().GetLogId(),
			Offset:       sub.GetPosition().GetOffset(),
		})
		if err != nil {
			zap.L().Error("failed to store eventlogs subscriber", zap.Any("subscriber", sub), zap.Error(err))
		}
	}

	return nil
}

type eventTypesSet map[eventlog.EventType]struct{}

func newEventTypeSet(vs []proto.EventType) eventTypesSet {
	m := make(eventTypesSet, len(vs))
	for _, v := range vs {
		m[eventlog.EventType(v)] = struct{}{}
	}
	return m
}

func (e eventTypesSet) has(v eventlog.EventType) bool {
	// empty set mean any events
	if len(e) == 0 {
		return true
	}
	_, ok := e[v]
	return ok
}
