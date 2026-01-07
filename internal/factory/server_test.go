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
