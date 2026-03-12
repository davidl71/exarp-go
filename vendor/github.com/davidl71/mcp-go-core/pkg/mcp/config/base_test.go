package config

import (
	"os"
	"testing"
)

func TestLoadBaseConfig(t *testing.T) {
	// Save original environment variables
	originalFramework := os.Getenv("MCP_FRAMEWORK")
	originalName := os.Getenv("MCP_SERVER_NAME")
	originalVersion := os.Getenv("MCP_VERSION")
	defer func() {
		if originalFramework != "" {
			os.Setenv("MCP_FRAMEWORK", originalFramework)
		} else {
			os.Unsetenv("MCP_FRAMEWORK")
		}
		if originalName != "" {
			os.Setenv("MCP_SERVER_NAME", originalName)
		} else {
			os.Unsetenv("MCP_SERVER_NAME")
		}
		if originalVersion != "" {
			os.Setenv("MCP_VERSION", originalVersion)
		} else {
			os.Unsetenv("MCP_VERSION")
		}
	}()

	t.Run("defaults", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("MCP_FRAMEWORK")
		os.Unsetenv("MCP_SERVER_NAME")
		os.Unsetenv("MCP_VERSION")

		cfg, err := LoadBaseConfig()
		if err != nil {
			t.Fatalf("LoadBaseConfig() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("LoadBaseConfig() returned nil config")
		}
		if cfg.Framework != FrameworkGoSDK {
			t.Errorf("LoadBaseConfig().Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
		}
		if cfg.Name != "mcp-server" {
			t.Errorf("LoadBaseConfig().Name = %q, want %q", cfg.Name, "mcp-server")
		}
		if cfg.Version != "1.0.0" {
			t.Errorf("LoadBaseConfig().Version = %q, want %q", cfg.Version, "1.0.0")
		}
	})

	t.Run("environment override", func(t *testing.T) {
		// Set environment variables
		os.Setenv("MCP_FRAMEWORK", "go-sdk")
		os.Setenv("MCP_SERVER_NAME", "custom-server")
		os.Setenv("MCP_VERSION", "2.0.0")

		cfg, err := LoadBaseConfig()
		if err != nil {
			t.Fatalf("LoadBaseConfig() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("LoadBaseConfig() returned nil config")
		}
		if cfg.Framework != FrameworkGoSDK {
			t.Errorf("LoadBaseConfig().Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
		}
		if cfg.Name != "custom-server" {
			t.Errorf("LoadBaseConfig().Name = %q, want %q", cfg.Name, "custom-server")
		}
		if cfg.Version != "2.0.0" {
			t.Errorf("LoadBaseConfig().Version = %q, want %q", cfg.Version, "2.0.0")
		}
	})

	t.Run("unsupported framework", func(t *testing.T) {
		os.Setenv("MCP_FRAMEWORK", "unsupported-framework")
		defer os.Unsetenv("MCP_FRAMEWORK")

		cfg, err := LoadBaseConfig()
		if err == nil {
			t.Error("LoadBaseConfig() expected error for unsupported framework")
		}
		if cfg != nil {
			t.Error("LoadBaseConfig() should return nil config on error")
		}
	})

	t.Run("partial environment override", func(t *testing.T) {
		// Set only name, leave others as defaults
		os.Setenv("MCP_SERVER_NAME", "partial-override")
		os.Unsetenv("MCP_FRAMEWORK")
		os.Unsetenv("MCP_VERSION")

		cfg, err := LoadBaseConfig()
		if err != nil {
			t.Fatalf("LoadBaseConfig() error = %v", err)
		}
		if cfg.Name != "partial-override" {
			t.Errorf("LoadBaseConfig().Name = %q, want %q", cfg.Name, "partial-override")
		}
		if cfg.Framework != FrameworkGoSDK {
			t.Errorf("LoadBaseConfig().Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
		}
		if cfg.Version != "1.0.0" {
			t.Errorf("LoadBaseConfig().Version = %q, want %q", cfg.Version, "1.0.0")
		}
	})
}
