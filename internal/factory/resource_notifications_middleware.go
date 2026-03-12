package factory

import (
	"context"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func resourceUpdateMiddleware(notifier framework.ResourceUpdateNotifier) func(gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
	if notifier == nil {
		return func(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
			return next
		}
	}

	return func(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
		return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx = tools.WithResourceUpdateContext(ctx)
			result, err := next(ctx, req)
			sendResourceNotifications(ctx, notifier)
			return result, err
		}
	}
}

func sendResourceNotifications(ctx context.Context, notifier framework.ResourceUpdateNotifier) {
	if notifier == nil {
		return
	}

	for _, uri := range tools.ResourceUpdates(ctx) {
		_ = notifier.NotifyResourceUpdated(ctx, &mcp.ResourceUpdatedNotificationParams{URI: uri})
	}
}
