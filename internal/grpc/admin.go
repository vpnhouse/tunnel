// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package grpc

import (
	"context"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/vpnhouse/tunnel/internal/admin"
	"github.com/vpnhouse/tunnel/proto"
)

type AdminServer struct {
	proto.AdminServiceServer
	adminService *admin.Service
}

// Event implements proto.AdminServiceServer.
func (s *AdminServer) Event(ctx context.Context, event *proto.EventRequest) (*proto.EventResponse, error) {
	serverTime, err := getServerTime(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "event has no X-Server-Time supplied")
	}

	// Event describes one of the dedicated action
	if event.GetAction().AddRestriction != nil {
		err := s.adminService.AddRestriction(ctx, &admin.AddRestrictionRequest{
			ServerTime:     *serverTime,
			UserId:         event.GetAction().GetAddRestriction().GetUserId(),
			InstallationId: event.GetAction().GetAddRestriction().GetInstallationId(),
			SessionId:      event.GetAction().GetAddRestriction().GetSessionId(),
		})
		if err != nil {
			return nil, err
		}
	} else if event.GetAction().DeleteRestriction != nil {
		err := s.adminService.DeleteRestriction(ctx, &admin.DeleteRestrictionRequest{
			ServerTime:     *serverTime,
			UserId:         event.GetAction().GetAddRestriction().GetUserId(),
			InstallationId: event.GetAction().GetAddRestriction().GetInstallationId(),
			SessionId:      event.GetAction().GetAddRestriction().GetSessionId(),
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "event has no action supplied")
	}
	return &proto.EventResponse{}, nil
}

func getServerTime(ctx context.Context) (*time.Time, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.DataLoss, "failed to get metadata")
	}
	serverTimeVals := md.Get("X-Server-Time")
	if len(serverTimeVals) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'X-Server-Time' header")
	}

	serverTimeSeconds, err := strconv.ParseInt(strings.TrimSpace(serverTimeVals[0]), 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "'X-Server-Time' is not valid")
	}

	serverTime := time.Unix(serverTimeSeconds, 0)

	return &serverTime, nil
}
