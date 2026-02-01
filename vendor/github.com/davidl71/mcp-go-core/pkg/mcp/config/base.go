// Package config provides base MCP server configuration and a builder for programmatic construction.
package config

import (
	"fmt"
	"os"
)

// BaseConfig holds the minimal MCP server configuration (framework, name, version).
// Compatible with exarp-go internal/config.Config for optional adoption.
type BaseConfig struct {
	Framework string
	Name      string
	Version   string
}

// DefaultFramework is the default MCP framework identifier (e.g. "go-sdk").
const DefaultFramework = "go-sdk"

// DefaultName is the default server name when not set via env.
const DefaultName = "mcp-server"

// DefaultVersion is the default server version when not set via env.
const DefaultVersion = "1.0.0"

// LoadBaseConfig loads base configuration from environment and applies defaults.
// Environment variables: MCP_FRAMEWORK, MCP_SERVER_NAME, MCP_VERSION.
func LoadBaseConfig() (*BaseConfig, error) {
	cfg := &BaseConfig{
		Framework: DefaultFramework,
		Name:      DefaultName,
		Version:   DefaultVersion,
	}
	if v := os.Getenv("MCP_FRAMEWORK"); v != "" {
		cfg.Framework = v
	}
	if v := os.Getenv("MCP_SERVER_NAME"); v != "" {
		cfg.Name = v
	}
	if v := os.Getenv("MCP_VERSION"); v != "" {
		cfg.Version = v
	}
	if err := validateBaseConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validateBaseConfig(c *BaseConfig) error {
	if c.Framework == "" {
		return fmt.Errorf("framework cannot be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if c.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	return nil
}
