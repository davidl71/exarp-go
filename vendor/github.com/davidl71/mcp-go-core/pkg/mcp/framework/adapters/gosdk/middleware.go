package gosdk

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Middleware represents a middleware function that can intercept and modify
// tool calls, prompt requests, and resource requests.
//
// Middleware functions are called in the order they are registered.
// Each middleware can:
//   - Modify the request context
//   - Validate or transform request parameters
//   - Log or monitor the request
//   - Modify the response
//   - Short-circuit the request (return error)
//
// Example usage:
//
//	func LoggingMiddleware(next ToolHandler) ToolHandler {
//		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//			log.Printf("Tool call: %s", req.Params.Name)
//			return next(ctx, req)
//		}
//	}
type Middleware interface {
	// ToolMiddleware wraps a tool handler
	ToolMiddleware(next ToolHandlerFunc) ToolHandlerFunc

	// PromptMiddleware wraps a prompt handler
	PromptMiddleware(next PromptHandlerFunc) PromptHandlerFunc

	// ResourceMiddleware wraps a resource handler
	ResourceMiddleware(next ResourceHandlerFunc) ResourceHandlerFunc
}

// ToolHandlerFunc is the function signature for tool handlers in middleware chain
type ToolHandlerFunc func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error)

// PromptHandlerFunc is the function signature for prompt handlers in middleware chain
type PromptHandlerFunc func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error)

// ResourceHandlerFunc is the function signature for resource handlers in middleware chain
type ResourceHandlerFunc func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error)

// MiddlewareChain manages a chain of middleware functions
type MiddlewareChain struct {
	toolMiddlewares     []func(ToolHandlerFunc) ToolHandlerFunc
	promptMiddlewares   []func(PromptHandlerFunc) PromptHandlerFunc
	resourceMiddlewares []func(ResourceHandlerFunc) ResourceHandlerFunc
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		toolMiddlewares:     make([]func(ToolHandlerFunc) ToolHandlerFunc, 0),
		promptMiddlewares:   make([]func(PromptHandlerFunc) PromptHandlerFunc, 0),
		resourceMiddlewares: make([]func(ResourceHandlerFunc) ResourceHandlerFunc, 0),
	}
}

// AddToolMiddleware adds a middleware function for tool calls
func (mc *MiddlewareChain) AddToolMiddleware(mw func(ToolHandlerFunc) ToolHandlerFunc) {
	mc.toolMiddlewares = append(mc.toolMiddlewares, mw)
}

// AddPromptMiddleware adds a middleware function for prompt requests
func (mc *MiddlewareChain) AddPromptMiddleware(mw func(PromptHandlerFunc) PromptHandlerFunc) {
	mc.promptMiddlewares = append(mc.promptMiddlewares, mw)
}

// AddResourceMiddleware adds a middleware function for resource requests
func (mc *MiddlewareChain) AddResourceMiddleware(mw func(ResourceHandlerFunc) ResourceHandlerFunc) {
	mc.resourceMiddlewares = append(mc.resourceMiddlewares, mw)
}

// WrapToolHandler wraps a tool handler with all registered middleware
// Optimized: skip wrapping if no middleware registered
func (mc *MiddlewareChain) WrapToolHandler(handler ToolHandlerFunc) ToolHandlerFunc {
	if len(mc.toolMiddlewares) == 0 {
		return handler // Fast path: no middleware
	}
	// Apply middleware in reverse order (last registered wraps first)
	wrapped := handler
	for i := len(mc.toolMiddlewares) - 1; i >= 0; i-- {
		wrapped = mc.toolMiddlewares[i](wrapped)
	}
	return wrapped
}

// WrapPromptHandler wraps a prompt handler with all registered middleware
// Optimized: skip wrapping if no middleware registered
func (mc *MiddlewareChain) WrapPromptHandler(handler PromptHandlerFunc) PromptHandlerFunc {
	if len(mc.promptMiddlewares) == 0 {
		return handler // Fast path: no middleware
	}
	// Apply middleware in reverse order (last registered wraps first)
	wrapped := handler
	for i := len(mc.promptMiddlewares) - 1; i >= 0; i-- {
		wrapped = mc.promptMiddlewares[i](wrapped)
	}
	return wrapped
}

// WrapResourceHandler wraps a resource handler with all registered middleware
// Optimized: skip wrapping if no middleware registered
func (mc *MiddlewareChain) WrapResourceHandler(handler ResourceHandlerFunc) ResourceHandlerFunc {
	if len(mc.resourceMiddlewares) == 0 {
		return handler // Fast path: no middleware
	}
	// Apply middleware in reverse order (last registered wraps first)
	wrapped := handler
	for i := len(mc.resourceMiddlewares) - 1; i >= 0; i-- {
		wrapped = mc.resourceMiddlewares[i](wrapped)
	}
	return wrapped
}

// ApplyMiddleware applies a full Middleware interface to the chain
func (mc *MiddlewareChain) ApplyMiddleware(mw Middleware) {
	if mw == nil {
		return
	}
	mc.AddToolMiddleware(mw.ToolMiddleware)
	mc.AddPromptMiddleware(mw.PromptMiddleware)
	mc.AddResourceMiddleware(mw.ResourceMiddleware)
}
