package config

import (
	"os"
	"testing"
)

func TestConfigBuilder_Build(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg, err := NewConfigBuilder().Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if cfg.Framework != DefaultFramework {
			t.Errorf("Framework = %q, want %q", cfg.Framework, DefaultFramework)
		}
		if cfg.Name != DefaultName {
			t.Errorf("Name = %q, want %q", cfg.Name, DefaultName)
		}
		if cfg.Version != DefaultVersion {
			t.Errorf("Version = %q, want %q", cfg.Version, DefaultVersion)
		}
	})

	t.Run("WithFramework", func(t *testing.T) {
		cfg, err := NewConfigBuilder().WithFramework("go-sdk").Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if cfg.Framework != "go-sdk" {
			t.Errorf("Framework = %q, want go-sdk", cfg.Framework)
		}
	})

	t.Run("WithName", func(t *testing.T) {
		cfg, err := NewConfigBuilder().WithName("my-server").Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if cfg.Name != "my-server" {
			t.Errorf("Name = %q, want my-server", cfg.Name)
		}
	})

	t.Run("WithVersion", func(t *testing.T) {
		cfg, err := NewConfigBuilder().WithVersion("2.0.0").Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if cfg.Version != "2.0.0" {
			t.Errorf("Version = %q, want 2.0.0", cfg.Version)
		}
	})

	t.Run("fluent chain", func(t *testing.T) {
		cfg, err := NewConfigBuilder().
			WithFramework("go-sdk").
			WithName("exarp-go").
			WithVersion("1.0.0").
			Build()
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if cfg.Framework != "go-sdk" || cfg.Name != "exarp-go" || cfg.Version != "1.0.0" {
			t.Errorf("got Framework=%q Name=%q Version=%q", cfg.Framework, cfg.Name, cfg.Version)
		}
	})

	t.Run("empty framework invalid", func(t *testing.T) {
		_, err := NewConfigBuilder().WithFramework("").Build()
		if err == nil {
			t.Error("Build() with empty framework want error")
		}
	})

	t.Run("empty name invalid", func(t *testing.T) {
		_, err := NewConfigBuilder().WithName("").Build()
		if err == nil {
			t.Error("Build() with empty name want error")
		}
	})

	t.Run("empty version invalid", func(t *testing.T) {
		_, err := NewConfigBuilder().WithVersion("").Build()
		if err == nil {
			t.Error("Build() with empty version want error")
		}
	})
}

func TestLoadBaseConfig(t *testing.T) {
	t.Run("defaults when no env", func(t *testing.T) {
		// Clear env for this test
		os.Unsetenv("MCP_FRAMEWORK")
		os.Unsetenv("MCP_SERVER_NAME")
		os.Unsetenv("MCP_VERSION")
		cfg, err := LoadBaseConfig()
		if err != nil {
			t.Fatalf("LoadBaseConfig() error = %v", err)
		}
		if cfg.Framework != DefaultFramework {
			t.Errorf("Framework = %q, want %q", cfg.Framework, DefaultFramework)
		}
		if cfg.Name != DefaultName {
			t.Errorf("Name = %q, want %q", cfg.Name, DefaultName)
		}
		if cfg.Version != DefaultVersion {
			t.Errorf("Version = %q, want %q", cfg.Version, DefaultVersion)
		}
	})

	t.Run("env overrides", func(t *testing.T) {
		os.Setenv("MCP_FRAMEWORK", "go-sdk")
		os.Setenv("MCP_SERVER_NAME", "test-server")
		os.Setenv("MCP_VERSION", "9.9.9")
		defer func() {
			os.Unsetenv("MCP_FRAMEWORK")
			os.Unsetenv("MCP_SERVER_NAME")
			os.Unsetenv("MCP_VERSION")
		}()
		cfg, err := LoadBaseConfig()
		if err != nil {
			t.Fatalf("LoadBaseConfig() error = %v", err)
		}
		if cfg.Framework != "go-sdk" {
			t.Errorf("Framework = %q, want go-sdk", cfg.Framework)
		}
		if cfg.Name != "test-server" {
			t.Errorf("Name = %q, want test-server", cfg.Name)
		}
		if cfg.Version != "9.9.9" {
			t.Errorf("Version = %q, want 9.9.9", cfg.Version)
		}
	})
}
