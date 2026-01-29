package gosdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GoSDKAdapter adapts the official Go SDK to the framework interface
type GoSDKAdapter struct {
	server       *mcp.Server
	name         string
	toolHandlers map[string]framework.ToolHandler // Pre-allocated map for O(1) lookups
	toolInfo     map[string]types.ToolInfo        // Pre-allocated map for O(1) lookups
	logger       *logging.Logger
	middleware   *MiddlewareChain
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
		logger:       logging.NewLogger(), // Default logger
		middleware:   NewMiddlewareChain(), // Default empty middleware chain
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

	a.logger.Debug("", "Registering tool: %s", name)

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

	a.logger.Info("", "Tool registered successfully: %s", name)
	return nil
}

// RegisterPrompt registers a prompt with the server
func (a *GoSDKAdapter) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
	a.logger.Debug("", "Registering prompt: %s", name)

	// Input validation
	if err := ValidateRegistration(name, description, handler); err != nil {
		return fmt.Errorf("prompt registration: %w", err)
	}

	// Create prompt definition
	prompt := &mcp.Prompt{
		Name:        name,
		Description: description,
	}

	// Create base prompt handler that matches the new API
	// The new API uses: func(context.Context, *GetPromptRequest) (*GetPromptResult, error)
	basePromptHandler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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

	// Wrap with middleware chain
	wrappedPromptHandler := a.middleware.WrapPromptHandler(basePromptHandler)

	// Convert PromptHandlerFunc to mcp.PromptHandler by wrapping (function signatures match)
	promptHandler := mcp.PromptHandler(func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return wrappedPromptHandler(ctx, req)
	})

	// Use server.AddPrompt with the new API
	a.server.AddPrompt(prompt, promptHandler)

	a.logger.Info("", "Prompt registered successfully: %s", name)
	return nil
}

// RegisterResource registers a resource with the server
func (a *GoSDKAdapter) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
	a.logger.Debug("", "Registering resource: %s", uri)

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

	// Create base resource handler that matches the new API
	// The new API uses: func(context.Context, *ReadResourceRequest) (*ReadResourceResult, error)
	baseResourceHandler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
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

	// Wrap with middleware chain
	wrappedResourceHandler := a.middleware.WrapResourceHandler(baseResourceHandler)

	// Convert ResourceHandlerFunc to mcp.ResourceHandler by wrapping (function signatures match)
	resourceHandler := mcp.ResourceHandler(func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return wrappedResourceHandler(ctx, req)
	})

	// Use server.AddResource with the new API
	a.server.AddResource(resource, resourceHandler)

	a.logger.Info("", "Resource registered successfully: %s", uri)
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
		// For SSE transport, we need to use the framework's SSETransport
		// The MCP SDK doesn't have a built-in SSE transport, so we'll use
		// the HTTP transport with SSE support
		// Note: The actual SSE handling is done by the framework transport
		// The adapter just needs to know that SSE is being used
		sseTransport, ok := transport.(*framework.SSETransport)
		if !ok {
			return fmt.Errorf("SSE transport must be of type *framework.SSETransport")
		}
		// The framework SSETransport manages the HTTP server
		// The MCP SDK will use stdio for now, but the framework transport
		// handles the SSE connection management
		// TODO: When MCP SDK adds SSE support, integrate it here
		a.logger.Warn("", "SSE transport: MCP SDK SSE support not yet available, using framework transport")
		// For now, we'll use stdio as a fallback, but the framework transport
		// will handle the actual SSE connections
		mcpTransport = &mcp.StdioTransport{}
		_ = sseTransport // Acknowledge SSE transport is provided
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
// Optimized for CLI usage with direct map lookup (O(1))
func (a *GoSDKAdapter) CallTool(ctx context.Context, name string, args json.RawMessage) ([]types.TextContent, error) {
	// Fast path: direct map lookup (O(1))
	handler, exists := a.toolHandlers[name]
	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return handler(ctx, args)
}

// ListTools returns all registered tools
// Optimized with pre-allocated slice capacity
func (a *GoSDKAdapter) ListTools() []types.ToolInfo {
	if len(a.toolInfo) == 0 {
		return nil // Return nil slice for empty (better than empty slice)
	}
	tools := make([]types.ToolInfo, 0, len(a.toolInfo))
	for _, info := range a.toolInfo {
		tools = append(tools, info)
	}
	return tools
}
