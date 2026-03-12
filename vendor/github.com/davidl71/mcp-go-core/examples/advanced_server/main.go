// Package main demonstrates an advanced MCP server with logging and middleware.
//
// This example shows how to:
//   - Create a server with custom logger
//   - Add middleware for request/response logging
//   - Use adapter options
//   - Handle errors gracefully
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/davidl71/mcp-go-core/pkg/mcp/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create custom logger with debug level
	logger := logging.NewLogger()
	logger.SetLevel(logging.LevelDebug)
	logger.Info("", "Starting advanced MCP server")

	// Create server with custom logger and middleware
	cfg, _ := config.LoadBaseConfig()
	cfg.Name = "advanced-server"
	cfg.Version = "1.0.0"

	// Create adapter with options
	adapter := gosdk.NewGoSDKAdapter(cfg.Name, cfg.Version,
		gosdk.WithLogger(logger),
		gosdk.WithMiddleware(&loggingMiddleware{logger: logger}),
	)

	// Register tools
	if err := registerAdvancedTools(adapter, logger); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Run server
	transport := &framework.StdioTransport{}
	if err := adapter.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// loggingMiddleware logs all tool calls, prompts, and resources
type loggingMiddleware struct {
	logger *logging.Logger
}

func (m *loggingMiddleware) ToolMiddleware(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()
		m.logger.Debug("", "Tool call started: %s", req.Params.Name)

		result, err := next(ctx, req)

		duration := time.Since(start)
		if err != nil {
			m.logger.Error("", "Tool call failed: %s (duration: %v): %v", req.Params.Name, duration, err)
		} else {
			m.logger.Info("", "Tool call completed: %s (duration: %v)", req.Params.Name, duration)
		}

		return result, err
	}
}

func (m *loggingMiddleware) PromptMiddleware(next gosdk.PromptHandlerFunc) gosdk.PromptHandlerFunc {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		m.logger.Debug("", "Prompt request: %s", req.Params.Name)
		result, err := next(ctx, req)
		if err != nil {
			m.logger.Error("", "Prompt request failed: %s: %v", req.Params.Name, err)
		}
		return result, err
	}
}

func (m *loggingMiddleware) ResourceMiddleware(next gosdk.ResourceHandlerFunc) gosdk.ResourceHandlerFunc {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		m.logger.Debug("", "Resource request: %s", req.Params.URI)
		result, err := next(ctx, req)
		if err != nil {
			m.logger.Error("", "Resource request failed: %s: %v", req.Params.URI, err)
		}
		return result, err
	}
}

func registerAdvancedTools(adapter *gosdk.GoSDKAdapter, logger *logging.Logger) error {
	// Register a tool with validation
	schema := types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"delay": map[string]interface{}{
				"type":        "number",
				"description": "Delay in seconds",
				"minimum":     0,
				"maximum":     60,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to return after delay",
			},
		},
		Required: []string{"delay", "message"},
	}

	handler := func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
		var params map[string]interface{}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		delay, _ := params["delay"].(float64)
		message, _ := params["message"].(string)

		// Simulate delay
		select {
		case <-time.After(time.Duration(delay) * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		logger.Info("", "Delayed tool completed after %v seconds", delay)

		return []types.TextContent{
			{Type: "text", Text: message},
		}, nil
	}

	return adapter.RegisterTool("delayed_echo", "Echo a message after a delay", schema, handler)
}
