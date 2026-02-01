package factory

import (
	"testing"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
)

func TestNewServer_GoSDK(t *testing.T) {
	server, err := NewServer(config.FrameworkGoSDK, "test-server", "1.0.0")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() server = nil")
	}

	if server.GetName() != "test-server" {
		t.Errorf("server.GetName() = %v, want test-server", server.GetName())
	}
}

func TestNewServer_UnknownFramework(t *testing.T) {
	server, err := NewServer("unknown-framework", "test-server", "1.0.0")
	if err == nil {
		t.Error("NewServer() error = nil, want error for unknown framework")
	}

	if server != nil {
		t.Errorf("NewServer() server = %v, want nil", server)
	}
}

func TestNewServerFromConfig(t *testing.T) {
	cfg := &config.Config{
		Framework: config.FrameworkGoSDK,
		Name:      "config-server",
		Version:   "1.0.0",
	}

	server, err := NewServerFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewServerFromConfig() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServerFromConfig() server = nil")
	}

	if server.GetName() != "config-server" {
		t.Errorf("server.GetName() = %v, want config-server", server.GetName())
	}
}

func TestNewServerFromConfig_InvalidFramework(t *testing.T) {
	cfg := &config.Config{
		Framework: "invalid",
		Name:      "test",
		Version:   "1.0.0",
	}

	server, err := NewServerFromConfig(cfg)
	if err == nil {
		t.Error("NewServerFromConfig() error = nil, want error")
	}

	if server != nil {
		t.Errorf("NewServerFromConfig() server = %v, want nil", server)
	}
}

func TestServer_ImplementsInterface(t *testing.T) {
	server, err := NewServer(config.FrameworkGoSDK, "test", "1.0.0")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Verify server implements MCPServer interface
	var _ framework.MCPServer = server
}

// TestNewServer_EmptyName tests server creation with empty name
func TestNewServer_EmptyName(t *testing.T) {
	server, err := NewServer(config.FrameworkGoSDK, "", "1.0.0")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() server = nil")
	}

	// Empty name should still create a server (gosdk adapter handles it)
	if server.GetName() != "" {
		t.Logf("server.GetName() = %v (gosdk may set default)", server.GetName())
	}
}

// TestNewServer_EmptyVersion tests server creation with empty version
func TestNewServer_EmptyVersion(t *testing.T) {
	server, err := NewServer(config.FrameworkGoSDK, "test-server", "")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() server = nil")
	}

	if server.GetName() != "test-server" {
		t.Errorf("server.GetName() = %v, want test-server", server.GetName())
	}
}

// TestNewServer_SpecialCharactersInName tests server with special characters
func TestNewServer_SpecialCharactersInName(t *testing.T) {
	specialNames := []string{
		"test-server-123",
		"test_server",
		"test.server",
		"test@server",
		"test/server",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			server, err := NewServer(config.FrameworkGoSDK, name, "1.0.0")
			if err != nil {
				t.Fatalf("NewServer() error = %v for name %q", err, name)
			}

			if server == nil {
				t.Fatalf("NewServer() server = nil for name %q", name)
			}

			if server.GetName() != name {
				t.Errorf("server.GetName() = %v, want %v", server.GetName(), name)
			}
		})
	}
}

// TestNewServer_VersionFormats tests various version formats
func TestNewServer_VersionFormats(t *testing.T) {
	versions := []string{
		"1.0.0",
		"1.0",
		"1",
		"v1.0.0",
		"1.0.0-beta",
		"1.0.0-rc.1",
		"1.0.0+build.123",
	}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			server, err := NewServer(config.FrameworkGoSDK, "test", version)
			if err != nil {
				t.Fatalf("NewServer() error = %v for version %q", err, version)
			}

			if server == nil {
				t.Fatalf("NewServer() server = nil for version %q", version)
			}
		})
	}
}

// TestNewServerFromConfig_NilConfig tests nil config handling
func TestNewServerFromConfig_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewServerFromConfig() did not panic with nil config")
		}
	}()

	_, _ = NewServerFromConfig(nil)
}

// TestNewServerFromConfig_EmptyConfig tests empty config fields
func TestNewServerFromConfig_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Framework: config.FrameworkGoSDK,
		Name:      "",
		Version:   "",
	}

	server, err := NewServerFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewServerFromConfig() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServerFromConfig() server = nil")
	}
}

// TestNewServer_CaseInsensitiveFramework tests framework type case handling
func TestNewServer_CaseInsensitiveFramework(t *testing.T) {
	// Test that framework type is case-sensitive (should fail)
	frameworks := []config.FrameworkType{
		"gosdk", // lowercase
		"GOSDK", // uppercase
		"GoSdk", // mixed case
	}

	for _, fw := range frameworks {
		t.Run(string(fw), func(t *testing.T) {
			server, err := NewServer(fw, "test", "1.0.0")
			if err == nil {
				t.Errorf("NewServer() expected error for framework %q, got server: %v", fw, server)
			}
			if server != nil {
				t.Errorf("NewServer() expected nil server for invalid framework %q", fw)
			}
		})
	}
}

// TestNewServer_MultipleInstances tests creating multiple server instances
func TestNewServer_MultipleInstances(t *testing.T) {
	servers := make([]framework.MCPServer, 3)
	names := []string{"server1", "server2", "server3"}

	for i, name := range names {
		server, err := NewServer(config.FrameworkGoSDK, name, "1.0.0")
		if err != nil {
			t.Fatalf("NewServer() error = %v for %s", err, name)
		}
		servers[i] = server
	}

	// Verify all servers are distinct and have correct names
	for i, server := range servers {
		if server == nil {
			t.Errorf("server[%d] is nil", i)
			continue
		}
		if server.GetName() != names[i] {
			t.Errorf("server[%d].GetName() = %v, want %v", i, server.GetName(), names[i])
		}
	}
}

// TestCreateLogger tests logger creation (internal function)
func TestCreateLogger(t *testing.T) {
	logger := createLogger()
	if logger == nil {
		t.Fatal("createLogger() returned nil")
	}

	// Logger should be configured with WARN level
	// (We can't directly test the level without exposing it, but we verify it doesn't panic)
	logger.Debug("test", "debug message")
	logger.Info("test", "info message")
	logger.Warn("test", "warn message")
}

// TestNewServerFromConfig_WithLoadedConfig tests config loaded from Load()
func TestNewServerFromConfig_WithLoadedConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	server, err := NewServerFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewServerFromConfig() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServerFromConfig() server = nil")
	}

	if server.GetName() != cfg.Name {
		t.Errorf("server.GetName() = %v, want %v", server.GetName(), cfg.Name)
	}
}
