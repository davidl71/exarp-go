package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDriver implements the Driver interface for MySQL
type MySQLDriver struct {
	dialect *MySQLDialect
}

// NewMySQLDriver creates a new MySQL driver
func NewMySQLDriver() *MySQLDriver {
	return &MySQLDriver{
		dialect: NewMySQLDialect(),
	}
}

// Type returns the driver type
func (d *MySQLDriver) Type() DriverType {
	return DriverMySQL
}

// Open opens a MySQL database connection
func (d *MySQLDriver) Open(dsn string) (*sql.DB, error) {
	// MySQL DSN format: user:password@tcp(host:port)/dbname?params
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return db, nil
}

// Configure performs MySQL-specific configuration
func (d *MySQLDriver) Configure(db *sql.DB) error {
	// Set MySQL connection settings
	configs := []string{
		"SET sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO'",
		"SET time_zone = '+00:00'", // Use UTC
	}

	for _, config := range configs {
		if _, err := db.Exec(config); err != nil {
			// Some configs might fail on older MySQL versions, continue
			_ = err
		}
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	return nil
}

// Dialect returns the MySQL dialect
func (d *MySQLDriver) Dialect() Dialect {
	return d.dialect
}

// Close performs cleanup (no-op for MySQL)
func (d *MySQLDriver) Close() error {
	return nil
}

// MySQLDialect implements Dialect for MySQL
type MySQLDialect struct{}

// NewMySQLDialect creates a new MySQL dialect
func NewMySQLDialect() *MySQLDialect {
	return &MySQLDialect{}
}

func (d *MySQLDialect) CurrentTimestamp() string {
	return "UNIX_TIMESTAMP()"
}

func (d *MySQLDialect) CurrentTimestampISO() string {
	return "NOW()"
}

func (d *MySQLDialect) CreateTableIfNotExists() string {
	return "CREATE TABLE IF NOT EXISTS"
}

func (d *MySQLDialect) AutoIncrement(columnType string) string {
	return fmt.Sprintf("%s AUTO_INCREMENT PRIMARY KEY", columnType)
}

func (d *MySQLDialect) IntegerType() string {
	return "BIGINT"
}

func (d *MySQLDialect) TextType() string {
	return "TEXT"
}

func (d *MySQLDialect) BooleanType() string {
	return "TINYINT(1)" // MySQL uses TINYINT(1) for booleans
}

func (d *MySQLDialect) JSONType() string {
	return "JSON"
}

func (d *MySQLDialect) Placeholder(index int) string {
	return "?"
}

func (d *MySQLDialect) LimitOffset(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d *MySQLDialect) TransactionIsolation(level string) string {
	return fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
}

func (d *MySQLDialect) SupportsTransactions() bool {
	return true
}

func (d *MySQLDialect) SupportsRowLevelLocking() bool {
	return true // MySQL supports SELECT ... FOR UPDATE
}

func (d *MySQLDialect) EscapeIdentifier(name string) string {
	// MySQL uses backticks for identifiers
	return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
}
