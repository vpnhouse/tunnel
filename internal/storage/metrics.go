package storage

import (
	"database/sql"
	"errors"

	"github.com/vpnhouse/common-lib-go/xerror"
)

func (storage *Storage) getMetric(name string) (int64, error) {
	const q = `SELECT value FROM metrics WHERE name = $1`
	var value int64
	if err := storage.db.QueryRow(q, name).Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, xerror.EEntryNotFound("no such metric", nil)
		}
		return 0, xerror.EStorageError("failed to query metric", err)
	}
	return value, nil
}

func (storage *Storage) setMetric(name string, value int64) error {
	const q = `INSERT INTO metrics(name, value) VALUES ($1, $2)
				ON CONFLICT(name) DO UPDATE SET value=$2`

	if _, err := storage.db.Exec(q, name, value); err != nil {
		return xerror.EStorageError("failed to insert metric", err)
	}

	return nil
}

func (storage *Storage) GetUpstreamMetric() int64 {
	value, _ := storage.getMetric("upstream")
	return value
}

func (storage *Storage) GetDownstreamMetric() int64 {
	value, _ := storage.getMetric("downstream")
	return value
}

func (storage *Storage) SetUpstreamMetric(value int64) {
	storage.setMetric("upstream", value)
}

func (storage *Storage) SetDownstreamMetric(value int64) {
	storage.setMetric("downstream", value)
}
