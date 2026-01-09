//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskAnalysisNative is a no-op for non-Apple platforms
func handleTaskAnalysisNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return nil, fmt.Errorf("Apple Foundation Models not available on this platform")
}
