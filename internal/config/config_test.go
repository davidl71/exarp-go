package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Save original env
	originalFramework := os.Getenv("MCP_FRAMEWORK")
	originalName := os.Getenv("MCP_SERVER_NAME")
	originalVersion := os.Getenv("MCP_VERSION")
	defer func() {
		os.Setenv("MCP_FRAMEWORK", originalFramework)
		os.Setenv("MCP_SERVER_NAME", originalName)
		os.Setenv("MCP_VERSION", originalVersion)
	}()

	// Clear env vars
	os.Unsetenv("MCP_FRAMEWORK")
	os.Unsetenv("MCP_SERVER_NAME")
	os.Unsetenv("MCP_VERSION")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Framework != FrameworkGoSDK {
		t.Errorf("cfg.Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
	}

	if cfg.Name != "exarp-go" {
		t.Errorf("cfg.Name = %v, want exarp-go", cfg.Name)
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("cfg.Version = %v, want 1.0.0", cfg.Version)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	// Save original env
	originalFramework := os.Getenv("MCP_FRAMEWORK")
	originalName := os.Getenv("MCP_SERVER_NAME")
	originalVersion := os.Getenv("MCP_VERSION")
	defer func() {
		os.Setenv("MCP_FRAMEWORK", originalFramework)
		os.Setenv("MCP_SERVER_NAME", originalName)
		os.Setenv("MCP_VERSION", originalVersion)
	}()

	// Set test env vars
	os.Setenv("MCP_FRAMEWORK", "go-sdk")
	os.Setenv("MCP_SERVER_NAME", "test-server")
	os.Setenv("MCP_VERSION", "2.0.0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Framework != FrameworkGoSDK {
		t.Errorf("cfg.Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
	}

	if cfg.Name != "test-server" {
		t.Errorf("cfg.Name = %v, want test-server", cfg.Name)
	}

	if cfg.Version != "2.0.0" {
		t.Errorf("cfg.Version = %v, want 2.0.0", cfg.Version)
	}
}

func TestLoad_UnsupportedFramework(t *testing.T) {
	// Save original env
	originalFramework := os.Getenv("MCP_FRAMEWORK")
	defer os.Setenv("MCP_FRAMEWORK", originalFramework)

	// Set unsupported framework
	os.Setenv("MCP_FRAMEWORK", "unsupported")

	cfg, err := Load()
	if err == nil {
		t.Error("Load() error = nil, want error for unsupported framework")
	}

	if cfg != nil {
		t.Errorf("Load() cfg = %v, want nil", cfg)
	}
}
