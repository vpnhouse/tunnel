package storage

import (
	"embed"

	libCommon "github.com/Codename-Uranium/common/common"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

//go:embed db/migrations
var migrations embed.FS

type Storage struct {
	db *sqlx.DB
}

func New(path string) (*Storage, error) {
	sqlDB, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, libCommon.EStorageError("can't open database", err, zap.String("path", path))
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
		return nil, libCommon.EStorageError("can't perform migration", err)
	}

	return storage, nil
}

func (storage *Storage) Shutdown() error {
	err := storage.db.Close()
	if err != nil {
		return libCommon.EStorageError("failed close database", err)
	}

	storage.db = nil
	return nil
}

func (storage *Storage) Running() bool {
	return storage.db != nil
}
