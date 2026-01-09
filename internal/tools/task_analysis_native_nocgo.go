//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskAnalysisNative handles task_analysis with native Go (no Apple FM)
// Platform-agnostic actions work on all platforms, hierarchy requires Apple FM
func handleTaskAnalysisNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "duplicates"
	}

	switch action {
	case "hierarchy":
		// Hierarchy requires Apple FM, not available on this platform
		return nil, fmt.Errorf("hierarchy action requires Apple Foundation Models (not available on this platform)")
	case "duplicates":
		return handleTaskAnalysisDuplicates(ctx, params)
	case "tags":
		return handleTaskAnalysisTags(ctx, params)
	case "dependencies":
		return handleTaskAnalysisDependencies(ctx, params)
	case "parallelization":
		return handleTaskAnalysisParallelization(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
