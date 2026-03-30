package database

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSqliteOpenDSN_fileURIAndForeignKeysPragma(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "db.sqlite")
	got := sqliteOpenDSN(path)
	if !strings.HasPrefix(got, "file:") {
		t.Fatalf("want file: URI, got %q", got)
	}
	if !strings.Contains(got, "_pragma=foreign_keys(on)") {
		t.Fatalf("want foreign_keys pragma in DSN, got %q", got)
	}
}

func TestSqliteOpenDSN_preservesExplicitFileURI(t *testing.T) {
	t.Parallel()
	in := "file:/tmp/x.db?cache=shared"
	if sqliteOpenDSN(in) != in {
		t.Fatalf("expected unchanged %q", in)
	}
}
