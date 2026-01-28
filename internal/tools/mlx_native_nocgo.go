//go:build !darwin || !cgo
// +build !darwin !cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// MLXNativeAvailable reports whether native MLX (luxfi/mlx) is available.
// On non-darwin or without CGO, native MLX is not available.
func MLXNativeAvailable() bool {
	return false
}

// handleMlxNative handles the mlx tool when luxfi/mlx is not available.
// Implements "models" (static list); all other actions return an error so the handler can use the bridge.
func handleMlxNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "status"
	}
	switch action {
	case "models":
		return response.FormatResult(mlxModelsResponse(), "")
	default:
		return nil, fmt.Errorf("MLX native not available (requires darwin with CGO); use Python bridge for status, hardware, generate")
	}
}
