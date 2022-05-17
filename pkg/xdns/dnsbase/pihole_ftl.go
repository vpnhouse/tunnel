package dnsbase

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

type ftlReader struct {
	db *sql.DB
}

func NewFTLReader(path string) (LookupInterface, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("ftl: failed to stat the DB at path %s: %w", path, err)
	}

	db, err := sql.Open("sqlite3", "file:"+path+"?cache=shared&mode=ro&_cache_size=1000000&immutable=true&_journal_mode=OFF")
	if err != nil {
		return nil, fmt.Errorf("ftl: failed to open DNS database at %s: %w", path, err)
	}

	return &ftlReader{db: db}, nil
}

func (ftl *ftlReader) Lookup(name string) (*Domain, error) {
	name = normalizeDomain(name)

	row := ftl.db.QueryRow("select 1 from gravity where domain = ?", name)
	var count int64 = -1
	if err := row.Scan(&count); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("FTL lookup failure: unexpected error: name=%s, err=%v\n", name, err)
		}
		return nil, fmt.Errorf("ftl: lookup failed: %w", err)
	}

	return &Domain{Name: name}, nil
}
