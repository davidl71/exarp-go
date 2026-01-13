package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// DB is the global database connection
var DB *sql.DB

// dbMutex protects the global DB variable from concurrent access
var dbMutex sync.Mutex

// Init initializes the SQLite database connection
// It creates the database file if it doesn't exist and runs migrations
// Thread-safe: uses mutex to prevent concurrent initialization
func Init(projectRoot string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Close existing connection if any
	if DB != nil {
		DB.Close()
		DB = nil
	}

	dbPath := filepath.Join(projectRoot, ".todo2", "todo2.db")

	// Ensure .todo2 directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set SQLite PRAGMA settings (must be done outside transactions)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to set foreign_keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to set journal_mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return fmt.Errorf("failed to set synchronous: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 30000"); err != nil {
		return fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	// Set connection pool settings
	// SQLite with WAL mode supports multiple readers concurrently
	// Allow multiple connections for better performance with concurrent queries
	db.SetMaxOpenConns(10)   // Allow multiple readers (WAL mode supports this)
	db.SetMaxIdleConns(5)    // Keep some idle connections ready
	db.SetConnMaxLifetime(0) // Connections don't expire

	DB = db

	// Run migrations
	if err := RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
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
