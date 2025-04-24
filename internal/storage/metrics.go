package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/vpnhouse/common-lib-go/xerror"
)

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

func (storage *Storage) GetMetricsLike(patterns []string) (map[string]int64, error) {
	q := `SELECT name, value FROM metrics WHERE`
	args := make([]string, len(patterns))
	for idx, p := range patterns {
		args = append(args, p)
		if idx == 0 {
			q += fmt.Sprintf(" name like '$%d'", idx+1)
		} else {
			q += fmt.Sprintf(" or name like '$%d'", idx+1)
		}
	}

	rows, err := storage.db.Query(q, args)
	if err != nil {
		return nil, xerror.EStorageError("failed to query metric", err)
	}

	result := map[string]int64{}
	for rows.Next() {
		var (
			name  string
			value int64
		)
		if err := rows.Scan(&name, &value); err != nil {
			return nil, xerror.EStorageError("failed to query metric", err)
		}
		result[name] = value
	}

	return result, nil
}

func (storage *Storage) SetMetrics(metrics map[string]int64) error {
	args := make([]map[string]interface{}, len(metrics))
	for name, value := range metrics {
		args = append(args, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	const query = `
		INSERT INTO metrics(name, value) 
		VALUES (:name, :value)
		ON CONFLICT(name) 
			DO UPDATE
			SET value = EXCLUDED.value`

	if _, err := storage.db.NamedExec(query, metrics); err != nil {
		return xerror.EStorageError("failed to insert metric", err)
	}

	return nil
}
