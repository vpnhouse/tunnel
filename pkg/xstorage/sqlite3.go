// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xstorage

import (
	"embed"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func NewSqlite3(path string, migrations embed.FS) (*sqlx.DB, error) {
	sqlDB, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, xerror.EStorageError("can't open database", err, zap.String("path", path))
	}

	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "db/migrations",
	}

	_, err = migrate.Exec(sqlDB.DB, "sqlite3", migrationSource, migrate.Up)
	if err != nil {
		return nil, xerror.EStorageError("can't perform migration", err)
	}

	return sqlDB, nil
}
