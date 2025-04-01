// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"crypto/rsa"

	"github.com/google/uuid"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
)

func (storage *Storage) readAuthorizedKeys() ([]types.AuthorizerKey, error) {
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

func (storage *Storage) writeAuthorizerKeys(keys []types.AuthorizerKey) error {
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (storage *Storage) deleteAuthorizerKey(id string) error {
	if len(id) == 0 {
		return xerror.EInvalidArgument("empty id given", nil)
	}

	const q = `delete from authorizer_keys where id = $1`
	if _, err := storage.db.Exec(q, id); err != nil {
		return xerror.EStorageError("failed to delete authorizer key", nil)
	}

	return nil
}

func (storage *Storage) cachePutAuthorizerKeys(keys []types.AuthorizerKey) {
	storage.keyCacheLock.Lock()
	defer storage.keyCacheLock.Unlock()

	for idx := range keys {
		storage.keyCache[keys[idx].ID] = keys[idx]
	}
}

func (storage *Storage) cacheListAuthorizedKeys() []types.AuthorizerKey {
	storage.keyCacheLock.RLock()
	defer storage.keyCacheLock.RUnlock()

	result := make([]types.AuthorizerKey, len(storage.keyCache))
	idx := 0
	for _, v := range storage.keyCache {
		result[idx] = v
		idx++
	}

	return result
}

func (storage *Storage) cacheGetAuthorizerKey(id string) (types.AuthorizerKey, error) {
	storage.keyCacheLock.RLock()
	defer storage.keyCacheLock.RUnlock()

	key, found := storage.keyCache[id]
	if !found {
		return types.AuthorizerKey{}, xerror.EEntryNotFound("no such key", nil)
	}

	return key, nil
}

func (storage *Storage) cacheDeleteAuthorizerKey(id string) {
	storage.keyCacheLock.Lock()
	defer storage.keyCacheLock.Unlock()

	delete(storage.keyCache, id)
}

func (storage *Storage) UpdateAuthorizerKeys(keys []types.AuthorizerKey) error {
	err := storage.writeAuthorizerKeys(keys)
	if err != nil {
		return err
	}

	storage.cachePutAuthorizerKeys(keys)
	return nil
}

func (storage *Storage) GetAuthorizerKeyByID(id string) (types.AuthorizerKey, error) {
	return storage.cacheGetAuthorizerKey(id)
}

func (storage *Storage) ListAuthorizerKeys() ([]types.AuthorizerKey, error) {

	return storage.cacheListAuthorizedKeys(), nil
}

func (storage *Storage) DeleteAuthorizerKey(id string) error {
	err := storage.deleteAuthorizerKey(id)
	if err != nil {
		return err
	}

	storage.cacheDeleteAuthorizerKey(id)
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
