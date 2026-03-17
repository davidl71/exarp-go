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
	"github.com/davidl71/exarp-go/internal/security"
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

// toolRateLimitMiddleware checks rate limits before executing a tool.
// Returns an error if the client has exceeded their rate limit.
func toolRateLimitMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		clientID := "default"
		if req != nil && req.Params != nil {
			if err := security.CheckRateLimit(clientID); err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("rate limit exceeded: %v", err)},
					},
				}, nil
			}
		}
		return next(ctx, req)
	}
}

// toolSemaphoreMiddleware limits concurrent tool executions using a semaphore.
// This prevents resource exhaustion from too many parallel tool calls.
func toolSemaphoreMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	semaphore := security.GetToolSemaphore(10)

	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if !semaphore.TryAcquire() {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "concurrent tool limit exceeded: please try again later"},
				},
			}, nil
		}
		defer semaphore.Release()

		return next(ctx, req)
	}
}

// toolAccessControlMiddleware checks access control before executing a tool.
// Returns an error if the tool is not allowed.
func toolAccessControlMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		toolName := ""
		if req != nil && req.Params != nil {
			toolName = req.Params.Name
		}
		if toolName != "" {
			if err := security.CheckToolAccess(toolName); err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("access denied: %v", err)},
					},
				}, nil
			}
		}
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
			gosdk.WithMiddleware(toolRateLimitMiddleware),
			gosdk.WithMiddleware(toolSemaphoreMiddleware),
			gosdk.WithMiddleware(toolAccessControlMiddleware),
			gosdk.WithMiddleware(toolContextCacheMiddleware),
			gosdk.WithMiddleware(toolLoggingMiddleware(logger)),
		}
		if cfg.hooks != nil {
			adapterOpts = append(adapterOpts, gosdk.WithMiddleware(toolHooksMiddleware(cfg.hooks)))
		}
		adapter := gosdk.NewGoSDKAdapter(name, version, adapterOpts...)
		wrapped := &exarpServer{MCPServer: adapter}

		if cfg.toolFilter != nil {
			return &filteredServer{MCPServer: wrapped, filter: cfg.toolFilter}, nil
		}
		return wrapped, nil
	default:
		return nil, fmt.Errorf("unknown framework: %s", frameworkType)
	}
}

// exarpServer wraps a GoSDKAdapter and adds ServerExtensionReporter so clients
// can discover which exarp-go capability extensions are enabled.
type exarpServer struct {
	framework.MCPServer
}

// ServerExtensions advertises the exarp-go MCP capability extensions.
func (s *exarpServer) ServerExtensions() map[string]any {
	return map[string]any{
		"davidl71/exarp-go": map[string]any{
			"projectRootContext":    true,
			"resourceTemplates":     true,
			"toolFiltering":         true,
			"resourceSubscriptions": true,
			"agentRunner":           true,
			"fmPlanExecute":         true,
		},
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

// NewServerFromConfig creates server from configuration with default options
// (workflow-mode tool filter enabled).
func NewServerFromConfig(cfg *config.Config, opts ...ServerOption) (framework.MCPServer, error) {
	return NewServer(cfg.Framework, cfg.Name, cfg.Version, opts...)
}

// UnwrapGoSDKServer unwraps the exarp-go server wrappers and returns the underlying
// go-sdk *mcp.Server. Returns nil if the server was not created from a GoSDKAdapter.
// Used by MCP Streamable HTTP mode to wire the real HTTP handler.
func UnwrapGoSDKServer(s framework.MCPServer) *mcp.Server {
	// Unwrap filteredServer → exarpServer → *gosdk.GoSDKAdapter
	type unwrapper interface{ Unwrap() framework.MCPServer }
	for {
		switch v := s.(type) {
		case *filteredServer:
			s = v.MCPServer
		case *exarpServer:
			s = v.MCPServer
		default:
			// Try the MCPServer() accessor we added to GoSDKAdapter.
			type mcpServerer interface{ MCPServer() *mcp.Server }
			if m, ok := s.(mcpServerer); ok {
				return m.MCPServer()
			}
			_ = unwrapper(nil) // satisfy compiler
			return nil
		}
	}
}
