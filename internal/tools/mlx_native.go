//go:build darwin && cgo
// +build darwin,cgo

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/luxfi/mlx"
)

// MLXNativeAvailable reports whether native MLX (luxfi/mlx) is available.
// On darwin with CGO, we consider it available if the package loads.
func MLXNativeAvailable() bool {
	return true
}

// handleMlxNative handles the mlx tool using native Go (luxfi/mlx).
// Implements status, hardware, and models; generate falls back to bridge (return error so caller can bridge).
func handleMlxNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "status"
	}

	switch action {
	case "status":
		info := mlx.Info()
		result := map[string]interface{}{
			"success":   true,
			"method":    "native_go",
			"info":      info,
			"backend":   mlx.GetBackend().String(),
			"available": true,
		}
		return response.FormatResult(result, "")
	case "hardware":
		backend := mlx.GetBackend()
		device := mlx.GetDevice()
		hw := map[string]interface{}{
			"backend": backend.String(),
			"platform": "apple_silicon",
			"mlx_supported": true,
			"metal_available": backend.String() == "Metal",
		}
		if device != nil {
			hw["device_name"] = device.Name
			hw["device_memory_bytes"] = device.Memory
			hw["device_type"] = device.Type.String()
		}
		result := map[string]interface{}{
			"success":  true,
			"method":   "native_go",
			"hardware": hw,
		}
		return response.FormatResult(result, "")
	case "models":
		return response.FormatResult(mlxModelsResponse(), "")
	case "generate":
		return nil, fmt.Errorf("mlx action %q not implemented in native Go; use Python bridge", action)
	default:
		return nil, fmt.Errorf("unknown mlx action %q; use status, hardware, models, or generate", action)
	}
}
