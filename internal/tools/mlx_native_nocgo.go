package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// MLXNativeAvailable reports whether native MLX is available. Always false after luxfi/mlx removal; Python bridge removed.
func MLXNativeAvailable() bool {
	return false
}

// handleMlxNative handles the mlx tool without Python. Implements "models" (static list); status/hardware/generate return unavailable message.
func handleMlxNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "status"
	}
	switch action {
	case "models":
		return response.FormatResult(mlxModelsResponse(), "")
	case "status", "hardware":
		msg := map[string]interface{}{
			"success": true,
			"message": "MLX not available in this build. Use action=models for static model list, or use ollama/apple_foundation_models for generation.",
		}
		return response.FormatResult(msg, "")
	case "generate":
		return nil, fmt.Errorf("MLX generate not available in this build; use ollama or apple_foundation_models tool instead")
	default:
		return nil, fmt.Errorf("unknown mlx action %q; use status, hardware, models, or generate", action)
	}
}
