package types

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

type ActionRuleType string

const (
	ActionRuleRestrict ActionRuleType = "restrict"
)

type ActionRule struct {
	UserId    string         `db:"user_id" json:"user_id,omitempty"`
	Action    ActionRuleType `db:"action" json:"action,omitempty"`
	Expires   *int64         `db:"expires" json:"expires,omitempty"`
	RulesJson string         `db:"rules_json" json:"rules_json,omitempty"`
}

func (s *ActionRule) IsEmpty() bool {
	if s == nil {
		return true
	}
	return s.Action == ""
}

func (s *ActionRule) IsActive(now time.Time) bool {
	if s.IsEmpty() {
		return false
	}
	// No rule no action
	if s == nil {
		return false
	}
	if s.Expires == nil {
		// The rules forever active
		return true
	}

	return now.Before(time.Unix(*s.Expires, 0))
}

func (s *ActionRule) GetRules() map[string]any {
	if s.RulesJson == "" {
		return nil
	}
	var rules map[string]any
	err := json.Unmarshal([]byte(s.RulesJson), &rules)
	if err != nil {
		zap.L().Error(
			"failed to parse action_rules",
			zap.String("action_rule_type", string(s.Action)),
			zap.String("user_id", s.UserId),
			zap.String("rules_jeson", s.RulesJson),
			zap.Error(err),
		)
		return nil
	}
	return rules
}
