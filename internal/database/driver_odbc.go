//go:build cgo && !no_odbc
// +build cgo,!no_odbc

package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/alexbrainman/odbc"
)

// ODBCDriver implements the Driver interface for ODBC connections
// This allows connecting to any database that has an ODBC driver
// Requires cgo and ODBC driver manager (unixODBC on Linux, iODBC on macOS, Windows ODBC)
type ODBCDriver struct {
	dialect *ODBCDialect
}

// NewODBCDriver creates a new ODBC driver
// Note: This requires cgo to be enabled (CGO_ENABLED=1)
// Returns the driver instance (no error when cgo is enabled)
func NewODBCDriver() *ODBCDriver {
	return &ODBCDriver{
		dialect: NewODBCDialect(),
	}
}

// Type returns the driver type
func (d *ODBCDriver) Type() DriverType {
	return DriverODBC
}

// Open opens an ODBC database connection
// DSN format examples:
//   - DSN name: "DSN=todo2"
//   - Full connection string: "Driver={PostgreSQL};Server=localhost;Database=todo2;UID=user;PWD=password"
//   - MySQL ODBC: "Driver={MySQL ODBC 8.0 Driver};Server=localhost;Database=todo2;UID=user;PWD=password"
//   - SQL Server: "Driver={ODBC Driver 17 for SQL Server};Server=localhost;Database=todo2;UID=user;PWD=password"
func (d *ODBCDriver) Open(dsn string) (*sql.DB, error) {
	// ODBC DSN format: DSN=name or full connection string
	db, err := sql.Open("odbc", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open ODBC database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ODBC database: %w", err)
	}

	return db, nil
}

// Configure performs ODBC-specific configuration
// ODBC configuration is typically done via the DSN or connection string
// Some databases may support SET commands
func (d *ODBCDriver) Configure(db *sql.DB) error {
	// Try to set timezone to UTC (may not work for all databases)
	configs := []string{
		"SET timezone = 'UTC'",
		"SET time_zone = '+00:00'",
	}

	for _, config := range configs {
		if _, err := db.Exec(config); err != nil {
			// Ignore errors - not all databases support these commands
			_ = err
		}
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	return nil
}

// Dialect returns the ODBC dialect
// ODBC dialect tries to be generic and compatible with most databases
func (d *ODBCDriver) Dialect() Dialect {
	return d.dialect
}

// Close performs cleanup (no-op for ODBC)
func (d *ODBCDriver) Close() error {
	return nil
}

// ODBCDialect implements Dialect for ODBC connections
// This is a generic dialect that tries to work with most databases
// The actual SQL syntax depends on the underlying database
type ODBCDialect struct {
	// Detect the underlying database type from connection
	// For now, we use generic SQL that works with most databases
}

// NewODBCDialect creates a new ODBC dialect
func NewODBCDialect() *ODBCDialect {
	return &ODBCDialect{}
}

// CurrentTimestamp returns a generic timestamp function
// Most databases support UNIX_TIMESTAMP() or similar
// Falls back to database-specific functions
func (d *ODBCDialect) CurrentTimestamp() string {
	// Try to use a generic function that works with most databases
	// Individual databases may need to override this
	return "UNIX_TIMESTAMP()" // Works with MySQL, some others
	// Alternative: "EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)" for PostgreSQL
	// SQLite would need: "strftime('%s', 'now')"
}

// CurrentTimestampISO returns ISO 8601 timestamp
func (d *ODBCDialect) CurrentTimestampISO() string {
	return "CURRENT_TIMESTAMP" // Standard SQL, works with most databases
}

// CreateTableIfNotExists returns the syntax for CREATE TABLE IF NOT EXISTS
func (d *ODBCDialect) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS" // Standard SQL, works with most databases
}

// AutoIncrement returns the syntax for auto-incrementing primary key
// This is database-specific, so we provide a generic version
func (d *ODBCDialect) AutoIncrement(columnType string) string {
	// Generic approach - may need database-specific handling
	// MySQL: AUTO_INCREMENT
	// PostgreSQL: SERIAL
	// SQL Server: IDENTITY(1,1)
	// SQLite: AUTOINCREMENT
	return fmt.Sprintf("%s AUTO_INCREMENT PRIMARY KEY", columnType)
}

// IntegerType returns the appropriate integer type
func (d *ODBCDialect) IntegerType() string {
	return "BIGINT" // Use BIGINT for compatibility
}

// TextType returns the appropriate text type
func (d *ODBCDialect) TextType() string {
	return "TEXT" // Standard SQL TEXT type
}

// BooleanType returns the appropriate boolean type
func (d *ODBCDialect) BooleanType() string {
	return "BOOLEAN" // Standard SQL boolean (may need conversion for some databases)
}

// JSONType returns the appropriate JSON type
func (d *ODBCDialect) JSONType() string {
	return "TEXT" // Use TEXT for maximum compatibility (JSON stored as text)
}

// Placeholder returns the placeholder style for prepared statements
// ODBC typically uses ? for most drivers, but some may use $1, $2, etc.
func (d *ODBCDialect) Placeholder(index int) string {
	// Most ODBC drivers use ? placeholders
	// PostgreSQL ODBC may use $1, $2, etc., but the driver handles this
	return "?"
}

// LimitOffset returns the syntax for LIMIT/OFFSET
func (d *ODBCDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// TransactionIsolation returns the syntax for setting transaction isolation
func (d *ODBCDialect) TransactionIsolation(level string) string {
	return fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
}

// SupportsTransactions returns whether the database supports transactions
func (d *ODBCDialect) SupportsTransactions() bool {
	return true // Most ODBC databases support transactions
}

// SupportsRowLevelLocking returns whether the database supports SELECT FOR UPDATE
func (d *ODBCDialect) SupportsRowLevelLocking() bool {
	return true // Most ODBC databases support row-level locking
}

// EscapeIdentifier returns the escaped identifier
// ODBC typically uses double quotes or backticks depending on the database
func (d *ODBCDialect) EscapeIdentifier(name string) string {
	// Use double quotes for standard SQL compatibility
	// Some databases may need backticks or brackets
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
