package database

import (
	"database/sql"
	"fmt"
)

// DriverType represents the database driver type
type DriverType string

const (
	DriverSQLite   DriverType = "sqlite"
	DriverMySQL    DriverType = "mysql"
	DriverPostgres DriverType = "postgres"
)

// Driver defines the interface for database drivers
// Each driver must implement database-specific functionality
type Driver interface {
	// Type returns the driver type
	Type() DriverType

	// Open opens a database connection
	Open(dsn string) (*sql.DB, error)

	// Configure performs database-specific configuration (PRAGMA, SET, etc.)
	Configure(db *sql.DB) error

	// Dialect returns SQL dialect helper functions
	Dialect() Dialect

	// Close performs any cleanup needed
	Close() error
}

// Dialect provides database-specific SQL functions and syntax
type Dialect interface {
	// CurrentTimestamp returns the SQL function for current timestamp
	// SQLite: strftime('%s', 'now')
	// MySQL: UNIX_TIMESTAMP()
	// PostgreSQL: EXTRACT(EPOCH FROM NOW())
	CurrentTimestamp() string

	// CurrentTimestampISO returns ISO 8601 timestamp
	// SQLite: datetime('now')
	// MySQL: NOW()
	// PostgreSQL: NOW()
	CurrentTimestampISO() string

	// CreateTableIfNotExists returns the syntax for CREATE TABLE IF NOT EXISTS
	CreateTableIfNotExists() string

	// AutoIncrement returns the syntax for auto-incrementing primary key
	// SQLite: INTEGER PRIMARY KEY AUTOINCREMENT
	// MySQL: AUTO_INCREMENT
	// PostgreSQL: SERIAL or BIGSERIAL
	AutoIncrement(columnType string) string

	// IntegerType returns the appropriate integer type
	// SQLite: INTEGER
	// MySQL: INT or BIGINT
	// PostgreSQL: INTEGER or BIGINT
	IntegerType() string

	// TextType returns the appropriate text type
	// SQLite: TEXT
	// MySQL: TEXT or VARCHAR
	// PostgreSQL: TEXT
	TextType() string

	// BooleanType returns the appropriate boolean type
	// SQLite: INTEGER (0 or 1)
	// MySQL: BOOLEAN or TINYINT(1)
	// PostgreSQL: BOOLEAN
	BooleanType() string

	// JSONType returns the appropriate JSON type
	// SQLite: TEXT
	// MySQL: JSON
	// PostgreSQL: JSONB
	JSONType() string

	// Placeholder returns the placeholder style for prepared statements
	// SQLite: ?
	// MySQL: ?
	// PostgreSQL: $1, $2, etc.
	Placeholder(index int) string

	// LimitOffset returns the syntax for LIMIT/OFFSET
	// SQLite: LIMIT ? OFFSET ?
	// MySQL: LIMIT ? OFFSET ?
	// PostgreSQL: LIMIT ? OFFSET ?
	LimitOffset(limit, offset int) string

	// TransactionIsolation returns the syntax for setting transaction isolation
	TransactionIsolation(level string) string

	// SupportsTransactions returns whether the database supports transactions
	SupportsTransactions() bool

	// SupportsRowLevelLocking returns whether the database supports SELECT FOR UPDATE
	SupportsRowLevelLocking() bool

	// EscapeIdentifier returns the escaped identifier (for table/column names)
	EscapeIdentifier(name string) string
}

// DriverRegistry manages available database drivers
type DriverRegistry struct {
	drivers map[DriverType]Driver
}

// NewDriverRegistry creates a new driver registry
func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{
		drivers: make(map[DriverType]Driver),
	}
}

// Register registers a database driver
func (r *DriverRegistry) Register(driver Driver) {
	r.drivers[driver.Type()] = driver
}

// Get retrieves a driver by type
func (r *DriverRegistry) Get(driverType DriverType) (Driver, error) {
	driver, ok := r.drivers[driverType]
	if !ok {
		return nil, fmt.Errorf("driver %s not registered", driverType)
	}
	return driver, nil
}

// List returns all registered driver types
func (r *DriverRegistry) List() []DriverType {
	types := make([]DriverType, 0, len(r.drivers))
	for t := range r.drivers {
		types = append(types, t)
	}
	return types
}

// Global driver registry
var globalRegistry *DriverRegistry

// init initializes the global registry with default drivers
func init() {
	globalRegistry = NewDriverRegistry()
	
	// Register default drivers
	globalRegistry.Register(NewSQLiteDriver())
	// MySQL and PostgreSQL drivers are registered on-demand when needed
	// to avoid requiring their dependencies for SQLite-only deployments
}

// RegisterDriver registers a driver in the global registry
func RegisterDriver(driver Driver) {
	globalRegistry.Register(driver)
}

// GetDriver retrieves a driver from the global registry
func GetDriver(driverType DriverType) (Driver, error) {
	return globalRegistry.Get(driverType)
}

// ListDrivers returns all registered driver types
func ListDrivers() []DriverType {
	return globalRegistry.List()
}
