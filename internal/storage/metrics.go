package storage

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/xerror"
	"go.uber.org/zap"
)

type metricRow struct {
	Name  string `db:"name"`
	Value int64  `db:"value"`
}

func (storage *Storage) GetMetric(name string) (int64, error) {
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

func (storage *Storage) GetMetrics(names []string) (map[string]int64, error) {
	query := `SELECT name, value FROM metrics WHERE name IN (?)`
	query, args, _ := sqlx.In(query, names)
	query = storage.db.Rebind(query)

	rows := []metricRow{}
	if err := storage.db.Select(&rows, query, args...); err != nil {
		return nil, xerror.EStorageError("failed to query metric", err)
	}

	result := map[string]int64{}
	for _, row := range rows {
		result[row.Name] = row.Value
	}

	return result, nil
}

func (storage *Storage) SetMetrics(metrics map[string]int64) error {
	if len(metrics) == 0 {
		zap.L().Warn("Skipping store of empty metrics")
		return nil
	}

	rows := make([]metricRow, 0, len(metrics))
	for k, v := range metrics {
		rows = append(rows, metricRow{
			Name:  k,
			Value: v,
		})
	}

	query := `
		INSERT INTO metrics(name, value) VALUES (:name, :value)
		ON CONFLICT(name)
			DO UPDATE
			SET value = EXCLUDED.value`

	if _, err := storage.db.NamedExec(query, rows); err != nil {
		return xerror.EStorageError("failed to insert metric", err, zap.String("query", query))
	}

	return nil
}
