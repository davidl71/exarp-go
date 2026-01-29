package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// MLXNativeAvailable reports whether native MLX is available. Always false after luxfi/mlx removal; use Python bridge.
func MLXNativeAvailable() bool {
	return false
}

// handleMlxNative handles the mlx tool without native luxfi/mlx. Implements "models" (static list); other actions use the bridge.
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
