//go:build !cgo
// +build !cgo

package database

import (
	"database/sql"
	"fmt"
)

// ODBCDriverNoCGO is a stub implementation when cgo is disabled
// ODBC requires cgo, so this provides a clear error message
type ODBCDriverNoCGO struct{}

// NewODBCDriver creates a new ODBC driver
// Returns a driver that will error on Open() when cgo is not enabled
func NewODBCDriver() *ODBCDriverNoCGO {
	return &ODBCDriverNoCGO{}
}

// Type returns the driver type
func (d *ODBCDriverNoCGO) Type() DriverType {
	return DriverODBC
}

// Open returns an error explaining that cgo is required
func (d *ODBCDriverNoCGO) Open(dsn string) (*sql.DB, error) {
	return nil, fmt.Errorf("ODBC driver requires cgo (CGO_ENABLED=1). Build with: go build -tags cgo or set CGO_ENABLED=1")
}

// Configure is a no-op for the stub
func (d *ODBCDriverNoCGO) Configure(db *sql.DB) error {
	return fmt.Errorf("ODBC driver requires cgo")
}

// Dialect returns a stub dialect
func (d *ODBCDriverNoCGO) Dialect() Dialect {
	return &ODBCDialectNoCGO{}
}

// Close is a no-op
func (d *ODBCDriverNoCGO) Close() error {
	return nil
}

// ODBCDialectNoCGO is a stub dialect for when cgo is disabled
type ODBCDialectNoCGO struct{}

func (d *ODBCDialectNoCGO) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

func (d *ODBCDialectNoCGO) CurrentTimestampISO() string {
	return "CURRENT_TIMESTAMP"
}

func (d *ODBCDialectNoCGO) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}

func (d *ODBCDialectNoCGO) AutoIncrement(columnType string) string {
	return fmt.Sprintf("%s AUTO_INCREMENT PRIMARY KEY", columnType)
}

func (d *ODBCDialectNoCGO) IntegerType() string {
	return "BIGINT"
}

func (d *ODBCDialectNoCGO) TextType() string {
	return "TEXT"
}

func (d *ODBCDialectNoCGO) BooleanType() string {
	return "BOOLEAN"
}

func (d *ODBCDialectNoCGO) JSONType() string {
	return "TEXT"
}

func (d *ODBCDialectNoCGO) Placeholder(index int) string {
	return "?"
}

func (d *ODBCDialectNoCGO) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d *ODBCDialectNoCGO) TransactionIsolation(level string) string {
	return fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
}

func (d *ODBCDialectNoCGO) SupportsTransactions() bool {
	return true
}

func (d *ODBCDialectNoCGO) SupportsRowLevelLocking() bool {
	return true
}

func (d *ODBCDialectNoCGO) EscapeIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}
