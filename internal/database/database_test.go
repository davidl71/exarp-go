package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	projectRoot := tmpDir

	// Initialize database
	err := Init(projectRoot)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify database file was created
	dbPath := filepath.Join(projectRoot, ".todo2", "todo2.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file not created at %s", dbPath)
	}

	// Verify connection works
	db, err := GetDB()
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}

	if db == nil {
		t.Fatal("GetDB() returned nil")
	}

	// Verify migrations ran
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion() error = %v", err)
	}

	if version < 1 {
		t.Errorf("Expected schema version >= 1, got %d", version)
	}

	// Cleanup
	if err := Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestInit_IdempotentSameConfig(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	tmpDir := t.TempDir()
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	db1, err := GetDBx()
	if err != nil {
		t.Fatalf("GetDBx() error = %v", err)
	}
	if err := Init(tmpDir); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
	db2, err := GetDBx()
	if err != nil {
		t.Fatalf("GetDBx() after second Init error = %v", err)
	}
	if db1 != db2 {
		t.Errorf("expected same *sqlx.DB after idempotent Init, got different pointers")
	}
	if err := Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestInit_SwitchesWhenProjectRootChanges(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	tmpA := t.TempDir()
	tmpB := t.TempDir()
	if err := Init(tmpA); err != nil {
		t.Fatalf("Init(A) error = %v", err)
	}
	pathA := filepath.Join(tmpA, ".todo2", "todo2.db")
	if err := Init(tmpB); err != nil {
		t.Fatalf("Init(B) error = %v", err)
	}
	pathB := filepath.Join(tmpB, ".todo2", "todo2.db")
	if _, err := os.Stat(pathA); err != nil {
		t.Errorf("A database: %v", err)
	}
	if _, err := os.Stat(pathB); err != nil {
		t.Errorf("B database: %v", err)
	}
	vB, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion after B: %v", err)
	}
	if vB < 1 {
		t.Errorf("expected B schema version >= 1, got %d", vB)
	}
	if err := Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
