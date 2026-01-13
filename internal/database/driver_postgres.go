package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// PostgresDriver implements the Driver interface for PostgreSQL
type PostgresDriver struct {
	dialect *PostgresDialect
}

// NewPostgresDriver creates a new PostgreSQL driver
func NewPostgresDriver() *PostgresDriver {
	return &PostgresDriver{
		dialect: NewPostgresDialect(),
	}
}

// Type returns the driver type
func (d *PostgresDriver) Type() DriverType {
	return DriverPostgres
}

// Open opens a PostgreSQL database connection
func (d *PostgresDriver) Open(dsn string) (*sql.DB, error) {
	// PostgreSQL DSN format: postgres://user:password@host:port/dbname?params
	// or: host=host port=port user=user password=password dbname=dbname sslmode=disable
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return db, nil
}

// Configure performs PostgreSQL-specific configuration
func (d *PostgresDriver) Configure(db *sql.DB) error {
	// Set PostgreSQL connection settings
	configs := []string{
		"SET timezone = 'UTC'",
	}

	for _, config := range configs {
		if _, err := db.Exec(config); err != nil {
			return fmt.Errorf("failed to set %s: %w", config, err)
		}
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	return nil
}

// Dialect returns the PostgreSQL dialect
func (d *PostgresDriver) Dialect() Dialect {
	return d.dialect
}

// Close performs cleanup (no-op for PostgreSQL)
func (d *PostgresDriver) Close() error {
	return nil
}

// PostgresDialect implements Dialect for PostgreSQL
type PostgresDialect struct{}

// NewPostgresDialect creates a new PostgreSQL dialect
func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{}
}

func (d *PostgresDialect) CurrentTimestamp() string {
	return "EXTRACT(EPOCH FROM NOW())"
}

func (d *PostgresDialect) CurrentTimestampISO() string {
	return "NOW()"
}

func (d *PostgresDialect) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}

func (d *PostgresDialect) AutoIncrement(columnType string) string {
	// PostgreSQL uses SERIAL or BIGSERIAL
	if strings.Contains(strings.ToUpper(columnType), "BIG") {
		return "BIGSERIAL PRIMARY KEY"
	}
	return "SERIAL PRIMARY KEY"
}

func (d *PostgresDialect) IntegerType() string {
	return "BIGINT"
}

func (d *PostgresDialect) TextType() string {
	return "TEXT"
}

func (d *PostgresDialect) BooleanType() string {
	return "BOOLEAN"
}

func (d *PostgresDialect) JSONType() string {
	return "JSONB" // Use JSONB for better performance
}

func (d *PostgresDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d *PostgresDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d *PostgresDialect) TransactionIsolation(level string) string {
	return fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
}

func (d *PostgresDialect) SupportsTransactions() bool {
	return true
}

func (d *PostgresDialect) SupportsRowLevelLocking() bool {
	return true // PostgreSQL supports SELECT ... FOR UPDATE
}

func (d *PostgresDialect) EscapeIdentifier(name string) string {
	// PostgreSQL uses double quotes for identifiers
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
