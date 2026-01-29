package database

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds database configuration
type Config struct {
	// Driver is the database driver type (sqlite, mysql, postgres)
	Driver DriverType

	// DSN is the database connection string
	// SQLite: file path or "file:path?mode=rwc"
	// MySQL: "user:password@tcp(host:port)/dbname?params"
	// PostgreSQL: "postgres://user:password@host:port/dbname?sslmode=disable"
	DSN string

	// AutoMigrate determines if migrations should run automatically
	AutoMigrate bool

	// MigrationsDir overrides the directory used to find migration files.
	// When set, migrations are loaded from this path instead of findProjectRoot()/migrations.
	// Used by tests when the DB is in a temp dir that has no migrations.
	MigrationsDir string
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
		case DriverSQLite, DriverMySQL, DriverPostgres:
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
	default:
		return ""
	}
}

// DatabaseConfigFields holds the database configuration fields from centralized config
// This struct is used to break the import cycle between database and config packages
type DatabaseConfigFields struct {
	SQLitePath          string
	JSONFallbackPath    string
	BackupPath          string
	MaxConnections      int
	ConnectionTimeout   int64 // Duration in seconds
	QueryTimeout        int64 // Duration in seconds
	RetryAttempts       int
	RetryInitialDelay   int64 // Duration in seconds
	RetryMaxDelay       int64 // Duration in seconds
	RetryMultiplier     float64
	AutoVacuum          bool
	WALMode             bool
	CheckpointInterval  int
	BackupRetentionDays int
}

// LoadConfigFromCentralizedFields loads database configuration from centralized config fields
// This allows database to use the centralized .exarp/config.yaml configuration
// without creating an import cycle (config package doesn't need to import database)
func LoadConfigFromCentralizedFields(projectRoot string, dbCfg DatabaseConfigFields) (*Config, error) {
	// Convert centralized DatabaseConfigFields to database.Config
	// Note: database.Config is simpler (Driver, DSN, AutoMigrate)
	// We use SQLite path from centralized config for DSN
	dbConfig := &Config{
		Driver:      DriverSQLite, // Default to SQLite (can be extended later)
		DSN:         filepath.Join(projectRoot, dbCfg.SQLitePath),
		AutoMigrate: true, // Default to true (can be made configurable later)
	}

	// Override with environment variables if set (environment takes precedence)
	if driverStr := os.Getenv("DB_DRIVER"); driverStr != "" {
		driver := DriverType(driverStr)
		switch driver {
		case DriverSQLite, DriverMySQL, DriverPostgres:
			dbConfig.Driver = driver
		default:
			return nil, fmt.Errorf("unsupported database driver: %s", driverStr)
		}
	}

	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		dbConfig.DSN = dsn
	} else if dbConfig.Driver == DriverSQLite {
		// Use centralized config path for SQLite
		if dbCfg.SQLitePath != "" {
			dbConfig.DSN = filepath.Join(projectRoot, dbCfg.SQLitePath)
		} else {
			// Fallback to default
			dbConfig.DSN = filepath.Join(projectRoot, ".todo2", "todo2.db")
		}
	}

	if autoMigrate := os.Getenv("DB_AUTO_MIGRATE"); autoMigrate == "false" {
		dbConfig.AutoMigrate = false
	}

	return dbConfig, nil
}
