// Package archive provides retention and deletion for docs/archive per ARCHIVE_RETENTION_POLICY.md.
package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CategoryDeleteAfter maps docs/archive category dir names to "Delete After" dates (YYYY-MM-DD).
// Source: docs/archive/ARCHIVE_RETENTION_POLICY.md.
var CategoryDeleteAfter = map[string]string{
	"status-updates":           "2026-02-07",
	"implementation-summaries": "2026-04-07",
	"migration-planning":       "2026-07-07",
	"research-phase":           "2026-07-07",
	"analysis":                 "2027-01-07",
}

// ExpiredCategories returns category names whose Delete After date is on or before asOf.
func ExpiredCategories(asOf time.Time) []string {
	asOfDate := asOf.Truncate(24 * time.Hour)

	var out []string

	for category, deleteAfterStr := range CategoryDeleteAfter {
		t, err := time.Parse("2006-01-02", deleteAfterStr)
		if err != nil {
			continue
		}

		t = t.Truncate(24 * time.Hour)
		if !asOfDate.Before(t) {
			out = append(out, category)
		}
	}

	return out
}

// DeleteExpired lists or deletes files in docs/archive categories that are expired as of asOf.
// If dryRun is true, no files are removed; deleted paths are still returned as if removed.
// projectRoot is the repo root (must contain docs/archive/).
func DeleteExpired(projectRoot string, asOf time.Time, dryRun bool) (deleted []string, err error) {
	archiveRoot := filepath.Join(projectRoot, "docs", "archive")
	if _, statErr := os.Stat(archiveRoot); statErr != nil && os.IsNotExist(statErr) {
		return nil, fmt.Errorf("archive root does not exist: %s", archiveRoot)
	}

	expired := ExpiredCategories(asOf)
	for _, category := range expired {
		dir := filepath.Join(archiveRoot, category)

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}

			return deleted, fmt.Errorf("read dir %s: %w", dir, readErr)
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}

			path := filepath.Join(dir, e.Name())
			deleted = append(deleted, path)

			if !dryRun {
				if removeErr := os.Remove(path); removeErr != nil {
					return deleted, fmt.Errorf("remove %s: %w", path, removeErr)
				}
			}
		}
	}

	return deleted, nil
}
