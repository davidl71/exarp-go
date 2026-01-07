package framework

import (
	"context"
	"encoding/json"
)

// MCPServer abstracts MCP server functionality
type MCPServer interface {
	// RegisterTool registers a tool handler
	RegisterTool(name, description string, schema ToolSchema, handler ToolHandler) error

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
	CallTool(ctx context.Context, name string, args json.RawMessage) ([]TextContent, error)
	
	// ListTools returns all registered tools
	ListTools() []ToolInfo
}

// JsonRawMessage is an alias for json.RawMessage to avoid import conflicts
type JsonRawMessage = json.RawMessage

// ToolInfo represents tool metadata
type ToolInfo struct {
	Name        string
	Description string
	Schema      ToolSchema
}

// ToolHandler handles tool execution
type ToolHandler func(ctx context.Context, args json.RawMessage) ([]TextContent, error)

// PromptHandler handles prompt requests
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ResourceHandler handles resource requests
type ResourceHandler func(ctx context.Context, uri string) ([]byte, string, error)

// TextContent represents MCP text content
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolSchema represents tool input schema
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Transport abstracts transport mechanism
type Transport interface {
	// Transport-specific methods
	// Each framework will implement this differently
}

