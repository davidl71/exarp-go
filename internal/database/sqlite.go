package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// DB is the global database connection
var DB *sql.DB

// currentDriver is the currently active database driver
var currentDriver Driver

// dbMutex protects the global DB variable from concurrent access
var dbMutex sync.Mutex

// Init initializes the database connection using the default SQLite driver
// This is kept for backward compatibility
// Thread-safe: uses mutex to prevent concurrent initialization
// Now uses centralized config system if available, falls back to legacy config
func Init(projectRoot string) error {
	// Try centralized config first
	cfg, err := LoadConfigFromCentralized(projectRoot)
	if err != nil {
		// Fall back to legacy config loading
		cfg, err = LoadConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		// Override to use SQLite for backward compatibility
		cfg.Driver = DriverSQLite
		cfg.DSN = filepath.Join(projectRoot, ".todo2", "todo2.db")
	}
	return InitWithConfig(cfg)
}

// InitWithConfig initializes the database connection using the provided configuration
// Supports multiple database backends (SQLite, MySQL, PostgreSQL, ODBC)
// Thread-safe: uses mutex to prevent concurrent initialization
func InitWithConfig(cfg *Config) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Close existing connection if any
	if DB != nil {
		DB.Close()
		DB = nil
		currentDriver = nil
	}

	// Get driver for the configured type
	driver, err := GetDriver(cfg.Driver)
	if err != nil {
		// Try to register missing drivers
		switch cfg.Driver {
		case DriverMySQL:
			RegisterDriver(NewMySQLDriver())
			driver, err = GetDriver(DriverMySQL)
		case DriverPostgres:
			RegisterDriver(NewPostgresDriver())
			driver, err = GetDriver(DriverPostgres)
		case DriverODBC:
			odbcDriver := NewODBCDriver()
			if odbcDriver == nil {
				return fmt.Errorf("ODBC driver requires cgo (CGO_ENABLED=1)")
			}
			RegisterDriver(odbcDriver)
			driver, err = GetDriver(DriverODBC)
		}
		if err != nil {
			return fmt.Errorf("failed to get driver %s: %w", cfg.Driver, err)
		}
	}

	// Open database connection
	db, err := driver.Open(cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure database (PRAGMA, SET, etc.)
	if err := driver.Configure(db); err != nil {
		db.Close()
		return fmt.Errorf("failed to configure database: %w", err)
	}

	DB = db
	currentDriver = driver

	// Run migrations if enabled
	if cfg.AutoMigrate {
		if err := RunMigrations(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
// Thread-safe: uses mutex to prevent concurrent close operations
func Close() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if DB != nil {
		err := DB.Close()
		DB = nil
		if currentDriver != nil {
			currentDriver.Close()
			currentDriver = nil
		}
		return err
	}
	return nil
}

// GetDB returns the global database connection
// Returns error if database is not initialized
// Note: Reading the DB pointer is safe without mutex (atomic pointer read)
// The mutex in Init()/Close() ensures proper initialization/cleanup
func GetDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized, call Init() first")
	}
	return DB, nil
}

// GetDriver returns the current database driver
// Returns error if database is not initialized
func GetCurrentDriver() (Driver, error) {
	if currentDriver == nil {
		return nil, fmt.Errorf("database not initialized, call Init() first")
	}
	return currentDriver, nil
}

// GetDialect returns the SQL dialect for the current database
// Returns error if database is not initialized
func GetDialect() (Dialect, error) {
	driver, err := GetCurrentDriver()
	if err != nil {
		return nil, err
	}
	return driver.Dialect(), nil
}
