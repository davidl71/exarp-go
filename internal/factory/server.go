// server.go — MCP server factory: creates and configures framework instances.
//
// Package factory provides MCP server framework instantiation from configuration.
package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
	"github.com/lawlielt/ctxcache"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createLogger creates a logger with WARN as minimum level (INFO suppressed on stderr/stdout).
func createLogger() *logging.Logger {
	logger := logging.NewLogger()
	logger.SetLevel(logging.LevelWarn)

	return logger
}

// toolLoggingMiddleware returns a tool middleware that logs calls at debug level (T-274).
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

// toolContextCacheMiddleware wraps each tool request context with a request-scoped cache (ctxcache)
// so handlers can use ctxcache.Get/Set for per-request memoization without cross-request bleed.
func toolContextCacheMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = ctxcache.NewContextWithCache(ctx)
		return next(ctx, req)
	}
}

// toolRecoveryMiddleware catches panics in tool handlers and returns a clean MCP error
// instead of crashing the server process. Should be registered first in the middleware chain.
func toolRecoveryMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(ctx context.Context, req *mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				name := ""
				if req != nil && req.Params != nil {
					name = req.Params.Name
				}
				slog.Error("panic recovered in tool handler",
					"tool", name,
					"panic", fmt.Sprintf("%v", r),
					"stack", string(debug.Stack()))
				result = &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("internal error in tool %s: panic recovered", name)},
					},
				}
				err = nil
			}
		}()
		return next(ctx, req)
	}
}

// toolHooksMiddleware runs before/after callbacks around each tool invocation.
func toolHooksMiddleware(hooks *framework.Hooks) func(gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := ""
			var rawArgs json.RawMessage
			if req != nil && req.Params != nil {
				name = req.Params.Name
				rawArgs = req.Params.Arguments
			}
			if hooks.BeforeToolCall != nil {
				hooks.BeforeToolCall(ctx, name, rawArgs)
			}
			result, err := next(ctx, req)
			if hooks.AfterToolCall != nil {
				hooks.AfterToolCall(ctx, name, rawArgs)
			}
			return result, err
		}
	}
}

// ServerOption configures NewServer behaviour.
type ServerOption func(*serverConfig)

type serverConfig struct {
	hooks      *framework.Hooks
	toolFilter framework.ToolFilterFunc
}

// WithHooks adds before/after tool call hooks.
func WithHooks(hooks *framework.Hooks) ServerOption {
	return func(c *serverConfig) { c.hooks = hooks }
}

// WithToolFilter adds per-session tool filtering.
func WithToolFilter(fn framework.ToolFilterFunc) ServerOption {
	return func(c *serverConfig) { c.toolFilter = fn }
}

// NewServer creates a new MCP server using the specified framework.
func NewServer(frameworkType config.FrameworkType, name, version string, opts ...ServerOption) (framework.MCPServer, error) {
	var cfg serverConfig
	for _, o := range opts {
		o(&cfg)
	}

	switch frameworkType {
	case config.FrameworkGoSDK:
		logger := createLogger()
		adapterOpts := []gosdk.AdapterOption{
			gosdk.WithLogger(logger),
			gosdk.WithMiddleware(toolRecoveryMiddleware),
			gosdk.WithMiddleware(toolContextCacheMiddleware),
			gosdk.WithMiddleware(toolLoggingMiddleware(logger)),
		}
		if cfg.hooks != nil {
			adapterOpts = append(adapterOpts, gosdk.WithMiddleware(toolHooksMiddleware(cfg.hooks)))
		}
		adapter := gosdk.NewGoSDKAdapter(name, version, adapterOpts...)
		if cfg.toolFilter != nil {
			return &filteredServer{MCPServer: adapter, filter: cfg.toolFilter}, nil
		}
		return adapter, nil
	default:
		return nil, fmt.Errorf("unknown framework: %s", frameworkType)
	}
}

// filteredServer wraps an MCPServer and applies a ToolFilterFunc to ListTools.
type filteredServer struct {
	framework.MCPServer
	filter framework.ToolFilterFunc
}

// ListTools applies the tool filter to the inner server's tool list.
func (f *filteredServer) ListTools() []framework.ToolInfo {
	return f.filter(context.Background(), f.MCPServer.ListTools())
}

// NewServerFromConfig creates server from configuration.
func NewServerFromConfig(cfg *config.Config) (framework.MCPServer, error) {
	return NewServer(cfg.Framework, cfg.Name, cfg.Version)
}
