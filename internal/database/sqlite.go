// sqlite.go — SQLite connection pool, initialization, and global DB handle.
package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DB is the global database connection (sqlx wrapped for convenience).
var DB *sqlx.DB

// GetDB returns the underlying *sql.DB for backward compatibility.
// For new code, prefer using sqlx methods via GetDBx().
func GetDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized, call Init() first")
	}
	return DB.DB, nil
}

// GetDBx returns the global sqlx database connection.
func GetDBx() (*sqlx.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized, call Init() first")
	}
	return DB, nil
}

// currentDriver is the currently active database driver.
var currentDriver Driver

// dbMutex protects the global DB variable from concurrent access.
var dbMutex sync.Mutex

// initCfgFingerprint records the last successful InitWithConfig inputs. When a subsequent init
// matches, we skip Close/reopen so queue workers and hot paths do not churn the global pool.
var initCfgFingerprint struct {
	driver        DriverType
	dsn           string
	autoMigrate   bool
	migrationsDir string
}

// Init initializes the database connection using the default SQLite driver
// This is kept for backward compatibility
// Thread-safe: uses mutex to prevent concurrent initialization
// Now uses centralized config system if available, falls back to legacy config
// Note: To use centralized config, call InitWithCentralizedConfig instead.
func Init(projectRoot string) error {
	cfg, err := LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// Override to use SQLite for backward compatibility
	cfg.Driver = DriverSQLite
	cfg.DSN = filepath.Join(projectRoot, ".todo2", "todo2.db")

	return InitWithConfig(cfg)
}

// InitWithCentralizedConfig initializes the database connection using centralized config fields
// This allows using centralized config without creating an import cycle.
func InitWithCentralizedConfig(projectRoot string, dbCfg DatabaseConfigFields) error {
	cfg, err := LoadConfigFromCentralizedFields(projectRoot, dbCfg)
	if err != nil {
		// Fall back to legacy config loading
		return Init(projectRoot)
	}

	return InitWithConfig(cfg)
}

// InitWithConfig initializes the database connection using the provided configuration
// Supports SQLite (default), MySQL, and PostgreSQL backends
// Thread-safe: uses mutex to prevent concurrent initialization.
func InitWithConfig(cfg *Config) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if DB != nil &&
		cfg != nil &&
		cfg.Driver == initCfgFingerprint.driver &&
		cfg.DSN == initCfgFingerprint.dsn &&
		cfg.AutoMigrate == initCfgFingerprint.autoMigrate &&
		cfg.MigrationsDir == initCfgFingerprint.migrationsDir {
		return nil
	}

	// Close existing connection if any
	if DB != nil {
		DB.Close()
		DB = nil
		currentDriver = nil
	}

	// Get driver for the configured type
	driver, err := GetDriver(cfg.Driver)
	if err != nil {
		// Try to register missing drivers (lazy registration to avoid pulling deps when unused)
		switch cfg.Driver {
		case DriverMySQL:
			RegisterDriver(NewMySQLDriver())
			driver, err = GetDriver(DriverMySQL)
		case DriverPostgres:
			RegisterDriver(NewPostgresDriver())
			driver, err = GetDriver(DriverPostgres)
		case DriverRqlite:
			RegisterDriver(NewRqliteDriver())
			driver, err = GetDriver(DriverRqlite)
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

	DB = sqlx.NewDb(db, string(driver.Type()))
	currentDriver = driver
	initCfgFingerprint.driver = cfg.Driver
	initCfgFingerprint.dsn = cfg.DSN
	initCfgFingerprint.autoMigrate = cfg.AutoMigrate
	initCfgFingerprint.migrationsDir = cfg.MigrationsDir

	// Run migrations if enabled
	if cfg.AutoMigrate {
		if err := RunMigrationsFromDir(cfg.MigrationsDir); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
// Thread-safe: uses mutex to prevent concurrent close operations.
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

		initCfgFingerprint.driver = ""
		initCfgFingerprint.dsn = ""
		initCfgFingerprint.autoMigrate = false
		initCfgFingerprint.migrationsDir = ""

		return err
	}

	return nil
}
