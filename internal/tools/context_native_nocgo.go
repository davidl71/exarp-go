//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleContextSummarizeNative uses the shared implementation from context_shared.go.
// The shared handler checks FMAvailable() and returns appropriate error.
func handleContextSummarizeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return HandleContextSummarizeShared(ctx, params)
}
