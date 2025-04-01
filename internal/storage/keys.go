// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/tunnel/internal/types"
)

func (storage *Storage) UpdateAuthorizerKeys(keys []types.AuthorizerKey) error {
	return storage.keys.Update(keys)
}

func (storage *Storage) GetAuthorizerKeyByID(id string) (types.AuthorizerKey, error) {
	return storage.keys.Get(id)
}

func (storage *Storage) ListAuthorizerKeys() ([]types.AuthorizerKey, error) {
	return storage.keys.List()
}

func (storage *Storage) DeleteAuthorizerKey(id string) error {
	return storage.keys.Delete(id)
}

func (storage *Storage) AsKeystore() auth.KeyStore {
	return storage.keys.AsKeystore()
}
