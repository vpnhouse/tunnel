// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"crypto/rsa"

	"github.com/google/uuid"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/xerror"
	"go.uber.org/zap"
)

func (storage *Storage) dbReadAuthorizerKeys() ([]types.AuthorizerKey, error) {
	q := `select id, source, key from authorizer_keys`

	rows := []struct {
		ID     string `db:"id"`
		Source string `db:"source"`
		Key    string `db:"key"`
	}{}
	err := storage.db.Select(&rows, q)
	if err != nil {
		return nil, xerror.EStorageError("failed to get authorizer keys", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	keys := make([]types.AuthorizerKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, types.AuthorizerKey{
			ID:     row.ID,
			Source: row.Source,
			Key:    row.Source,
		})
	}

	return keys, nil
}

func (storage *Storage) dbWriteAuthorizerKeys(keys []types.AuthorizerKey) error {
	if len(keys) == 0 {
		// grumble about the api misuse
		return xerror.EInvalidArgument("empty key list given", nil)
	}

	tx, err := storage.db.Begin()
	if err != nil {
		return xerror.EStorageError("failed to start transaction", err)
	}
	defer tx.Rollback() //nolint:errcheck

	q := `
		insert into authorizer_keys(id, source, key) 
		values ($1, $2, $3)
		on conflict(id) do update set source=$2,key=$3
	`

	for _, key := range keys {
		_, err := tx.Exec(q, key.ID, key.Source, key.Key)
		if err != nil {
			return xerror.EStorageError("failed to insert key", err,
				zap.String("id", key.ID), zap.String("source", key.Source))
		}
	}

	err = tx.Commit()
	if err != nil {
		return xerror.EStorageError("failed to commit key changes", err)
	}

	return nil
}

func (storage *Storage) dbDeleteAuthorizerKey(id string) error {
	if id == "" {
		return xerror.EInvalidArgument("key id is empty", nil)
	}

	q := `delete from authorizer_keys where id = $1`
	_, err := storage.db.Exec(q, id)
	if err != nil {
		return xerror.EStorageError("failed to delete authorizer key", nil)
	}

	return nil
}

func (storage *Storage) UpdateAuthorizerKeys(keys []types.AuthorizerKey) error {
	err := storage.dbWriteAuthorizerKeys(keys)
	if err != nil {
		return err
	}

	storage.keycache.Put(keys)
	return nil
}

func (storage *Storage) GetAuthorizerKeyByID(id string) (types.AuthorizerKey, error) {
	return storage.keycache.Get(id)
}

func (storage *Storage) ListAuthorizerKeys() ([]types.AuthorizerKey, error) {
	return storage.keycache.List(), nil
}

func (storage *Storage) DeleteAuthorizerKey(id string) error {
	err := storage.dbDeleteAuthorizerKey(id)
	if err != nil {
		return err
	}

	storage.keycache.Delete(id)
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
