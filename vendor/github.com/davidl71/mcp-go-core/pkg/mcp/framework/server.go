// Package framework provides framework-agnostic abstractions for MCP servers.
//
// The framework package defines interfaces and adapters that allow MCP servers
// to work with different underlying frameworks (go-sdk, mcp-go, etc.) without
// changing application code.
//
// Example:
//
//	server, err := factory.NewServer(config.FrameworkGoSDK, "my-server", "1.0.0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	server.RegisterTool("my_tool", "Description", schema, handler)
//	server.Run(ctx, transport)
package framework

import (
	"context"
	"encoding/json"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// MCPServer abstracts MCP server functionality
type MCPServer interface {
	// RegisterTool registers a tool handler
	RegisterTool(name, description string, schema types.ToolSchema, handler ToolHandler) error

	// RegisterPrompt registers a prompt template
	RegisterPrompt(name, description string, handler PromptHandler) error

	// RegisterResource registers a resource handler
	RegisterResource(uri, name, description, mimeType string, handler ResourceHandler) error

	// Run starts the server with the given transport
	Run(ctx context.Context, transport Transport) error

	// GetName returns the server name
	GetName() string

	// CLI support methods
	// CallTool executes a tool directly (for CLI mode)
	CallTool(ctx context.Context, name string, args json.RawMessage) ([]types.TextContent, error)

	// ListTools returns all registered tools
	ListTools() []types.ToolInfo
}

// JsonRawMessage is an alias for json.RawMessage to avoid import conflicts
type JsonRawMessage = json.RawMessage

// ToolHandler handles tool execution
type ToolHandler func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error)

// PromptHandler handles prompt requests
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ResourceHandler handles resource requests
type ResourceHandler func(ctx context.Context, uri string) ([]byte, string, error)

// Transport is defined in transport.go
// Imported here for backward compatibility
