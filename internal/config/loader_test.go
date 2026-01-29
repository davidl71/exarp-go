package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Test loading config when no file exists (should use defaults)
	cfg, err := LoadConfig("/tmp/nonexistent")
	if err != nil {
		t.Fatalf("LoadConfig should not fail when file doesn't exist: %v", err)
	}

	// Verify defaults are used
	if cfg.Timeouts.TaskLockLease != 30*time.Minute {
		t.Errorf("Expected TaskLockLease %v, got %v", 30*time.Minute, cfg.Timeouts.TaskLockLease)
	}
	if cfg.Thresholds.SimilarityThreshold != 0.85 {
		t.Errorf("Expected SimilarityThreshold 0.85, got %f", cfg.Thresholds.SimilarityThreshold)
	}
	if cfg.Tasks.DefaultStatus != "Todo" {
		t.Errorf("Expected DefaultStatus 'Todo', got '%s'", cfg.Tasks.DefaultStatus)
	}
}

func TestLoadConfig_OnlyYAMLReturnsError(t *testing.T) {
	// Protobuf mandatory: if only config.yaml exists, LoadConfig must error
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	yamlPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte("version: \"1.0\"\n"), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	_, err := LoadConfig(tmpDir)
	if err == nil {
		t.Fatal("LoadConfig must fail when only config.yaml exists (protobuf mandatory)")
	}
	if !strings.Contains(err.Error(), "protobuf") {
		t.Errorf("error should mention protobuf: %v", err)
	}
}

func TestLoadConfig_WithFile(t *testing.T) {
	// Create temporary directory and write config as protobuf (mandatory format)
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	fileConfig := GetDefaults()
	fileConfig.Version = "1.0"
	fileConfig.Timeouts.TaskLockLease = 45 * time.Minute
	fileConfig.Timeouts.ToolDefault = 120 * time.Second
	fileConfig.Thresholds.SimilarityThreshold = 0.9
	fileConfig.Thresholds.MinCoverage = 85
	fileConfig.Tasks.DefaultStatus = "In Progress"
	fileConfig.Tasks.StaleThresholdHours = 4
	if err := WriteConfigToProtobufFile(tmpDir, fileConfig); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify custom values are used
	if cfg.Timeouts.TaskLockLease != 45*time.Minute {
		t.Errorf("Expected TaskLockLease %v, got %v", 45*time.Minute, cfg.Timeouts.TaskLockLease)
	}
	if cfg.Timeouts.ToolDefault != 120*time.Second {
		t.Errorf("Expected ToolDefault %v, got %v", 120*time.Second, cfg.Timeouts.ToolDefault)
	}
	if cfg.Thresholds.SimilarityThreshold != 0.9 {
		t.Errorf("Expected SimilarityThreshold 0.9, got %f", cfg.Thresholds.SimilarityThreshold)
	}
	if cfg.Thresholds.MinCoverage != 85 {
		t.Errorf("Expected MinCoverage 85, got %d", cfg.Thresholds.MinCoverage)
	}
	if cfg.Tasks.DefaultStatus != "In Progress" {
		t.Errorf("Expected DefaultStatus 'In Progress', got '%s'", cfg.Tasks.DefaultStatus)
	}
	if cfg.Tasks.StaleThresholdHours != 4 {
		t.Errorf("Expected StaleThresholdHours 4, got %d", cfg.Tasks.StaleThresholdHours)
	}

	// Verify defaults are still used for unspecified values
	if cfg.Thresholds.MinTaskConfidence != 0.7 {
		t.Errorf("Expected MinTaskConfidence 0.7 (default), got %f", cfg.Thresholds.MinTaskConfidence)
	}
}

func TestLoadConfig_MergeDefaults(t *testing.T) {
	// Create temporary directory and write partial config as protobuf
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	fileConfig := GetDefaults()
	fileConfig.Version = "1.0"
	fileConfig.Timeouts.TaskLockLease = 45 * time.Minute
	fileConfig.Thresholds.SimilarityThreshold = 0.9
	if err := WriteConfigToProtobufFile(tmpDir, fileConfig); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify custom values
	if cfg.Timeouts.TaskLockLease != 45*time.Minute {
		t.Errorf("Expected TaskLockLease %v, got %v", 45*time.Minute, cfg.Timeouts.TaskLockLease)
	}
	if cfg.Thresholds.SimilarityThreshold != 0.9 {
		t.Errorf("Expected SimilarityThreshold 0.9, got %f", cfg.Thresholds.SimilarityThreshold)
	}

	// Verify defaults are still used for unspecified values
	if cfg.Timeouts.ToolDefault != 60*time.Second {
		t.Errorf("Expected ToolDefault %v (default), got %v", 60*time.Second, cfg.Timeouts.ToolDefault)
	}
	if cfg.Tasks.DefaultStatus != "Todo" {
		t.Errorf("Expected DefaultStatus 'Todo' (default), got '%s'", cfg.Tasks.DefaultStatus)
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	exarpDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(exarpDir, 0755); err != nil {
		t.Fatalf("Failed to create .exarp directory: %v", err)
	}

	// Change to subdirectory
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Find project root
	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot failed: %v", err)
	}

	// Should find the directory with .exarp (normalize paths for macOS /var vs /private/var)
	rootNorm, _ := filepath.EvalSymlinks(root)
	tmpNorm, _ := filepath.EvalSymlinks(tmpDir)
	if rootNorm != tmpNorm {
		t.Errorf("Expected project root %s (normalized %s), got %s (normalized %s)", tmpDir, tmpNorm, root, rootNorm)
	}
}

func TestGetGlobalConfig(t *testing.T) {
	// Initially should return defaults
	cfg := GetGlobalConfig()
	if cfg.Timeouts.TaskLockLease != 30*time.Minute {
		t.Errorf("Expected default TaskLockLease %v, got %v", 30*time.Minute, cfg.Timeouts.TaskLockLease)
	}

	// Set custom config
	customCfg := GetDefaults()
	customCfg.Timeouts.TaskLockLease = 45 * time.Minute
	SetGlobalConfig(customCfg)

	// Should return custom config
	cfg = GetGlobalConfig()
	if cfg.Timeouts.TaskLockLease != 45*time.Minute {
		t.Errorf("Expected custom TaskLockLease %v, got %v", 45*time.Minute, cfg.Timeouts.TaskLockLease)
	}
}

func TestGetters(t *testing.T) {
	// Set global config
	cfg := GetDefaults()
	cfg.Timeouts.TaskLockLease = 45 * time.Minute
	cfg.Thresholds.SimilarityThreshold = 0.9
	cfg.Tasks.DefaultStatus = "In Progress"
	SetGlobalConfig(cfg)

	// Test getters
	if TaskLockLease() != 45*time.Minute {
		t.Errorf("TaskLockLease() = %v, want %v", TaskLockLease(), 45*time.Minute)
	}
	if SimilarityThreshold() != 0.9 {
		t.Errorf("SimilarityThreshold() = %f, want 0.9", SimilarityThreshold())
	}
	if DefaultTaskStatus() != "In Progress" {
		t.Errorf("DefaultTaskStatus() = %s, want 'In Progress'", DefaultTaskStatus())
	}

	// Test tool timeout getter
	if ToolTimeout("scorecard") != 60*time.Second {
		t.Errorf("ToolTimeout('scorecard') = %v, want %v", ToolTimeout("scorecard"), 60*time.Second)
	}
}
