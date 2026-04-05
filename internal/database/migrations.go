// migrations.go — Database schema migration management and execution.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RunMigrationsFromDir runs all pending migrations.
// When migrationsDirFromConfig is non-empty it must be a valid directory (EXARP_MIGRATIONS_DIR / cfg).
// When empty, migrations are resolved in order: EXARP_GO_ROOT/migrations, paths next to the binary,
// project <root>/migrations if present, else built-in embedded SQL from the binary.
func RunMigrationsFromDir(migrationsDirFromConfig string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	dir, useEmbed, _, err := ResolveMigrationsSource(migrationsDirFromConfig)
	if err != nil {
		return err
	}
	return runPendingMigrations(dir, useEmbed)
}

func runPendingMigrations(dir string, useEmbed bool) error {
	// Create schema_migrations table if it doesn't exist
	if err := createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	var migrations []Migration

	var err error

	if useEmbed {
		migrations, err = getMigrationFilesFromEmbed()
	} else {
		migrations, err = getMigrationFilesFromDir(dir)
	}

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

// Migration represents a migration file.
type Migration struct {
	Version     int
	Filename    string
	SQL         string
	Description string
}

// getMigrationFilesFromDir reads migration files from the given directory.
// Used by tests when the DB is in a temp dir that has no migrations (dir is repo migrations path).
func getMigrationFilesFromDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", dir, err)
	}

	var migrations []Migration

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
			continue
		}

		migrationPath := filepath.Join(dir, filename)

		sql, err := os.ReadFile(migrationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
		}

		description := migrationDescriptionFromParts(parts)
		migrations = append(migrations, Migration{
			Version:     version,
			Filename:    filename,
			SQL:         string(sql),
			Description: description,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func migrationDescriptionFromParts(parts []string) string {
	if len(parts) < 2 {
		return ""
	}
	description := strings.TrimSuffix(parts[1], ".sql")
	description = strings.ReplaceAll(description, "_", " ")
	return description
}

// getAppliedMigrations returns a map of applied migration versions.
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

// applyMigration applies a single migration.
// For migrations 002 and 003, runs each statement in its own transaction so that a
// "duplicate column" on one ALTER (e.g. version already in 001) does not stop later
// ALTERs (assignee, lock_until) from running. SQLite stops at the first error when
// Exec runs multiple statements, and a failed Exec can leave the tx unusable.
func applyMigration(migration Migration) error {
	runPerStatement := migration.Version == 2 || migration.Version == 3 || migration.Version == 6 || migration.Version == 8 || migration.Version == 11 || migration.Version == 18
	if runPerStatement {
		for _, stmt := range splitMigrationSQL(migration.SQL) {
			_, err := DB.Exec(stmt)
			if err != nil {
				errStr := strings.ToLower(err.Error())

				isDuplicateColumn := strings.Contains(errStr, "duplicate column") ||
					strings.Contains(errStr, "duplicate column name") ||
					strings.Contains(errStr, "already exists") ||
					(strings.Contains(errStr, "sql logic error") && strings.Contains(errStr, "duplicate"))
				if !isDuplicateColumn {
					return fmt.Errorf("failed to execute migration %d SQL: %w", migration.Version, err)
				}
				// Skip this statement, continue with next
			}
		}
		// Record migration in its own transaction
		_, err := DB.Exec(
			"INSERT OR IGNORE INTO schema_migrations (version, description) VALUES (?, ?)",
			migration.Version,
			migration.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	_, err = tx.Exec(
		"INSERT OR IGNORE INTO schema_migrations (version, description) VALUES (?, ?)",
		migration.Version,
		migration.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// splitMigrationSQL splits SQL into single statements, trims and skips empty/comment-only.
func splitMigrationSQL(sql string) []string {
	var out []string

	for _, s := range strings.Split(sql, ";") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// Skip blocks that are only comments and whitespace
		lines := strings.Split(s, "\n")

		var hasContent bool

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				hasContent = true
				break
			}
		}

		if hasContent {
			out = append(out, s)
		}
	}

	return out
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist.
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

// GetCurrentVersion returns the current schema version.
func GetCurrentVersion() (int, error) {
	var version int

	err := DB.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil // No migrations applied yet
	}

	return version, err
}
