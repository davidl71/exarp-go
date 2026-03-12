// Package config provides configuration management for MCP servers.
//
// This package includes base configuration structures, environment variable loading,
// and a builder pattern for programmatic configuration. The BaseConfig can be embedded
// by projects to add their own configuration fields.
//
// Example:
//
//	// Load from environment
//	cfg, err := config.LoadBaseConfig()
//
//	// Or use builder pattern
//	cfg, err := config.NewConfigBuilder().
//		WithName("my-server").
//		WithVersion("1.0.0").
//		Build()
package config

import (
	"fmt"
	"os"
)

// FrameworkType represents the type of MCP framework
type FrameworkType string

const (
	// FrameworkGoSDK is the official Go SDK
	FrameworkGoSDK FrameworkType = "go-sdk"
)

// BaseConfig holds the base server configuration
// Projects can embed this and add their own fields
type BaseConfig struct {
	Framework FrameworkType `yaml:"framework" env:"MCP_FRAMEWORK"`
	Name      string        `yaml:"name" env:"MCP_SERVER_NAME"`
	Version   string        `yaml:"version" env:"MCP_VERSION"`
}

// LoadBaseConfig loads base configuration from environment or defaults
func LoadBaseConfig() (*BaseConfig, error) {
	cfg := &BaseConfig{
		Framework: FrameworkGoSDK, // Default to go-sdk
		Name:      "mcp-server",   // Default name (projects should override)
		Version:   "1.0.0",        // Default version
	}

	// Override from environment
	if frameworkStr := os.Getenv("MCP_FRAMEWORK"); frameworkStr != "" {
		cfg.Framework = FrameworkType(frameworkStr)
	}
	if name := os.Getenv("MCP_SERVER_NAME"); name != "" {
		cfg.Name = name
	}
	if version := os.Getenv("MCP_VERSION"); version != "" {
		cfg.Version = version
	}

	// Validate framework
	if cfg.Framework != FrameworkGoSDK {
		return nil, fmt.Errorf("unsupported framework: %s", cfg.Framework)
	}

	return cfg, nil
}
