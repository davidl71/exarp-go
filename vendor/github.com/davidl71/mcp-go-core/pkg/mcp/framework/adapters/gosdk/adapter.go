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
// Options can be provided to configure the adapter (e.g., WithLogger, WithMiddleware)
func NewGoSDKAdapter(name, version string, opts ...AdapterOption) *GoSDKAdapter {
	adapter := &GoSDKAdapter{
		server: mcp.NewServer(&mcp.Implementation{
			Name:    name,
			Version: version,
		}, nil),
		name:         name,
		toolHandlers: make(map[string]framework.ToolHandler),
		toolInfo:     make(map[string]types.ToolInfo),
	}

	// Apply options
	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// RegisterTool registers a tool with the server using the new v1.2.0 API
func (a *GoSDKAdapter) RegisterTool(name, description string, schema types.ToolSchema, handler framework.ToolHandler) error {
	// Input validation
	if err := ValidateRegistration(name, description, handler); err != nil {
		return fmt.Errorf("tool registration: %w", err)
	}
	if schema.Type == "" {
		schema.Type = "object" // Default to object type
	}
	if schema.Type != "object" {
		return fmt.Errorf("tool schema type must be 'object', got %q", schema.Type)
	}

	// Convert framework ToolSchema to go-sdk InputSchema
	// The schema must be a JSON object with type "object"
	inputSchemaMap := ToolSchemaToMCP(schema)

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
		if err := ValidateContext(ctx); err != nil {
			return nil, err
		}

		// Validate request
		if err := ValidateCallToolRequest(req); err != nil {
			return nil, err
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
		contents := TextContentToMCP(result)

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
	if err := ValidateRegistration(name, description, handler); err != nil {
		return fmt.Errorf("prompt registration: %w", err)
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
		if err := ValidateContext(ctx); err != nil {
			return nil, err
		}

		// Validate request
		if err := ValidateGetPromptRequest(req); err != nil {
			return nil, err
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
	if err := ValidateResourceRegistration(uri, name, description, handler); err != nil {
		return fmt.Errorf("resource registration: %w", err)
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
		if err := ValidateContext(ctx); err != nil {
			return nil, err
		}

		// Validate request
		if err := ValidateReadResourceRequest(req); err != nil {
			return nil, err
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
	if err := ValidateContext(ctx); err != nil {
		return err
	}

	// Validate server
	if a.server == nil {
		return fmt.Errorf("server is nil")
	}

	// Use provided transport or default to stdio
	if transport == nil {
		transport = &framework.StdioTransport{}
	}

	// Convert framework transport to go-sdk transport based on type
	var mcpTransport mcp.Transport
	switch transport.Type() {
	case "stdio":
		mcpTransport = &mcp.StdioTransport{}
	case "sse":
		// SSE transport would be implemented here when needed
		// For now, fall back to stdio
		return fmt.Errorf("SSE transport not yet implemented for go-sdk adapter")
	default:
		return fmt.Errorf("unsupported transport type: %s", transport.Type())
	}

	// Start the transport
	if err := transport.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Run the server with the transport
	if err := a.server.Run(ctx, mcpTransport); err != nil {
		// Try to stop transport on error
		_ = transport.Stop(ctx)
		return fmt.Errorf("server run failed: %w", err)
	}

	// Stop the transport when done
	if err := transport.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop transport: %w", err)
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
