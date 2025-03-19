package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vpnhouse/common-lib-go/xutils"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
)

var ActionRulesCheckErrors = map[types.ActionRuleType]error{
	types.ActionRuleRestrict: errors.New(string(types.ActionRuleRestrict)),
}

type AddRestrictionRequest struct {
	UserId         string
	InstallationId string
	SessionId      string
	ExpiredTo      int64
}

type DeleteRestrictionRequest struct {
	UserId string
}

func (s *Service) AddRestriction(ctx context.Context, req *AddRestrictionRequest) error {
	var expires *int64
	if req.ExpiredTo > 0 {
		expires = &req.ExpiredTo
	}
	err := s.storage.AddActionRule(ctx, &types.ActionRule{
		UserId:  req.UserId,
		Expires: expires,
		Action:  types.ActionRuleRestrict,
	})
	if err != nil {
		return err
	}

	zap.L().Info("action rule added",
		zap.String("user_id", req.UserId),
		zap.Int64("expires", req.ExpiredTo),
		zap.String("action_type", string(types.ActionRuleRestrict)))

	s.usersToKillSessions.Set(xutils.StringToBytes(req.UserId), nil)

	return nil
}

func (s *Service) DeleteRestriction(ctx context.Context, req *DeleteRestrictionRequest) error {
	now := time.Now().UTC()
	err := s.storage.DeleteActionRules(ctx, req.UserId, []string{string(types.ActionRuleRestrict)}, &now)
	if err != nil {
		return err
	}

	key := []byte(req.UserId + "/" + string(types.ActionRuleRestrict))
	s.actionsCache.Del(key)
	zap.L().Info("action rule deleted",
		zap.String("user_id", req.UserId), zap.String("action_type", string(types.ActionRuleRestrict)))

	s.usersToKillSessions.Del(xutils.StringToBytes(req.UserId))

	return nil
}

func (s *Service) CheckUserByActionRules(ctx context.Context, userId string, serverTime ...*time.Time) error {
	now := time.Now().UTC()
	if len(serverTime) == 1 {
		now = *(serverTime[0])
	}
	for actionType, actionCheckError := range ActionRulesCheckErrors {
		// Can return nil if error or no rule for this user
		key := []byte(userId + "/" + string(actionType))
		var actionRule types.ActionRule
		v, err := s.actionsCache.GetSet(key, func() ([]byte, error) {
			r := s.getActionRule(ctx, userId, actionType, &now)
			if r == nil {
				r = &types.ActionRule{}
			}
			return json.Marshal(r)
		})
		if err != nil {
			return err
		}

		err = json.Unmarshal(v, &actionRule)
		if err != nil {
			return fmt.Errorf("failed to unmarshal action rule")
		}

		if actionRule.IsActive(now) {
			return actionCheckError
		}
	}
	return nil
}

func (s *Service) getActionRule(ctx context.Context, userId string, actionRuleType types.ActionRuleType, now *time.Time) *types.ActionRule {
	if now == nil {
		nowTime := time.Now().UTC()
		now = &nowTime
	}
	actionRules, err := s.storage.FindActionRules(ctx, userId, []string{string(actionRuleType)}, now)
	if err != nil {
		zap.L().Error("failed to get action_rules",
			zap.String("action_rule_type", string(actionRuleType)),
			zap.String("user_id", userId),
			zap.Error(err),
		)
		return nil
	}

	// No any rules
	if len(actionRules) == 0 {
		zap.L().Debug("no action rules found for user",
			zap.String("user_id", userId), zap.String("action", string(actionRuleType)), zap.Int64("timestamp", now.Unix()))
		return nil
	}

	zap.L().Debug("action rules for user",
		zap.String("user_id", userId), zap.String("action", string(actionRuleType)), zap.Int("count", len(actionRules)))

	// Returns most recent one
	return actionRules[0]
}
