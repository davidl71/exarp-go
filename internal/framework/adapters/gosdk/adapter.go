package gosdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GoSDKAdapter adapts the official Go SDK to the framework interface
type GoSDKAdapter struct {
	server       *mcp.Server
	name         string
	toolHandlers map[string]framework.ToolHandler
	toolInfo     map[string]types.ToolInfo
}

// NewGoSDKAdapter creates a new Go SDK adapter
func NewGoSDKAdapter(name, version string) *GoSDKAdapter {
	return &GoSDKAdapter{
		server: mcp.NewServer(&mcp.Implementation{
			Name:    name,
			Version: version,
		}, nil),
		name:         name,
		toolHandlers: make(map[string]framework.ToolHandler),
		toolInfo:     make(map[string]types.ToolInfo),
	}
}

// RegisterTool registers a tool with the server using the new v1.2.0 API
func (a *GoSDKAdapter) RegisterTool(name, description string, schema types.ToolSchema, handler framework.ToolHandler) error {
	// Input validation
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("tool description cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}
	if schema.Type == "" {
		schema.Type = "object" // Default to object type
	}
	if schema.Type != "object" {
		return fmt.Errorf("tool schema type must be 'object', got %q", schema.Type)
	}

	// Convert framework ToolSchema to go-sdk InputSchema
	// The schema must be a JSON object with type "object"
	inputSchemaMap := map[string]interface{}{
		"type":       schema.Type,
		"properties": schema.Properties,
	}
	if len(schema.Required) > 0 {
		inputSchemaMap["required"] = schema.Required
	}

	// Create tool definition with input schema
	tool := &mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchemaMap,
	}

	// Create handler function that matches ToolHandler signature
	// ToolHandler: func(context.Context, *CallToolRequest) (*CallToolResult, error)
	toolHandler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		// Validate request
		if req == nil {
			return nil, fmt.Errorf("call tool request cannot be nil")
		}
		if req.Params == nil {
			return nil, fmt.Errorf("call tool request params cannot be nil")
		}

		// Call framework handler with raw arguments
		result, err := handler(ctx, req.Params.Arguments)
		if err != nil {
			// Return error as tool error (not protocol error)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Tool execution error: %v", err),
					},
				},
			}, nil
		}

		// Validate result
		if result == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{},
			}, nil
		}

		// Convert framework TextContent to go-sdk Content
		contents := make([]mcp.Content, len(result))
		for i, content := range result {
			// TextContent is a struct, not a pointer, so we check for empty text
			if content.Text == "" {
				// Empty text is valid, but we'll include it anyway
			}
			contents[i] = &mcp.TextContent{
				Text: content.Text,
			}
		}

		return &mcp.CallToolResult{
			Content: contents,
		}, nil
	}

	// Use server.AddTool (low-level API) since we're using ToolHandler
	a.server.AddTool(tool, toolHandler)

	// Store handler and info for CLI access
	a.toolHandlers[name] = handler
		a.toolInfo[name] = types.ToolInfo{
		Name:        name,
		Description: description,
		Schema:      schema,
	}

	return nil
}

// RegisterPrompt registers a prompt with the server
func (a *GoSDKAdapter) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
	// Input validation
	if name == "" {
		return fmt.Errorf("prompt name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("prompt description cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("prompt handler cannot be nil")
	}

	// Create prompt definition
	prompt := &mcp.Prompt{
		Name:        name,
		Description: description,
	}

	// Create prompt handler that matches the new API
	// The new API uses: func(context.Context, *GetPromptRequest) (*GetPromptResult, error)
	promptHandler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		// Validate request
		if req == nil {
			return nil, fmt.Errorf("get prompt request cannot be nil")
		}
		if req.Params == nil {
			return nil, fmt.Errorf("get prompt request params cannot be nil")
		}

		// Convert req.Params.Arguments (map[string]any) to map[string]interface{}
		argsInterface := make(map[string]interface{})
		for k, v := range req.Params.Arguments {
			argsInterface[k] = v
		}

		// Call framework handler
		result, err := handler(ctx, argsInterface)
		if err != nil {
			return nil, fmt.Errorf("prompt handler failed: %w", err)
		}

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: result},
				},
			},
		}, nil
	}

	// Use server.AddPrompt with the new API
	a.server.AddPrompt(prompt, promptHandler)
	return nil
}

// RegisterResource registers a resource with the server
func (a *GoSDKAdapter) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
	// Input validation
	if uri == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("resource description cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("resource handler cannot be nil")
	}

	// Create resource definition
	resource := &mcp.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    mimeType,
	}

	// Create resource handler that matches the new API
	// The new API uses: func(context.Context, *ReadResourceRequest) (*ReadResourceResult, error)
	resourceHandler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		// Validate request
		if req == nil {
			return nil, fmt.Errorf("read resource request cannot be nil")
		}
		if req.Params == nil {
			return nil, fmt.Errorf("read resource request params cannot be nil")
		}
		if req.Params.URI == "" {
			return nil, fmt.Errorf("resource URI cannot be empty")
		}

		// Call framework handler with URI from params
		data, mimeType, err := handler(ctx, req.Params.URI)
		if err != nil {
			return nil, fmt.Errorf("resource handler failed for URI %q: %w", req.Params.URI, err)
		}

		// Validate data
		if data == nil {
			data = []byte{} // Empty data is valid
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: mimeType,
					Text:     string(data),
				},
			},
		}, nil
	}

	// Use server.AddResource with the new API
	a.server.AddResource(resource, resourceHandler)
	return nil
}

// Run starts the server with the given transport
func (a *GoSDKAdapter) Run(ctx context.Context, transport framework.Transport) error {
	// Check context cancellation
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate server
	if a.server == nil {
		return fmt.Errorf("server is nil")
	}

	// Use stdio transport for go-sdk
	// Note: transport parameter is ignored for now as go-sdk always uses StdioTransport
	// Future: implement transport type checking and conversion
	stdioTransport := &mcp.StdioTransport{}
	if err := a.server.Run(ctx, stdioTransport); err != nil {
		return fmt.Errorf("server run failed: %w", err)
	}
	return nil
}

// GetName returns the server name
func (a *GoSDKAdapter) GetName() string {
	return a.name
}

// CallTool executes a tool directly (for CLI mode)
func (a *GoSDKAdapter) CallTool(ctx context.Context, name string, args json.RawMessage) ([]types.TextContent, error) {
	handler, exists := a.toolHandlers[name]
	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return handler(ctx, args)
}

// ListTools returns all registered tools
func (a *GoSDKAdapter) ListTools() []types.ToolInfo {
	tools := make([]types.ToolInfo, 0, len(a.toolInfo))
	for _, info := range a.toolInfo {
		tools = append(tools, info)
	}
	return tools
}
