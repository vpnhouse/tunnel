// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"embed"
	"errors"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xstorage"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations
var migrations embed.FS

var ErrNotFound = errors.New("not found")

type Storage struct {
	db *sqlx.DB

	keyCacheLock sync.RWMutex
	keyCache     map[string]types.AuthorizerKey
}

func New(path string, authKeyCacheInterval time.Duration) (*Storage, error) {
	db, err := xstorage.NewSqlite3(path, migrations)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db:       db,
		keyCache: map[string]types.AuthorizerKey{},
	}

	keys, err := storage.readAuthorizedKeys()
	if err != nil {
		zap.L().Error("Failed to read authorizer keys, waiting for remote update", zap.Error(err))
		return storage, nil
	} else {
		storage.cachePutAuthorizerKeys(keys)
	}

	return storage, nil
}

func (storage *Storage) Shutdown() error {
	err := storage.db.Close()
	if err != nil {
		return xerror.EStorageError("failed close database", err)
	}

	storage.db = nil
	return nil
}

func (storage *Storage) Running() bool {
	return storage.db != nil
}
