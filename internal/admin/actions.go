package admin

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type AddRestrictionRequest struct {
	ServerTime     time.Time
	UserId         string
	InstallationId string
	SessionId      string
}

type DeleteRestrictionRequest struct {
	ServerTime     time.Time
	UserId         string
	InstallationId string
	SessionId      string
}

func (s *Service) AddRestriction(ctx context.Context, req *AddRestrictionRequest) error {
	zap.L().Debug("add restriction", zap.Time("server_time", req.ServerTime))

	// TODO: Add handling add restriction event
	return nil
}

func (s *Service) DeleteRestriction(ctx context.Context, req *DeleteRestrictionRequest) error {
	zap.L().Debug("delete restriction", zap.Time("server_time", req.ServerTime))

	// TODO: Add handling delete restriction event
	return nil
}
