package factory

import (
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

// isGitHookContext detects if we're running in a git hook context
// Checks for GIT_HOOK env var or non-interactive mode (no TTY)
func isGitHookContext() bool {
	// Check explicit GIT_HOOK env var
	if os.Getenv("GIT_HOOK") == "1" || os.Getenv("GIT_HOOK") == "true" {
		return true
	}

	// Check if we're in a non-interactive context (git hooks typically aren't)
	// This is a heuristic - git hooks usually don't have a TTY
	// But we only use this as a fallback since explicit env var is more reliable
	return false // Don't auto-detect from TTY to avoid false positives
}

// createLogger creates a logger with appropriate level based on context
// Suppresses INFO logs except for MCP-related operations (tool/prompt/resource registration)
// MCP registration logs (tool/prompt/resource) are already MCP-related and will show at INFO level
// Other INFO logs from non-MCP sources (like database initialization) use standard log package
func createLogger() *logging.Logger {
	logger := logging.NewLogger()

	// Keep INFO level to show MCP registration logs
	// The MCP registration logs ("Tool registered successfully", etc.) are MCP-related
	// and will show at INFO level. Other INFO logs from standard log package are separate.
	
	// In git hook contexts, suppress all INFO messages (including MCP)
	if isGitHookContext() {
		logger.SetLevel(logging.LevelWarn)
	}

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
