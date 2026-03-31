// migrations_resolve.go — Resolve migration SQL source for dev trees, installed binaries, and client projects.
package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/projectroot"
	appmigrations "github.com/davidl71/exarp-go/migrations"
)

// ResolveMigrationsSource picks where to load schema migrations from.
//
// Order when userOverride is empty:
//  1. EXARP_GO_ROOT/migrations (development / sibling clone)
//  2. Next to os.Executable: ../share/exarp-go/migrations, then migrations/
//  3. PROJECT_ROOT/migrations via projectroot.Find() if it contains migration files (no mkdir)
//  4. Built-in embedded *.sql from the exarp-go build (go install / MCP from other repos)
//
// When userOverride is non-empty (EXARP_MIGRATIONS_DIR / cfg): path must exist, be a directory,
// and contain at least one numbered *.sql file. Returns an error otherwise (strict).
//
// fingerprint is stored in initCfgFingerprint so migration source changes trigger DB re-init.
func ResolveMigrationsSource(userOverride string) (dir string, useEmbed bool, fingerprint string, err error) {
	if strings.TrimSpace(userOverride) != "" {
		abs, err := filepath.Abs(filepath.Clean(userOverride))
		if err != nil {
			return "", false, "", fmt.Errorf("EXARP_MIGRATIONS_DIR: %w", err)
		}
		fi, err := os.Stat(abs)
		if err != nil {
			return "", false, "", fmt.Errorf("EXARP_MIGRATIONS_DIR=%q: %w", userOverride, err)
		}
		if !fi.IsDir() {
			return "", false, "", fmt.Errorf("EXARP_MIGRATIONS_DIR=%q: not a directory", abs)
		}
		if !dirHasNumberedMigrationSQL(abs) {
			return "", false, "", fmt.Errorf("EXARP_MIGRATIONS_DIR=%q: no numbered *.sql migrations (expected names like 001_initial.sql)", abs)
		}
		return abs, false, abs, nil
	}

	if root := strings.TrimSpace(os.Getenv("EXARP_GO_ROOT")); root != "" {
		candidate := filepath.Join(filepath.Clean(root), "migrations")
		if dirHasNumberedMigrationSQL(candidate) {
			abs, _ := filepath.Abs(candidate)
			return abs, false, abs, nil
		}
	}

	if exe, exeErr := os.Executable(); exeErr == nil && exe != "" {
		exeDir := filepath.Dir(exe)
		for _, rel := range []string{
			filepath.Join("..", "share", "exarp-go", "migrations"),
			"migrations",
		} {
			candidate := filepath.Clean(filepath.Join(exeDir, rel))
			if dirHasNumberedMigrationSQL(candidate) {
				abs, _ := filepath.Abs(candidate)
				return abs, false, abs, nil
			}
		}
	}

	if projectRoot, rootErr := projectroot.Find(); rootErr == nil && projectRoot != "" {
		candidate := filepath.Join(projectRoot, "migrations")
		if dirHasNumberedMigrationSQL(candidate) {
			abs, _ := filepath.Abs(candidate)
			return abs, false, abs, nil
		}
	}

	// Built-in copy shipped with the binary.
	if err := assertEmbeddedMigrationsNonEmpty(); err != nil {
		return "", false, "", err
	}
	return "", true, ":embedded:", nil
}

func assertEmbeddedMigrationsNonEmpty() error {
	entries, err := appmigrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("embedded migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 1 {
			continue
		}
		if _, err := strconv.Atoi(parts[0]); err == nil {
			return nil
		}
	}
	return fmt.Errorf("embedded migrations: no numbered *.sql files (rebuild exarp-go)")
}

func dirHasNumberedMigrationSQL(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 1 {
			continue
		}
		if _, err := strconv.Atoi(parts[0]); err == nil {
			return true
		}
	}
	return false
}

func getMigrationFilesFromEmbed() ([]Migration, error) {
	entries, err := appmigrations.Files.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
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
		data, err := fs.ReadFile(appmigrations.Files, filename)
		if err != nil {
			return nil, fmt.Errorf("read embedded migration %s: %w", filename, err)
		}
		description := migrationDescriptionFromParts(parts)
		migrations = append(migrations, Migration{
			Version:     version,
			Filename:    filename,
			SQL:         string(data),
			Description: description,
		})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

// resolvedMigrationsFingerprintForInit returns a stable fingerprint of the migration source when
// AutoMigrate is enabled; empty when migrations are not run (skips filesystem/embed work).
func resolvedMigrationsFingerprintForInit(cfg *Config) (string, error) {
	if cfg == nil || !cfg.AutoMigrate {
		return "", nil
	}
	_, _, fp, err := ResolveMigrationsSource(cfg.MigrationsDir)
	if err != nil {
		return "", fmt.Errorf("migrations: %w", err)
	}
	return fp, nil
}
