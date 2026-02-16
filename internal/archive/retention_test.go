package archive

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExpiredCategories(t *testing.T) {
	parse := func(s string) time.Time {
		t.Helper()

		tt, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}

		return tt
	}

	tests := []struct {
		asOf    string
		wantAny []string // at least these should be expired
	}{
		{"2026-02-06", nil},
		{"2026-02-07", []string{"status-updates"}},
		{"2026-02-08", []string{"status-updates"}},
		{"2026-04-06", []string{"status-updates"}},
		{"2026-04-07", []string{"status-updates", "implementation-summaries"}},
		{"2026-07-07", []string{"status-updates", "implementation-summaries", "migration-planning", "research-phase"}},
		{"2027-01-07", []string{"status-updates", "implementation-summaries", "migration-planning", "research-phase", "analysis"}},
	}
	for _, tt := range tests {
		t.Run(tt.asOf, func(t *testing.T) {
			got := ExpiredCategories(parse(tt.asOf))

			for _, want := range tt.wantAny {
				found := false

				for _, c := range got {
					if c == want {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("ExpiredCategories(%s) = %v, want to include %q", tt.asOf, got, want)
				}
			}
		})
	}
}

func TestDeleteExpired(t *testing.T) {
	tmp := t.TempDir()
	archiveRoot := filepath.Join(tmp, "docs", "archive")

	statusDir := filepath.Join(archiveRoot, "status-updates")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		t.Fatal(err)
	}

	f1 := filepath.Join(statusDir, "a.md")
	f2 := filepath.Join(statusDir, "b.md")

	for _, p := range []string{f1, f2} {
		if err := os.WriteFile(p, []byte("# test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// As of 2026-02-08, status-updates is expired
	asOf, _ := time.Parse("2006-01-02", "2026-02-08")

	deleted, err := DeleteExpired(tmp, asOf, true)
	if err != nil {
		t.Fatalf("DeleteExpired(dryRun=true): %v", err)
	}

	if len(deleted) != 2 {
		t.Errorf("DeleteExpired(dryRun=true) returned %d paths, want 2: %v", len(deleted), deleted)
	}
	// Files should still exist (dry run)
	for _, p := range []string{f1, f2} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("dry run should not remove %s", p)
		}
	}

	// Actually delete
	deleted2, err := DeleteExpired(tmp, asOf, false)
	if err != nil {
		t.Fatalf("DeleteExpired(dryRun=false): %v", err)
	}

	if len(deleted2) != 2 {
		t.Errorf("DeleteExpired(dryRun=false) returned %d paths, want 2", len(deleted2))
	}

	for _, p := range []string{f1, f2} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("file should have been removed: %s", p)
		}
	}
}

func TestDeleteExpired_NoArchiveRoot(t *testing.T) {
	tmp := t.TempDir()
	asOf, _ := time.Parse("2006-01-02", "2026-02-08")

	_, err := DeleteExpired(tmp, asOf, true)
	if err == nil {
		t.Error("DeleteExpired with no docs/archive should error")
	}
}
