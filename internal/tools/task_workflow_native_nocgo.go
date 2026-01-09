//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskWorkflowNative handles task_workflow with native Go (no Apple FM)
func handleTaskWorkflowNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "sync"
	}

	switch action {
	case "approve":
		// Approve doesn't need Apple FM, so it works on all platforms
		return handleTaskWorkflowApprove(ctx, params)
	case "sync":
		return handleTaskWorkflowSync(ctx, params)
	case "clarity":
		return handleTaskWorkflowClarity(ctx, params)
	case "cleanup":
		return handleTaskWorkflowCleanup(ctx, params)
	case "clarify":
		// Clarify requires Apple FM which is not available on this platform
		return nil, fmt.Errorf("clarify action requires Apple Foundation Models (not available on this platform)")
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
