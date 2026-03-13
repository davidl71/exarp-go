//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleContextSummarizeNative handles context summarization using native Go with Apple FM.
// Now uses the shared implementation from context_shared.go.
func handleContextSummarizeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return HandleContextSummarizeShared(ctx, params)
}
