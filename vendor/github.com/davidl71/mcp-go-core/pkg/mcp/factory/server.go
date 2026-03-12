// Package factory provides factory functions for creating MCP servers.
//
// The factory package centralizes server creation logic and enables
// configuration-driven server creation. It supports different framework types
// and provides a consistent API for server instantiation.
//
// Example:
//
//	server, err := factory.NewServer(config.FrameworkGoSDK, "my-server", "1.0.0")
//	// or
//	cfg, _ := config.LoadBaseConfig()
//	server, err := factory.NewServerFromConfig(cfg)
package factory

import (
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
)

// NewServer creates a new MCP server using the specified framework
func NewServer(frameworkType config.FrameworkType, name, version string) (framework.MCPServer, error) {
	switch frameworkType {
	case config.FrameworkGoSDK:
		return gosdk.NewGoSDKAdapter(name, version), nil
	default:
		return nil, fmt.Errorf("unknown framework: %s", frameworkType)
	}
}

// NewServerFromConfig creates server from configuration
func NewServerFromConfig(cfg *config.BaseConfig) (framework.MCPServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}
