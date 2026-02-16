package config

import (
	"testing"
	"time"
)

func TestGetGlobalConfig_DefaultsWhenNil(t *testing.T) {
	// Reset global config for test isolation
	configMu.Lock()
	orig := globalConfig
	globalConfig = nil
	configMu.Unlock()

	defer func() {
		configMu.Lock()
		globalConfig = orig
		configMu.Unlock()
	}()

	cfg := GetGlobalConfig()
	if cfg == nil {
		t.Fatal("GetGlobalConfig() returned nil")
	}

	if cfg.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", cfg.Version)
	}
}

func TestSetGlobalConfig_GetGlobalConfig(t *testing.T) {
	configMu.Lock()
	orig := globalConfig
	configMu.Unlock()

	defer func() {
		SetGlobalConfig(orig)
	}()

	cfg := GetDefaults()
	cfg.Version = "test-1.0"
	SetGlobalConfig(cfg)

	got := GetGlobalConfig()
	if got.Version != "test-1.0" {
		t.Errorf("GetGlobalConfig().Version = %q, want test-1.0", got.Version)
	}
}

func TestMemoryCategories(t *testing.T) {
	SetGlobalConfig(GetDefaults())

	cats := MemoryCategories()
	if len(cats) == 0 {
		t.Error("MemoryCategories() returned empty slice")
	}

	want := []string{"debug", "research", "architecture", "preference", "insight"}
	if len(cats) != len(want) {
		t.Errorf("MemoryCategories() length = %d, want %d", len(cats), len(want))
	}
}

func TestMemoryStoragePath(t *testing.T) {
	SetGlobalConfig(GetDefaults())

	path := MemoryStoragePath()
	if path != ".exarp/memories" {
		t.Errorf("MemoryStoragePath() = %q, want .exarp/memories", path)
	}
}

func TestProjectExarpPath(t *testing.T) {
	SetGlobalConfig(GetDefaults())

	path := ProjectExarpPath()
	if path != ".exarp" {
		t.Errorf("ProjectExarpPath() = %q, want .exarp", path)
	}
}

func TestWorkflowDefaultMode(t *testing.T) {
	SetGlobalConfig(GetDefaults())

	mode := WorkflowDefaultMode()
	if mode != "development" {
		t.Errorf("WorkflowDefaultMode() = %q, want development", mode)
	}
}

func TestDatabaseRetryGetters(t *testing.T) {
	SetGlobalConfig(GetDefaults())

	if DatabaseRetryAttempts() != 3 {
		t.Errorf("DatabaseRetryAttempts() = %d, want 3", DatabaseRetryAttempts())
	}

	if DatabaseConnectionTimeout() != 30*time.Second {
		t.Errorf("DatabaseConnectionTimeout() = %v, want 30s", DatabaseConnectionTimeout())
	}

	if DatabaseQueryTimeout() != 60*time.Second {
		t.Errorf("DatabaseQueryTimeout() = %v, want 60s", DatabaseQueryTimeout())
	}
}
