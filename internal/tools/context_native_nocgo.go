//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleContextSummarizeNative is a no-op for non-Apple platforms
// The actual implementation is in context_native.go (requires darwin && arm64 && cgo)
func handleContextSummarizeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// This function should never be called on non-Apple platforms
	// because handleContext checks platform support before calling it
	return nil, fmt.Errorf("Apple Foundation Models not available on this platform")
}
