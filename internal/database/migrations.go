package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RunMigrations runs all pending migrations
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create schema_migrations table if it doesn't exist
	if err := createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue // Already applied
		}

		if err := applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// Migration represents a migration file
type Migration struct {
	Version     int
	Filename    string
	SQL         string
	Description string
}

// getMigrationFiles reads all migration files from the migrations directory
func getMigrationFiles() ([]Migration, error) {
	var migrations []Migration

	// Find project root (look for .todo2 directory)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Read migration directory
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filename := entry.Name()
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 1 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Skip files that don't start with a number
		}

		// Read migration SQL
		migrationPath := filepath.Join(migrationsDir, filename)
		sql, err := os.ReadFile(migrationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
		}

		// Extract description from filename
		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		migrations = append(migrations, Migration{
			Version:     version,
			Filename:    filename,
			SQL:         string(sql),
			Description: description,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// findProjectRoot finds the project root by looking for .todo2 directory
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		todo2Path := filepath.Join(dir, ".todo2")
		if _, err := os.Stat(todo2Path); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (no .todo2 directory)")
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := DB.Query("SELECT version FROM schema_migrations")
	if err != nil {
		// Table might not exist yet, return empty map
		return applied, nil
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func applyMigration(migration Migration) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
		migration.Version,
		migration.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func createMigrationsTable() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			description TEXT
		)
	`)
	return err
}

// GetCurrentVersion returns the current schema version
func GetCurrentVersion() (int, error) {
	var version int
	err := DB.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil // No migrations applied yet
	}
	return version, err
}
