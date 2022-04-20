// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations
var migrations embed.FS

type Storage struct {
	db *sqlx.DB
}

func New(path string) (*Storage, error) {
	sqlDB, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, xerror.EStorageError("can't open database", err, zap.String("path", path))
	}

	storage := &Storage{
		db: sqlDB,
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "db/migrations",
	}

	_, err = migrate.Exec(storage.db.DB, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return nil, xerror.EStorageError("can't perform migration", err)
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

func getFields(i interface{}, omitNull bool) (columns []string) {
	v := reflect.Indirect(reflect.ValueOf(i))

	for i := 0; i < v.NumField(); i++ {
		columnName := v.Type().Field(i).Tag.Get("db")

		fName := v.Type().Field(i).Name
		if fName == "ID" {
			if v.Field(i).IsZero() {
				continue
			}
		}

		field := v.Field(i)
		if columnName != "" {
			if omitNull && field.Kind() == reflect.Ptr && field.IsNil() {
				continue
			}
			columns = append(columns, columnName)
		}

		if field.Kind() == reflect.Struct {
			columns = append(columns, getFields(v.Field(i).Interface(), omitNull)...)
		}
	}

	return
}

func getSelectRequest(table string, i interface{}) (string, error) {
	names := getFields(i, true)
	if len(names) == 0 {
		return fmt.Sprintf("SELECT * FROM %v", table), nil
	}

	selectors := ""
	for i, n := range names {
		if i == 0 {
			selectors = fmt.Sprintf("%v=:%v", n, n)
		} else {
			selectors += fmt.Sprintf(" AND %v=:%v", n, n)
		}
	}

	return fmt.Sprintf("SELECT * FROM %v WHERE %v", table, selectors), nil
}

func getInsertRequest(table string, i interface{}) (string, error) {
	names := getFields(i, false)
	if len(names) == 0 {
		return fmt.Sprintf("INSERT INTO %v", table), nil
	}

	fields := strings.Join(names, ", ")
	refs := ":" + strings.Join(names, ", :")
	return fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v)", table, fields, refs), nil
}

func getUpdateRequest(table string, idName string, i interface{}, skipped []string) (string, error) {
	allNames := getFields(i, false)
	notNullNames := getFields(i, true)

	haveId := false
	for _, n := range notNullNames {
		if n == idName {
			haveId = true
		}
	}

	if !haveId {
		return "", errors.New("id is not specified")
	}

	skippedMap := make(map[string]bool)
	for _, s := range skipped {
		skippedMap[s] = true
	}

	values := ""
	for _, n := range allNames {
		if _, ok := skippedMap[n]; ok || n == idName {
			// Skipping specific field
			continue
		}

		if len(values) == 0 {
			values = fmt.Sprintf("%v=:%v", n, n)
		} else {
			values += fmt.Sprintf(", %v=:%v", n, n)
		}
	}

	return fmt.Sprintf("UPDATE %v SET %v WHERE %v=:%v", table, values, idName, idName), nil
}
