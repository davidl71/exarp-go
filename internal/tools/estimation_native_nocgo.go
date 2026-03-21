//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
)

// handleEstimationNative handles estimation using native Go (non-Apple platforms).
// Now uses the shared implementation with runtime feature detection.
func handleEstimationNative(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	return HandleEstimationNative(ctx, projectRoot, params)
}
