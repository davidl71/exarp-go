// Package database: rqlite driver (self-hosted distributed SQLite).
// Uses gorqlite stdlib for database/sql. DSN format: http://host:4001 or https://user:pass@host:4001.
// See docs/RQLITE_BACKEND_PLAN.md.
package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/rqlite/gorqlite/stdlib"
)

// RqliteDriver implements the Driver interface for rqlite (SQLite-compatible SQL).
type RqliteDriver struct {
	dialect *RqliteDialect
}

// NewRqliteDriver creates a new rqlite driver.
func NewRqliteDriver() *RqliteDriver {
	return &RqliteDriver{
		dialect: NewRqliteDialect(),
	}
}

// Type returns the driver type.
func (d *RqliteDriver) Type() DriverType {
	return DriverRqlite
}

// Open opens a rqlite connection. DSN must be a URL (e.g. http://localhost:4001).
func (d *RqliteDriver) Open(dsn string) (*sql.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("rqlite DSN is required")
	}
	if !strings.HasPrefix(dsn, "http://") && !strings.HasPrefix(dsn, "https://") {
		return nil, fmt.Errorf("rqlite DSN must be http or https URL (e.g. http://localhost:4001)")
	}

	db, err := sql.Open("rqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open rqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping rqlite: %w", err)
	}
	return db, nil
}

// Configure sets connection pool; rqlite does not support PRAGMA (remote).
func (d *RqliteDriver) Configure(db *sql.DB) error {
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)
	return nil
}

// Dialect returns the SQL dialect (SQLite-compatible).
func (d *RqliteDriver) Dialect() Dialect {
	return d.dialect
}

// Close is a no-op for rqlite.
func (d *RqliteDriver) Close() error {
	return nil
}

// RqliteDialect implements Dialect for rqlite (same SQL as SQLite).
type RqliteDialect struct{}

// NewRqliteDialect creates a new rqlite dialect.
func NewRqliteDialect() *RqliteDialect {
	return &RqliteDialect{}
}

func (d *RqliteDialect) CurrentTimestamp() string       { return "strftime('%s', 'now')" }
func (d *RqliteDialect) CurrentTimestampISO() string    { return "datetime('now')" }
func (d *RqliteDialect) CreateTableIfNotExists() string { return "CREATE TABLE IF NOT EXISTS" }
func (d *RqliteDialect) AutoIncrement(columnType string) string {
	return fmt.Sprintf("%s PRIMARY KEY AUTOINCREMENT", columnType)
}
func (d *RqliteDialect) IntegerType() string          { return "INTEGER" }
func (d *RqliteDialect) TextType() string             { return "TEXT" }
func (d *RqliteDialect) BooleanType() string          { return "INTEGER" }
func (d *RqliteDialect) JSONType() string             { return "TEXT" }
func (d *RqliteDialect) Placeholder(index int) string { return "?" }
func (d *RqliteDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
func (d *RqliteDialect) TransactionIsolation(level string) string { return "" }
func (d *RqliteDialect) SupportsTransactions() bool               { return true }
func (d *RqliteDialect) SupportsRowLevelLocking() bool            { return false }
func (d *RqliteDialect) EscapeIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
