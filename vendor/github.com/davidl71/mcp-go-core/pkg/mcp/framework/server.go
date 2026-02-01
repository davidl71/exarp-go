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

// RootInfo describes a client root (workspace boundary) for path-sensitive handlers.
// When using an adapter that supports MCP Roots, resource handlers can call
// RootsFromContext(ctx) to get the list of roots the client has exposed.
type RootInfo struct {
	URI  string
	Name string
}

// rootsContextKey is the context key for client roots (MCP Roots feature).
type rootsContextKey struct{}

// RootsContextKey is the key used to attach []RootInfo to context in resource handlers.
// Adapters that support MCP Roots (e.g. gosdk) set this when the client provides roots.
var RootsContextKey = &rootsContextKey{}

// RootsFromContext returns the client roots from context, or nil if not set or unsupported.
func RootsFromContext(ctx context.Context) []RootInfo {
	v := ctx.Value(RootsContextKey)
	if v == nil {
		return nil
	}
	infos, ok := v.([]RootInfo)
	if !ok {
		return nil
	}
	return infos
}

// Eliciter allows tool handlers to request user input from the client (MCP Elicitation).
// When using an adapter that supports Elicitation, tool handlers can call EliciterFromContext(ctx)
// to get an Eliciter; if nil, the client does not support elicitation or the handler is not in MCP mode.
type Eliciter interface {
	// ElicitForm requests form-mode user input. message is shown to the user; schema is a JSON schema
	// for the form (e.g. map with "type":"object", "properties":{...}). Returns action ("accept"/"decline"/"cancel")
	// and content (form values when action is "accept"), or an error if the client does not support elicitation.
	ElicitForm(ctx context.Context, message string, schema map[string]interface{}) (action string, content map[string]interface{}, err error)
}

// eliciterContextKey is the context key for the Eliciter (MCP Elicitation feature).
type eliciterContextKey struct{}

// EliciterContextKey is the key used to attach an Eliciter to context in tool handlers.
var EliciterContextKey = &eliciterContextKey{}

// EliciterFromContext returns the Eliciter from context, or nil if not set or unsupported.
func EliciterFromContext(ctx context.Context) Eliciter {
	v := ctx.Value(EliciterContextKey)
	if v == nil {
		return nil
	}
	e, ok := v.(Eliciter)
	if !ok {
		return nil
	}
	return e
}

// Transport is defined in transport.go
// Imported here for backward compatibility
