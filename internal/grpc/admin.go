package grpc

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/vpnhouse/tunnel/internal/iprose"
	"github.com/vpnhouse/tunnel/internal/manager"
	"github.com/vpnhouse/tunnel/proto"
)

type AdminServer struct {
	proto.AdminServiceServer

	// Manager to control over WG sessions stuff
	mgr *manager.Manager

	// Manager to control over IPRose sessions stuff
	ipr *iprose.Instance
}

func newAdminServer(mgr *manager.Manager, ipr *iprose.Instance) *AdminServer {
	return &AdminServer{
		mgr: mgr,
		ipr: ipr,
	}
}

// Event implements proto.AdminServiceServer.
func (s *AdminServer) Event(ctx context.Context, event *proto.EventRequest) (*proto.EventResponse, error) {
	if event.GetAction().AddRestriction != nil {
		err := s.addRestriction(ctx, event.GetAction().AddRestriction)
		if err != nil {
			return nil, err
		}
	}
	if event.GetAction().DeleteRestriction != nil {
		err := s.deleteRestriction(ctx, event.GetAction().DeleteRestriction)
		if err != nil {
			return nil, err
		}
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

func (s *AdminServer) addRestriction(ctx context.Context, add *proto.AddRestriction) error {
	serverTime, err := getServerTime(ctx)
	if err != nil {
		return err
	}

	zap.L().Debug("add restriction action event", zap.Timep("server_time", serverTime))

	// TODO: Add handling restriction event
	return nil
}

func (s *AdminServer) deleteRestriction(ctx context.Context, del *proto.DeleteRestriction) error {
	serverTime, err := getServerTime(ctx)
	if err != nil {
		return err
	}

	zap.L().Debug("delete restriction action event", zap.Timep("server_time", serverTime))

	// TODO: Add handling restriction event
	return nil
}
