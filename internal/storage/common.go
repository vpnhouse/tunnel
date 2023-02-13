// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"embed"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xstorage"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations
var migrations embed.FS

var ErrNotFound = errors.New("not found")

type Storage struct {
	db *sqlx.DB
}

func New(path string) (*Storage, error) {
	db, err := xstorage.NewSqlite3(path, migrations)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
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
