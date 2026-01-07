package factory

import (
	"fmt"

	"github.com/davidl/mcp-stdio-tools/internal/config"
	"github.com/davidl/mcp-stdio-tools/internal/framework"
	"github.com/davidl/mcp-stdio-tools/internal/framework/adapters/gosdk"
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
func NewServerFromConfig(cfg *config.Config) (framework.MCPServer, error) {
	return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}

