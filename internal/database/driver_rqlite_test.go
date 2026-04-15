package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRqliteDriver_Type(t *testing.T) {
	d := NewRqliteDriver()
	if d.Type() != DriverRqlite {
		t.Errorf("Type() = %v, want %v", d.Type(), DriverRqlite)
	}
}

func TestRqliteDialect_Placeholder(t *testing.T) {
	d := NewRqliteDialect()
	if got := d.Placeholder(1); got != "?" {
		t.Errorf("Placeholder(1) = %q, want ?", got)
	}
}

func TestGetDefaultDSN_Rqlite(t *testing.T) {
	projectRoot := "/tmp/proj"
	got := GetDefaultDSN(DriverRqlite, projectRoot)
	if got != "http://localhost:4001" {
		t.Errorf("GetDefaultDSN(DriverRqlite, %q) = %q, want http://localhost:4001", projectRoot, got)
	}
}

func TestLoadConfig_AcceptsRqlite(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("DB_DRIVER", "rqlite")
	os.Setenv("DB_DSN", "http://localhost:4001")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_DSN")
	}()

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Driver != DriverRqlite {
		t.Errorf("Driver = %v, want rqlite", cfg.Driver)
	}
	if cfg.DSN != "http://localhost:4001" {
		t.Errorf("DSN = %q, want http://localhost:4001", cfg.DSN)
	}
}

// TestRqliteDriver_Open_SkipsWhenNoServer tries to open rqlite; skips if no server (e.g. connection refused).
func TestRqliteDriver_Open_SkipsWhenNoServer(t *testing.T) {
	d := NewRqliteDriver()
	db, err := d.Open("http://localhost:4001")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "refused") {
			t.Skip("rqlite not running at localhost:4001:", err)
		}
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("rqlite ping failed:", err)
	}
	// If we get here, rqlite is running; no further assertion needed
}

func TestListDrivers_IncludesRqliteAfterRegister(t *testing.T) {
	// Rqlite is registered on first use (InitWithConfig); ensure we can register and list
	RegisterDriver(NewRqliteDriver())
	list := ListDrivers()
	hasRqlite := false
	for _, dt := range list {
		if dt == DriverRqlite {
			hasRqlite = true
			break
		}
	}
	if !hasRqlite {
		t.Errorf("ListDrivers() = %v, expected to include rqlite after Register", list)
	}
}

func TestLoadConfig_RqliteDefaultDSN(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("DB_DRIVER", "rqlite")
	defer os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DB_DSN")

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Driver != DriverRqlite {
		t.Fatalf("Driver = %v", cfg.Driver)
	}
	// Default DSN when DB_DSN not set
	want := "http://localhost:4001"
	if cfg.DSN != want {
		t.Errorf("DSN = %q, want %q", cfg.DSN, want)
	}
}

func TestGetDefaultDSN_AllDrivers(t *testing.T) {
	projectRoot := filepath.Join(t.TempDir(), "proj")
	tests := []struct {
		driver DriverType
		want   string
	}{
		{DriverSQLite, filepath.Join(projectRoot, ".todo2", "todo2.db")},
		{DriverRqlite, "http://localhost:4001"},
	}
	for _, tt := range tests {
		got := GetDefaultDSN(tt.driver, projectRoot)
		if got != tt.want {
			t.Errorf("GetDefaultDSN(%v) = %q, want %q", tt.driver, got, tt.want)
		}
	}
}
