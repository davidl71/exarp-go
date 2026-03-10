package gosdk

import (
	"context"
	"encoding/json"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HooksToolMiddleware returns a tool middleware that runs Hooks.BeforeToolCall and
// Hooks.AfterToolCall around each tool invocation. If hooks is nil, the middleware
// is a no-op.
func HooksToolMiddleware(hooks *framework.Hooks) func(ToolHandlerFunc) ToolHandlerFunc {
	if hooks == nil {
		return func(next ToolHandlerFunc) ToolHandlerFunc { return next }
	}
	return func(next ToolHandlerFunc) ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := ""
			var rawArgs json.RawMessage
			if req != nil && req.Params != nil {
				name = req.Params.Name
				rawArgs = req.Params.Arguments
			}
			if hooks.BeforeToolCall != nil {
				hooks.BeforeToolCall(ctx, name, rawArgs)
			}
			result, err := next(ctx, req)
			if hooks.AfterToolCall != nil {
				hooks.AfterToolCall(ctx, name, rawArgs)
			}
			return result, err
		}
	}
}
