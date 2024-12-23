// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xtime"
	"go.uber.org/zap"

	"github.com/vpnhouse/tunnel/internal/types"
)

func (storage *Storage) AddAction(action *types.Action) error {
	query := `
		INSERT INTO 
			actions(user_id, expires, rules_json) 
		VALUES(:user_id, :expires, :rules_json)
	`
	_, err := storage.db.NamedExec(query, action)
	if err != nil {
		return xerror.EStorageError("can't insert action", err, zap.Any("action", action))
	}
}

func (storage *Storage) FindActions(userId string) ([]*types.Action, error) {
	query := `
		SELECT 
			user_id,
			expires,
			rules_json
		FROM 
			actions
		WHERE
		    user_id = :user_id
			AND coalesce(expires, :now) >= :now
	`

	now := xtime.Now()

	args := map[string]any{
		"now":     now,
		"user_id": userId,
	}

	rows, err := storage.db.NamedQuery(query, args)
	if err != nil {
		return nil, xerror.EStorageError(
			"can't read actions",
			err,
			zap.Timep("now", now.TimePtr()),
			zap.String("user_id", userId),
		)
	}
	defer rows.Close()

	var actions []*types.Action
	for rows.Next() {
		var act types.Action
		err = rows.StructScan(&act)
		if err != nil {
			zap.L().Error("can't scan action", zap.Error(err), zap.String("user_id", userId))
			continue
		}
		actions = append(actions, &act)
	}
	return actions, nil
}

func (storage *Storage) CleanupExpiredActions() {
	query := `
		DELETE FROM actions where expires is not NULL AND expires < :now
	`

	now := xtime.Now()

	args := map[string]any{
		"now": now,
	}

	res, err := storage.db.NamedExec(query, args)
	if err != nil {
		zap.L().Error(
			"failed to run cleanup actions query",
			zap.Timep("now", now.TimePtr()),
			zap.Error(err),
		)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows after cleanup actions",
			zap.Timep("now", now.TimePtr()),
			zap.Error(err),
		)
		return
	}

	zap.L().Info("cleanup actions completed", zap.Int64("rows deleted", count))

}
