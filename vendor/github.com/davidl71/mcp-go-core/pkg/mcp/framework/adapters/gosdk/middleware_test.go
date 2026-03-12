package gosdk

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMiddlewareChain_AddToolMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()

	callOrder := []string{}

	// Add middleware in order
	chain.AddToolMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			callOrder = append(callOrder, "mw1-before")
			result, err := next(ctx, req)
			callOrder = append(callOrder, "mw1-after")
			return result, err
		}
	})

	chain.AddToolMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			callOrder = append(callOrder, "mw2-before")
			result, err := next(ctx, req)
			callOrder = append(callOrder, "mw2-after")
			return result, err
		}
	})

	// Create handler
	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		callOrder = append(callOrder, "handler")
		return &mcp.CallToolResult{}, nil
	}

	// Wrap handler
	wrapped := chain.WrapToolHandler(handler)

	// Call wrapped handler
	_, _ = wrapped(context.Background(), &mcp.CallToolRequest{})

	// Verify call order: mw1-before, mw2-before, handler, mw2-after, mw1-after
	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(callOrder) != len(expected) {
		t.Errorf("Call order length = %d, want %d", len(callOrder), len(expected))
	}
	for i, want := range expected {
		if i >= len(callOrder) || callOrder[i] != want {
			t.Errorf("Call order[%d] = %q, want %q", i, callOrder[i], want)
		}
	}
}

func TestMiddlewareChain_AddPromptMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()

	callOrder := []string{}

	chain.AddPromptMiddleware(func(next PromptHandlerFunc) PromptHandlerFunc {
		return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			callOrder = append(callOrder, "mw-before")
			result, err := next(ctx, req)
			callOrder = append(callOrder, "mw-after")
			return result, err
		}
	})

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		callOrder = append(callOrder, "handler")
		return &mcp.GetPromptResult{}, nil
	}

	wrapped := chain.WrapPromptHandler(handler)
	_, _ = wrapped(context.Background(), &mcp.GetPromptRequest{})

	if len(callOrder) != 3 || callOrder[0] != "mw-before" || callOrder[1] != "handler" || callOrder[2] != "mw-after" {
		t.Errorf("Call order = %v, want [mw-before, handler, mw-after]", callOrder)
	}
}

func TestMiddlewareChain_AddResourceMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()

	callOrder := []string{}

	chain.AddResourceMiddleware(func(next ResourceHandlerFunc) ResourceHandlerFunc {
		return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			callOrder = append(callOrder, "mw-before")
			result, err := next(ctx, req)
			callOrder = append(callOrder, "mw-after")
			return result, err
		}
	})

	handler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		callOrder = append(callOrder, "handler")
		return &mcp.ReadResourceResult{}, nil
	}

	wrapped := chain.WrapResourceHandler(handler)
	_, _ = wrapped(context.Background(), &mcp.ReadResourceRequest{})

	if len(callOrder) != 3 || callOrder[0] != "mw-before" || callOrder[1] != "handler" || callOrder[2] != "mw-after" {
		t.Errorf("Call order = %v, want [mw-before, handler, mw-after]", callOrder)
	}
}

func TestMiddlewareChain_ApplyMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()

	callOrder := []string{}

	mw := &testMiddleware{
		toolFunc: func(next ToolHandlerFunc) ToolHandlerFunc {
			return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				callOrder = append(callOrder, "tool-mw")
				return next(ctx, req)
			}
		},
		promptFunc: func(next PromptHandlerFunc) PromptHandlerFunc {
			return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				callOrder = append(callOrder, "prompt-mw")
				return next(ctx, req)
			}
		},
		resourceFunc: func(next ResourceHandlerFunc) ResourceHandlerFunc {
			return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				callOrder = append(callOrder, "resource-mw")
				return next(ctx, req)
			}
		},
	}

	chain.ApplyMiddleware(mw)

	// Test tool middleware
	toolHandler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		callOrder = append(callOrder, "tool-handler")
		return &mcp.CallToolResult{}, nil
	}
	wrapped := chain.WrapToolHandler(toolHandler)
	_, _ = wrapped(context.Background(), &mcp.CallToolRequest{})

	if len(callOrder) != 2 || callOrder[0] != "tool-mw" || callOrder[1] != "tool-handler" {
		t.Errorf("Tool middleware call order = %v, want [tool-mw, tool-handler]", callOrder)
	}
}

// testMiddleware is a test implementation of Middleware interface
type testMiddleware struct {
	toolFunc     func(ToolHandlerFunc) ToolHandlerFunc
	promptFunc   func(PromptHandlerFunc) PromptHandlerFunc
	resourceFunc func(ResourceHandlerFunc) ResourceHandlerFunc
}

func (tm *testMiddleware) ToolMiddleware(next ToolHandlerFunc) ToolHandlerFunc {
	if tm.toolFunc != nil {
		return tm.toolFunc(next)
	}
	return next
}

func (tm *testMiddleware) PromptMiddleware(next PromptHandlerFunc) PromptHandlerFunc {
	if tm.promptFunc != nil {
		return tm.promptFunc(next)
	}
	return next
}

func (tm *testMiddleware) ResourceMiddleware(next ResourceHandlerFunc) ResourceHandlerFunc {
	if tm.resourceFunc != nil {
		return tm.resourceFunc(next)
	}
	return next
}

func TestWithMiddleware_Interface(t *testing.T) {
	adapter := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(&testMiddleware{}))

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter.middleware == nil {
		t.Error("Expected middleware chain to be created")
	}
}

func TestWithMiddleware_ConfigFunc(t *testing.T) {
	adapter := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(func(chain *MiddlewareChain) {
		chain.AddToolMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
			return next
		})
	}))

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
}

func TestWithMiddleware_ToolMiddlewareFunc(t *testing.T) {
	adapter := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
		return next
	}))

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
}

func TestWithMiddleware_Nil(t *testing.T) {
	// Should not panic with nil middleware
	adapter := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(nil))

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
}
