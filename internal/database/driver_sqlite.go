// driver_sqlite.go — SQLite driver implementation (primary storage backend).
package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// SQLiteDriver implements the Driver interface for SQLite.
type SQLiteDriver struct {
	dialect *SQLiteDialect
}

// NewSQLiteDriver creates a new SQLite driver.
func NewSQLiteDriver() *SQLiteDriver {
	return &SQLiteDriver{
		dialect: NewSQLiteDialect(),
	}
}

// Type returns the driver type.
func (d *SQLiteDriver) Type() DriverType {
	return DriverSQLite
}

// sqliteOpenDSN converts a filesystem path to a modernc file URI with per-connection PRAGMAs.
// SQLite foreign_keys default off per connection; URI _pragma applies when each connection opens.
func sqliteOpenDSN(dsn string) string {
	if strings.HasPrefix(dsn, "file:") || strings.Contains(dsn, "?") {
		return dsn
	}
	abs := dsn
	if a, err := filepath.Abs(dsn); err == nil {
		abs = a
	}
	path := filepath.ToSlash(abs)
	if len(path) >= 2 && path[1] == ':' {
		// Windows: C:/x -> /C:/x for host-less file URL
		path = "/" + path
	}
	u := url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_pragma=foreign_keys(on)",
	}
	return u.String()
}

// Open opens a SQLite database connection.
func (d *SQLiteDriver) Open(dsn string) (*sql.DB, error) {
	openDSN := dsn
	if !strings.HasPrefix(dsn, "file:") && !strings.Contains(dsn, "?") {
		dir := filepath.Dir(dsn)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		}
		openDSN = sqliteOpenDSN(dsn)
	}

	db, err := sql.Open("sqlite", openDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return db, nil
}

// Configure performs SQLite-specific configuration.
func (d *SQLiteDriver) Configure(db *sql.DB) error {
	// Set SQLite PRAGMA settings (must be done outside transactions)
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 30000",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set %s: %w", pragma, err)
		}
	}

	// Set connection pool settings
	// SQLite with WAL mode supports multiple readers concurrently
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	return nil
}

// Dialect returns the SQLite dialect.
func (d *SQLiteDriver) Dialect() Dialect {
	return d.dialect
}

// Close performs cleanup (no-op for SQLite).
func (d *SQLiteDriver) Close() error {
	return nil
}

// SQLiteDialect implements Dialect for SQLite.
type SQLiteDialect struct{}

// NewSQLiteDialect creates a new SQLite dialect.
func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{}
}

func (d *SQLiteDialect) CurrentTimestamp() string {
	return "strftime('%s', 'now')"
}

func (d *SQLiteDialect) CurrentTimestampISO() string {
	return "datetime('now')"
}

func (d *SQLiteDialect) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}

func (d *SQLiteDialect) AutoIncrement(columnType string) string {
	return fmt.Sprintf("%s PRIMARY KEY AUTOINCREMENT", columnType)
}

func (d *SQLiteDialect) IntegerType() string {
	return "INTEGER"
}

func (d *SQLiteDialect) TextType() string {
	return "TEXT"
}

func (d *SQLiteDialect) BooleanType() string {
	return "INTEGER" // SQLite uses INTEGER (0 or 1) for booleans
}

func (d *SQLiteDialect) JSONType() string {
	return "TEXT" // SQLite stores JSON as TEXT
}

func (d *SQLiteDialect) Placeholder(index int) string {
	return "?"
}

func (d *SQLiteDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d *SQLiteDialect) TransactionIsolation(level string) string {
	// SQLite doesn't support transaction isolation levels
	return ""
}

func (d *SQLiteDialect) SupportsTransactions() bool {
	return true
}

func (d *SQLiteDialect) SupportsRowLevelLocking() bool {
	return false // SQLite doesn't support SELECT FOR UPDATE effectively
}

func (d *SQLiteDialect) EscapeIdentifier(name string) string {
	// SQLite uses double quotes for identifiers
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
