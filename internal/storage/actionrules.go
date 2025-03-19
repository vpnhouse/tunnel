// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/xerror"
	"go.uber.org/zap"

	"github.com/vpnhouse/tunnel/internal/types"
)

func (storage *Storage) AddActionRule(ctx context.Context, action *types.ActionRule) error {
	query := `
		INSERT INTO
			action_rules(user_id, action, expires, rules_json)
		VALUES(:user_id, :action, :expires, :rules_json)
	`
	_, err := storage.db.NamedExecContext(ctx, query, action)
	if err != nil {
		return xerror.EStorageError("can't insert action_rule", err, zap.Any("action", action))
	}
	return nil
}

func (storage *Storage) DeleteActionRules(
	ctx context.Context,
	userId string,
	actionRuleTypes []string,
	now *time.Time,
) error {
	query := `
		DELETE FROM
			action_rules
		WHERE
		   user_id = :user_id
		   AND action IN (:actions)
	`

	if now == nil {
		nowTime := time.Now().UTC()
		now = &nowTime
	}

	params := map[string]any{
		"user_id": userId,
		"actions": actionRuleTypes,
	}

	query, args, err := sqlx.Named(query, params)
	if err != nil {
		return xerror.EStorageError(
			"can't build delete action_rules query",
			err,
			zap.Timep("now", now),
			zap.String("user_id", userId),
		)
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return xerror.EStorageError(
			"can't build delete action_rules query",
			err,
			zap.Time("now", *now),
			zap.String("user_id", userId),
		)
	}
	query = storage.db.Rebind(query)

	_, err = storage.db.ExecContext(ctx, query, args...)
	if err != nil {
		return xerror.EStorageError(
			"can't delete action_rules",
			err,
			zap.String("user_id", userId),
			zap.Any("action_rule_types", actionRuleTypes),
		)
	}
	return nil
}

func (storage *Storage) FindActionRules(
	ctx context.Context,
	userId string,
	actionRuleTypes []string,
	now *time.Time,
) ([]*types.ActionRule, error) {
	query := `
		SELECT 
			user_id,
			action,
			expires,
			rules_json
		FROM 
			action_rules
		WHERE
		    user_id = :user_id
			AND action IN (:actions)
			AND coalesce(expires, :now) >= :now
		ORDER BY expires DESC
	`

	if now == nil {
		nowTime := time.Now().UTC()
		now = &nowTime
	}

	params := map[string]any{
		"now":     now.Unix(),
		"user_id": userId,
		"actions": actionRuleTypes,
	}

	query, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, xerror.EStorageError(
			"can't build action_rules query",
			err,
			zap.Timep("now", now),
			zap.String("user_id", userId),
		)
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, xerror.EStorageError(
			"can't build action_rules query",
			err,
			zap.Timep("now", now),
			zap.String("user_id", userId),
		)
	}
	query = storage.db.Rebind(query)

	rows, err := storage.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, xerror.EStorageError(
			"can't read action_rules",
			err,
			zap.Timep("now", now),
			zap.String("user_id", userId),
		)
	}
	defer rows.Close()

	var actions []*types.ActionRule
	for rows.Next() {
		var act types.ActionRule
		err = rows.StructScan(&act)
		if err != nil {
			zap.L().Error("can't scan action_rule",
				zap.Error(err),
				zap.String("user_id", userId),
				zap.Strings("action_rule_types", actionRuleTypes),
			)
			continue
		}
		actions = append(actions, &act)
	}
	return actions, nil
}

func (storage *Storage) CleanupExpiredActionRules(ctx context.Context) int64 {
	query := `
		DELETE FROM
			action_rules
		WHERE
			coalesce(expires, :now) < :now
	`

	now := time.Now().UTC()

	args := map[string]any{
		"now": now.Unix(),
	}

	res, err := storage.db.NamedExecContext(ctx, query, args)
	if err != nil {
		zap.L().Error(
			"failed to run cleanup action_rules query",
			zap.Time("now", now),
			zap.Error(err),
		)
		return 0
	}

	count, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows after cleanup action_rules",
			zap.Time("now", now),
			zap.Error(err),
		)
		return 0
	}

	zap.L().Info(
		"cleanup of expired action_rules completed",
		zap.Int64("rows deleted", count),
	)

	return count
}
