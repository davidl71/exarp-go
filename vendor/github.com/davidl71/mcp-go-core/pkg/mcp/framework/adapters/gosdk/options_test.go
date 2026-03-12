package gosdk

import (
	"context"
	"testing"

	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewGoSDKAdapter_WithOptions(t *testing.T) {
	// Test that adapter can be created with options (even if they're placeholders)
	adapter := NewGoSDKAdapter("test-server", "1.0.0", WithLogger(nil), WithMiddleware(nil))

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter.GetName() != "test-server" {
		t.Errorf("adapter.GetName() = %q, want %q", adapter.GetName(), "test-server")
	}
}

func TestNewGoSDKAdapter_WithoutOptions(t *testing.T) {
	// Test backward compatibility - adapter should work without options
	adapter := NewGoSDKAdapter("test-server", "1.0.0")

	if adapter == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter.GetName() != "test-server" {
		t.Errorf("adapter.GetName() = %q, want %q", adapter.GetName(), "test-server")
	}
}

func TestAdapterOption_WithLogger(t *testing.T) {
	// Test with nil logger (should use default)
	adapter1 := NewGoSDKAdapter("test", "1.0.0", WithLogger(nil))
	if adapter1 == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter1.logger == nil {
		t.Error("Expected default logger to be created")
	}

	// Test with custom logger
	customLogger := logging.NewLogger()
	customLogger.SetLevel(logging.LevelDebug)
	adapter2 := NewGoSDKAdapter("test", "1.0.0", WithLogger(customLogger))
	if adapter2 == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter2.logger != customLogger {
		t.Error("Expected custom logger to be set")
	}
	if adapter2.logger != customLogger {
		t.Error("Logger should be the same instance")
	}
}

func TestAdapterOption_WithMiddleware(t *testing.T) {
	// Test with nil middleware (should not crash)
	adapter1 := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(nil))
	if adapter1 == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
	if adapter1.middleware == nil {
		t.Error("Expected middleware chain to be created")
	}

	// Test with Middleware interface
	mw := &testMiddlewareForOptions{
		toolFunc: func(next ToolHandlerFunc) ToolHandlerFunc {
			return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return next(ctx, req)
			}
		},
	}
	adapter2 := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(mw))
	if adapter2 == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}

	// Test with config function
	adapter3 := NewGoSDKAdapter("test", "1.0.0", WithMiddleware(func(chain *MiddlewareChain) {
		chain.AddToolMiddleware(func(next ToolHandlerFunc) ToolHandlerFunc {
			return next
		})
	}))
	if adapter3 == nil {
		t.Fatal("NewGoSDKAdapter() returned nil")
	}
}

// testMiddlewareForOptions is a test implementation for options_test.go
type testMiddlewareForOptions struct {
	toolFunc func(ToolHandlerFunc) ToolHandlerFunc
}

func (tm *testMiddlewareForOptions) ToolMiddleware(next ToolHandlerFunc) ToolHandlerFunc {
	if tm.toolFunc != nil {
		return tm.toolFunc(next)
	}
	return next
}

func (tm *testMiddlewareForOptions) PromptMiddleware(next PromptHandlerFunc) PromptHandlerFunc {
	return next
}

func (tm *testMiddlewareForOptions) ResourceMiddleware(next ResourceHandlerFunc) ResourceHandlerFunc {
	return next
}
