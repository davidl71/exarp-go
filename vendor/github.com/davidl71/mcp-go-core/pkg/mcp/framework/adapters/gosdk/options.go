package gosdk

import "github.com/davidl71/mcp-go-core/pkg/mcp/logging"

// AdapterOption configures a GoSDKAdapter
type AdapterOption func(*GoSDKAdapter)

// WithLogger sets a custom logger for the adapter
// If not provided, a default logger will be created.
// The logger is used for tool registration, tool calls, and other adapter operations.
func WithLogger(logger *logging.Logger) AdapterOption {
	return func(a *GoSDKAdapter) {
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithMiddleware adds middleware to the adapter
// Middleware can be provided as:
//   - A Middleware interface (applies to all handler types)
//   - Individual middleware functions for tools, prompts, or resources
//
// Example:
//
//	// Using Middleware interface
//	adapter := NewGoSDKAdapter("server", "1.0.0",
//		WithMiddleware(myMiddleware),
//	)
//
//	// Using individual middleware functions
//	adapter := NewGoSDKAdapter("server", "1.0.0",
//		WithMiddleware(func(chain *MiddlewareChain) {
//			chain.AddToolMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
//				return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//					// Do something before
//					result, err := next(ctx, req)
//					// Do something after
//					return result, err
//				}
//			})
//		}),
//	)
func WithMiddleware(middleware interface{}) AdapterOption {
	return func(a *GoSDKAdapter) {
		if middleware == nil {
			return
		}

		// Check if it's a Middleware interface
		if mw, ok := middleware.(Middleware); ok {
			a.middleware.ApplyMiddleware(mw)
			return
		}

		// Check if it's a function that configures the middleware chain
		if configFunc, ok := middleware.(func(*MiddlewareChain)); ok {
			configFunc(a.middleware)
			return
		}

		// If it's a single tool middleware function
		if toolMw, ok := middleware.(func(ToolHandlerFunc) ToolHandlerFunc); ok {
			a.middleware.AddToolMiddleware(toolMw)
			return
		}

		// If it's a single prompt middleware function
		if promptMw, ok := middleware.(func(PromptHandlerFunc) PromptHandlerFunc); ok {
			a.middleware.AddPromptMiddleware(promptMw)
			return
		}

		// If it's a single resource middleware function
		if resourceMw, ok := middleware.(func(ResourceHandlerFunc) ResourceHandlerFunc); ok {
			a.middleware.AddResourceMiddleware(resourceMw)
			return
		}
	}
}
