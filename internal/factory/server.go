package factory

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

// createLogger creates a logger with WARN as minimum level (INFO suppressed on stderr/stdout)
func createLogger() *logging.Logger {
	logger := logging.NewLogger()
	logger.SetLevel(logging.LevelWarn)
	return logger
}

// NewServer creates a new MCP server using the specified framework
func NewServer(frameworkType config.FrameworkType, name, version string) (framework.MCPServer, error) {
	switch frameworkType {
	case config.FrameworkGoSDK:
		// Create logger with appropriate level for context
		logger := createLogger()
		return gosdk.NewGoSDKAdapter(name, version, gosdk.WithLogger(logger)), nil
	default:
		return nil, fmt.Errorf("unknown framework: %s", frameworkType)
	}
}

// NewServerFromConfig creates server from configuration
func NewServerFromConfig(cfg *config.Config) (framework.MCPServer, error) {
	return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}
