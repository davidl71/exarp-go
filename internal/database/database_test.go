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
