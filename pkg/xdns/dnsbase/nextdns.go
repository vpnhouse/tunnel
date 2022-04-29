package dnsbase

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Categories []string

func (c Categories) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return strings.Join(c, ","), nil
}

func (c *Categories) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	dValue, err := driver.String.ConvertValue(v)
	if err != nil {
		return err
	}

	s, ok := dValue.(string)
	if !ok {
		return fmt.Errorf("scan into Categories: expecting string, got %T", dValue)
	}

	if len(s) > 0 {
		*c = strings.Split(s, ",")
	}

	return nil
}

type Domain struct {
	ID         int64
	ParentID   int64
	Name       string
	Categories Categories
	Parent     *Domain
}

func (d *Domain) Unchain() string {
	var ss []string
	for c := d; c != nil; c = c.Parent {
		ss = append(ss, c.Name)
	}
	return strings.Join(ss, ".")
}

func (d *Domain) String() string {
	name := d.Unchain()
	return fmt.Sprintf("id=%d: %s (p=%d, cat=%v)", d.ID, name, d.ParentID, d.Categories)
}

func (d *Domain) isChild() bool {
	return d.Categories != nil
}

type reader struct {
	db *sql.DB
}

func NewReader(path string) (*reader, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("reader: failed to stat the DB at path %s: %w", path, err)
	}

	db, err := sql.Open("sqlite3", "file:"+path+"?cache=shared&mode=ro&_cache_size=1000000&immutable=true&_journal_mode=OFF")
	if err != nil {
		return nil, fmt.Errorf("reader: failed to open DNS database at %s: %w", path, err)
	}

	return &reader{db: db}, nil
}

func (r *reader) Lookup(name string) (*Domain, error) {
	if len(name) < 2 {
		return nil, fmt.Errorf("lookup: invalid domain name given")
	}

	name = normalizeDomain(name)
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("lookup: invalid domain name given")
	}
	parentName := parts[len(parts)-1]

	// lookup parent category, like ".org", ".ru", ".com", etc
	parent, err := r.lookupParent(parentName)
	if err != nil {
		return nil, err
	}

	var wildcard *Domain
	for i := len(parts) - 2; i >= 0; i-- {
		sub, exactErr := r.lookupSubdomain(parts[i], parent.ID)
		if wc, wcErr := r.lookupSubdomain("*", parent.ID); wcErr == nil {
			wildcard = wc
		}

		if exactErr != nil {
			// exact match does not found,
			// try to return the wildcard match, if any.
			if wildcard != nil {
				wildcard.Parent = parent
				return wildcard, nil
			}
			return nil, exactErr
		}

		sub.Parent = parent
		parent = sub
	}

	// Handle the situation when the db contains the tree like
	// com -> google -> mail
	// and we're trying to match `google.com` domain.
	// In that case we ended up here having a pointer to a `google` node,
	// which is not the child node.
	if !parent.isChild() {
		return nil, fmt.Errorf("lookup: no matches")
	}

	return parent, nil
}

func (r *reader) Close() {
	_ = r.db.Close()
}

func (r *reader) lookupParent(name string) (*Domain, error) {
	return lookupRecord(r.db, name, 0)
}

func (r *reader) lookupSubdomain(name string, parentID int64) (*Domain, error) {
	return lookupRecord(r.db, name, parentID)
}

type sqlQuerier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func lookupRecord(db sqlQuerier, name string, parentID int64) (*Domain, error) {
	row := db.QueryRow("SELECT id, category FROM domains WHERE name = ? and parent_id = ?", name, parentID)
	if row.Err() != nil {
		return nil, fmt.Errorf("sql: failed to query subdomain `%s` (pid=%d): %w", name, parentID, row.Err())
	}

	dom := Domain{
		Name:     name,
		ParentID: parentID,
	}
	if err := row.Scan(&dom.ID, &dom.Categories); err != nil {
		// note: sql.ErrNoRows may happen here, we MUST preserve the info about such an error
		return nil, fmt.Errorf("sql: failed to scan results (%s %d): %w", name, parentID, err)
	}

	return &dom, nil
}

func insertRecord(db sqlExecutor, name string, parentID int64, cats Categories) (int64, error) {
	res, err := db.Exec("insert into domains(name, parent_id, category) values (?, ?, ?)", name, parentID, cats)
	if err != nil {
		return -1, fmt.Errorf("sql: failed to insert record (%s, pid=%d): %w", name, parentID, err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func updateRecordCategories(db sqlExecutor, id int64, newCats Categories) error {
	// TODO(nikonov): remove after tests
	if id == 0 {
		panic("empty id")
	}
	if len(newCats) == 0 {
		panic("no cats")
	}

	res, err := db.Exec("update domains set category = ? where id = ?", newCats, id)
	if err != nil {
		return fmt.Errorf("sql: failed to update categories of id=%d: %w", id, err)
	}
	if ra, _ := res.RowsAffected(); ra != 1 {
		return fmt.Errorf("sql: got unexpected number of rows affected (want=1, got=%d)", ra)
	}
	return nil
}

const schemaSQL = `CREATE TABLE IF NOT EXISTS domains (
		"id" integer NOT NULL PRIMARY KEY,
		"parent_id" integer NOT NULL,
		"name" TEXT not null,
		"category" TEXT NULL -- NULL if not child node
	  );
	create index if not exists parent_id_idx on domains(parent_id);
	create index if not exists name_idx on domains(name);
	create index if not exists category_idx on domains(category);
    create unique index if not exists parent_id_name on domains(parent_id, name);`

type writer struct {
	db *sql.DB
}

func NewWriter(path string) (*writer, error) {
	db, err := sql.Open("sqlite3", "file:"+path+"?_journal=MEMORY&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("writer: failed to open DNS database at %s: %w", path, err)
	}

	if _, err = db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("writer: failed to apply migrations: %w", err)
	}

	return &writer{db: db}, nil
}

func validateDomain(parts []string) error {
	if len(parts) < 2 {
		return errors.New("too few parts")
	}

	for i := 0; i < len(parts); i++ {
		if parts[i] == "*" && i != 0 {
			return errors.New("only wildcards at the beginning are supported")
		}
	}

	return nil
}

func (w *writer) Write(domain string, cats Categories) error {
	domain = normalizeDomain(domain)
	parts := strings.Split(domain, ".")

	if err := validateDomain(parts); err != nil {
		return fmt.Errorf("writer: failed to validate the domain %s: %w", domain, err)
	}

	if len(cats) == 0 {
		return fmt.Errorf("writer: empty categories list given")
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("writer: failed to begin transaction")
	}
	commit := false
	defer func() {
		if commit {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	var parentID int64
	for i := len(parts) - 1; i >= 0; i-- {
		dom, err := lookupRecord(tx, parts[i], parentID)
		if err != nil {
			// we don't have such record at all
			if errors.Is(err, sql.ErrNoRows) {
				// insert categories only to the child, but not to intermediate parents
				var pid int64
				if i == 0 {
					pid, err = insertRecord(tx, parts[i], parentID, cats)
				} else {
					pid, err = insertRecord(tx, parts[i], parentID, nil)
				}
				if err != nil {
					return err
				}

				parentID = pid
				continue
			}

			return fmt.Errorf("writer: lookup failed: %w", err)
		}

		// we already have the child node, so the full record does exist, return the error
		if i == 0 && dom.isChild() {
			return fmt.Errorf("writer: given record `%s` does already exists", domain)
		}

		parentID = dom.ID
	}

	// we have the node for such a fqdn part, but it's not the child node,
	// so turning it to the child.
	if err := updateRecordCategories(tx, parentID, cats); err != nil {
		return err
	}

	commit = true
	return nil
}

// TODO(nikonov): delete the whole sub-tree / prefix
func (w *writer) Delete(id int64) error {
	if _, err := w.db.Exec("DELETE from domains where id = ?", id); err != nil {
		return fmt.Errorf("writer: failed to delete the record: %w", err)
	}
	return nil
}

// Close closes the db, one must not use
// the writer instance after calling Close.
func (w *writer) Close() error {
	return w.db.Close()
}

func isUniqueConstraintError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "UNIQUE constraint failed")
}

func normalizeDomain(s string) string {
	if s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
