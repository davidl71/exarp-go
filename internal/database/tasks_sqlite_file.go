// tasks_sqlite_file.go — Read-only SQLite helpers for importing external Todo2 DBs.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// sqliteReadOnlyURI builds a modernc/sqlite file URI with mode=ro.
func sqliteReadOnlyURI(path string) (string, error) {
	abs := path
	if a, err := filepath.Abs(path); err == nil {
		abs = a
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	p := filepath.ToSlash(abs)
	if len(p) >= 2 && p[1] == ':' {
		p = "/" + p
	}
	u := url.URL{
		Scheme: "file",
		Path:   p,
	}
	// mode=ro avoids accidental writes; foreign_keys for consistent reads
	u.RawQuery = "mode=ro&_pragma=foreign_keys(on)"

	return u.String(), nil
}

// OpenSQLiteReadOnly opens a todo2.db (or any SQLite file) read-only. Caller must Close().
func OpenSQLiteReadOnly(sqlitePath string) (*sqlx.DB, error) {
	dsn, err := sqliteReadOnlyURI(sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite import path %q: %w", sqlitePath, err)
	}

	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite read-only: %w", err)
	}

	db := sqlx.NewDb(raw, "sqlite")
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

// ListTasksFromSQLiteFile loads tasks (with tags + dependencies) from a standalone SQLite file.
func ListTasksFromSQLiteFile(ctx context.Context, sqlitePath string, filters *TaskFilters) ([]*Todo2Task, error) {
	db, err := OpenSQLiteReadOnly(sqlitePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx = ensureContext(ctx)
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()

	return listTasksFromDB(queryCtx, db, filters)
}
