// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"embed"
	"errors"
	"time"

	"github.com/erwint/ttlcache"
	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xstorage"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations
var migrations embed.FS

var ErrNotFound = errors.New("not found")

type Storage struct {
	db              *sqlx.DB
	authKeyCache    *ttlcache.Cache
	authKeyCacheTtl time.Duration
}

func New(path string, authKeyCacheInterval time.Duration) (*Storage, error) {
	db, err := xstorage.NewSqlite3(path, migrations)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db:              db,
		authKeyCache:    ttlcache.NewCache(),
		authKeyCacheTtl: authKeyCacheInterval,
	}, nil
}

func (storage *Storage) Shutdown() error {
	storage.authKeyCache.Close()
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
