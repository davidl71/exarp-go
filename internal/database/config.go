package database

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds database configuration
type Config struct {
	// Driver is the database driver type (sqlite, mysql, postgres, odbc)
	Driver DriverType

	// DSN is the database connection string
	// SQLite: file path or "file:path?mode=rwc"
	// MySQL: "user:password@tcp(host:port)/dbname?params"
	// PostgreSQL: "postgres://user:password@host:port/dbname?sslmode=disable"
	// ODBC: DSN name or connection string
	DSN string

	// AutoMigrate determines if migrations should run automatically
	AutoMigrate bool
}

// LoadConfig loads database configuration from environment variables
// with sensible defaults
func LoadConfig(projectRoot string) (*Config, error) {
	cfg := &Config{
		Driver:      DriverSQLite, // Default to SQLite
		AutoMigrate: true,
	}

	// Load driver from environment
	if driverStr := os.Getenv("DB_DRIVER"); driverStr != "" {
		driver := DriverType(driverStr)
		switch driver {
		case DriverSQLite, DriverMySQL, DriverPostgres, DriverODBC:
			cfg.Driver = driver
		default:
			return nil, fmt.Errorf("unsupported database driver: %s", driverStr)
		}
	}

	// Load DSN from environment or use default
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.DSN = dsn
	} else {
		// Default DSN based on driver
		switch cfg.Driver {
		case DriverSQLite:
			cfg.DSN = filepath.Join(projectRoot, ".todo2", "todo2.db")
		case DriverMySQL:
			cfg.DSN = "root:password@tcp(localhost:3306)/todo2?charset=utf8mb4&parseTime=True&loc=UTC"
		case DriverPostgres:
			cfg.DSN = "postgres://postgres:password@localhost:5432/todo2?sslmode=disable"
		case DriverODBC:
			cfg.DSN = "DSN=todo2"
		}
	}

	// Load auto-migrate setting
	if autoMigrate := os.Getenv("DB_AUTO_MIGRATE"); autoMigrate == "false" {
		cfg.AutoMigrate = false
	}

	return cfg, nil
}

// GetDefaultDSN returns the default DSN for a driver type
func GetDefaultDSN(driver DriverType, projectRoot string) string {
	switch driver {
	case DriverSQLite:
		return filepath.Join(projectRoot, ".todo2", "todo2.db")
	case DriverMySQL:
		return "root:password@tcp(localhost:3306)/todo2?charset=utf8mb4&parseTime=True&loc=UTC"
	case DriverPostgres:
		return "postgres://postgres:password@localhost:5432/todo2?sslmode=disable"
	case DriverODBC:
		return "DSN=todo2"
	default:
		return ""
	}
}
