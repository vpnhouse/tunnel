package admin

import (
	"context"
	"time"

	"github.com/vpnhouse/common-lib-go/xtime"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
)

type ActionRuleType string

const (
	Restrict ActionRuleType = "restrict"
)

type AddRestrictionRequest struct {
	ServerTime     time.Time
	UserId         string
	InstallationId string
	SessionId      string
	ExpiredTo      int64
}

type DeleteRestrictionRequest struct {
	ServerTime     time.Time
	UserId         string
	InstallationId string
	SessionId      string
}

func (s *Service) AddRestriction(ctx context.Context, req *AddRestrictionRequest) error {
	zap.L().Debug("add restriction", zap.Time("server_time", req.ServerTime))

	err := s.storage.AddActionRule(&types.ActionRule{
		UserId:  req.UserId,
		Expires: &xtime.Time{Time: time.Unix(req.ExpiredTo, 0)},
		Action:  string(Restrict),
	})

	if err != nil {
		return err
	}

	// TODO: Add handling add restriction event
	return nil
}

func (s *Service) DeleteRestriction(ctx context.Context, req *DeleteRestrictionRequest) error {
	zap.L().Debug("delete restriction", zap.Time("server_time", req.ServerTime))

	err := s.storage.DeleteActionRule(&types.ActionRule{
		UserId: req.UserId,
		Action: string(Restrict),
	})

	if err != nil {
		return err
	}
	// TODO: Add handling delete restriction event
	return nil
}
