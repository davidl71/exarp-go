package factory

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createLogger creates a logger with WARN as minimum level (INFO suppressed on stderr/stdout)
func createLogger() *logging.Logger {
	logger := logging.NewLogger()
	logger.SetLevel(logging.LevelWarn)
	return logger
}

// toolLoggingMiddleware returns a tool middleware that logs calls at debug level (T-274)
func toolLoggingMiddleware(logger *logging.Logger) func(gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := ""
			if req != nil && req.Params != nil {
				name = req.Params.Name
			}
			logger.Debug("", "Tool call: %s", name)
			return next(ctx, req)
		}
	}
}

// NewServer creates a new MCP server using the specified framework
func NewServer(frameworkType config.FrameworkType, name, version string) (framework.MCPServer, error) {
	switch frameworkType {
	case config.FrameworkGoSDK:
		logger := createLogger()
		// T-274: Add tool middleware (logging) - middleware chain applied in adapter
		return gosdk.NewGoSDKAdapter(name, version,
			gosdk.WithLogger(logger),
			gosdk.WithMiddleware(toolLoggingMiddleware(logger)),
		), nil
	default:
		return nil, fmt.Errorf("unknown framework: %s", frameworkType)
	}
}

// NewServerFromConfig creates server from configuration
func NewServerFromConfig(cfg *config.Config) (framework.MCPServer, error) {
	return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}
