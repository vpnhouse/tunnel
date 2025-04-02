// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"embed"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xstorage"
	"github.com/vpnhouse/tunnel/internal/storage/keycache"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations
var migrations embed.FS

var ErrNotFound = errors.New("not found")

type Storage struct {
	db       *sqlx.DB
	keycache *keycache.Instance
}

func New(path string) (*Storage, error) {
	db, err := xstorage.NewSqlite3(path, migrations)
	if err != nil {
		return nil, err
	}

	storage := &Storage{
		db:       db,
		keycache: keycache.New(),
	}

	keys, err := storage.dbReadAuthorizerKeys()
	if err != nil {
		zap.L().Error("Failed to read authorizer keys, waiting for remote update", zap.Error(err))
	} else {
		storage.keycache.Put(keys)
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
