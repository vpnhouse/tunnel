// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xcrypto

import (
	"crypto/rsa"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RSAKeeper struct {
	lock    sync.RWMutex
	root    string
	running bool
}

type KeyInfo struct {
	Id  uuid.UUID
	Key *rsa.PublicKey
}

func NewRSAKeeper(root string) (*RSAKeeper, error) {
	keeper := &RSAKeeper{
		root: root,
	}

	return keeper, nil
}

func (keeper *RSAKeeper) Shutdown() error {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	keeper.running = false
	return nil
}

func (keeper *RSAKeeper) Running() bool {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	return keeper.running
}

func (keeper *RSAKeeper) ListKeys() ([]KeyInfo, error) {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	files, err := ioutil.ReadDir(keeper.root)
	if err != nil {
		return nil, xerror.EInternalError("can't list keys", err)
	}

	result := make([]KeyInfo, 0)
	for _, f := range files {
		keyUUID, err := uuid.Parse(f.Name())
		if err != nil {
			zap.L().Error("Skipping file with non-uuid name", zap.String("name", f.Name()), zap.Error(err))
			continue
		}

		key, err := keeper.readKey(keyUUID)
		if err != nil {
			zap.L().Error("file with invalid key data", zap.String("name", f.Name()), zap.Error(err))
			continue
		}

		result = append(result, KeyInfo{keyUUID, key})
	}

	return result, nil
}

func (keeper *RSAKeeper) DeleteKey(keyUUID uuid.UUID) error {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	if keyUUID == uuid.Nil {
		return xerror.EInvalidArgument("nil uuid is not allowed", nil)
	}

	err := os.Remove(keeper.keyPath(keyUUID))
	if err != nil {
		return makeError(err, "can't delete key")
	}

	return nil
}

func (keeper *RSAKeeper) GetKey(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
	keeper.lock.RLock()
	defer keeper.lock.RUnlock()

	return keeper.readKey(keyUUID)
}

func (keeper *RSAKeeper) AddKey(keyUUID uuid.UUID, key *rsa.PublicKey) error {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	return keeper.writeKey(keyUUID, key, os.O_CREATE|os.O_EXCL|os.O_WRONLY)
}

func (keeper *RSAKeeper) ChangeKey(keyUUID uuid.UUID, key *rsa.PublicKey) error {
	keeper.lock.Lock()
	defer keeper.lock.Unlock()

	return keeper.writeKey(keyUUID, key, os.O_WRONLY)
}

func (keeper *RSAKeeper) keyPath(keyUUID uuid.UUID) string {
	return filepath.Join(keeper.root, keyUUID.String())
}

func makeError(err error, opMessage string) error {
	if os.IsNotExist(err) {
		return xerror.EEntryNotFound("key not found", err)
	} else {
		return xerror.EInternalError(opMessage, err)
	}
}

func (keeper *RSAKeeper) readKey(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
	if keyUUID == uuid.Nil {
		return nil, xerror.EInvalidArgument("nil uuid is not allowed", nil)
	}

	pemBytes, err := ioutil.ReadFile(keeper.keyPath(keyUUID))
	if err != nil {
		return nil, makeError(err, "can't read key")
	}

	return UnmarshalPublicKey(pemBytes)
}

func (keeper *RSAKeeper) writeKey(keyUUID uuid.UUID, key *rsa.PublicKey, osFlags int) error {
	if keyUUID == uuid.Nil {
		return xerror.EInvalidArgument("nil uuid is not allowed", nil)
	}

	keyBytes, err := MarshalPublicKey(key)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(keeper.keyPath(keyUUID), osFlags, 0600)
	if err != nil {
		if os.IsExist(err) {
			return xerror.EExists("key with same ID already exists", err)
		} else if os.IsNotExist(err) {
			return xerror.EExists("key does not exists", err)
		} else {
			return xerror.EInternalError("can't store key", err)
		}
	}

	if _, err := fd.Write(keyBytes); err != nil {
		return xerror.EInternalError("can't store key", err)
	}

	return nil
}
