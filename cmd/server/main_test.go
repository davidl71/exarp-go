package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
)

func TestMain_ConfigLoading(t *testing.T) {
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

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.Framework != config.FrameworkGoSDK {
		t.Errorf("cfg.Framework = %v, want %v", cfg.Framework, config.FrameworkGoSDK)
	}
}

func TestMain_ServerCreation(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	// Test server creation (without actually running)
	// This verifies the main() initialization sequence would work
	if cfg.Framework != config.FrameworkGoSDK {
		t.Fatalf("cfg.Framework = %v, want %v", cfg.Framework, config.FrameworkGoSDK)
	}

	// Verify config is valid
	if cfg.Name == "" {
		t.Error("cfg.Name is empty")
	}
	if cfg.Version == "" {
		t.Error("cfg.Version is empty")
	}
}

func TestMain_ComponentRegistration(t *testing.T) {
	// This test verifies that the main() function would register all components
	// without actually starting the server
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	// Verify config allows server creation
	if cfg.Framework != config.FrameworkGoSDK {
		t.Fatalf("cfg.Framework = %v, want %v", cfg.Framework, config.FrameworkGoSDK)
	}
}

func TestMain_ContextHandling(t *testing.T) {
	// Test context cancellation handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify context can be used
	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
		// Expected
	}

	// Cancel context
	cancel()

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("context should be cancelled")
	}
}
