// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"crypto/rsa"
	"database/sql"
	"errors"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateAuthorizerKeys updates or inserts the given keys.
// The methods do not validate the key content.
func (storage *Storage) UpdateAuthorizerKeys(keys []types.AuthorizerKey) error {
	if len(keys) == 0 {
		// grumble about the api misuse
		return xerror.EInvalidArgument("empty key list given", nil)
	}

	tx, err := storage.db.Begin()
	if err != nil {
		return xerror.EStorageError("failed to start transaction", err)
	}

	const q = `insert into authorizer_keys(id, source, key) values ($1, $2, $3)
				on conflict(id) do update set source=$2,key=$3`

	for _, key := range keys {
		if _, err := tx.Exec(q, key.ID, key.Source, key.Key); err != nil {
			_ = tx.Rollback()
			return xerror.EStorageError("failed to insert key", err,
				zap.String("id", key.ID), zap.String("source", key.Source))
		}
	}

	return tx.Commit()
}

func (storage *Storage) GetAuthorizerKeyByID(id string) (types.AuthorizerKey, error) {
	var key types.AuthorizerKey
	const q = `select id, source, key from authorizer_keys where id = $1`
	if err := storage.db.QueryRow(q, id).Scan(&key.ID, &key.Source, &key.Key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.AuthorizerKey{}, xerror.EEntryNotFound("no such key", nil)
		}
		return types.AuthorizerKey{}, xerror.EStorageError("failed to query for a key with a given id", err)
	}
	return key, nil
}

func (storage *Storage) ListAuthorizerKeys() ([]types.AuthorizerKey, error) {
	const q = `select id, source, key from authorizer_keys`
	rows, err := storage.db.Query(q)
	if err != nil {
		return nil, xerror.EStorageError("failed to query authorizer keys", err)
	}

	defer rows.Close()

	var keys []types.AuthorizerKey
	for rows.Next() {
		var key types.AuthorizerKey
		if err := rows.Scan(&key.ID, &key.Source, &key.Key); err != nil {
			return nil, xerror.EStorageError("failed to scan into the types.AuthorizerKey", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (storage *Storage) DeleteAuthorizerKey(id string) error {
	if len(id) == 0 {
		return xerror.EInvalidArgument("empty id given", nil)
	}

	const q = `delete from authorizer_keys where id = $1`
	if _, err := storage.db.Exec(q, id); err != nil {
		return xerror.EStorageError("failed to delete authorizer key", nil)
	}

	return nil
}

func (storage *Storage) AsKeystore() auth.KeyStore {
	return &auth.KeyStoreWrapper{
		Fn: func(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
			key, err := storage.GetAuthorizerKeyByID(keyUUID.String())
			if err != nil {
				return nil, err
			}
			pubKey, err := key.Unwrap()
			if err != nil {
				return nil, err
			}
			return pubKey.Key, nil
		},
	}
}
