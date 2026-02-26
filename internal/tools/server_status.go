package tools

import (
	"context"
	"encoding/json"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleServerStatusNative handles the server_status tool with native Go implementation
// Returns server operational status, version, and project root information.
func handleServerStatusNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		projectRoot = "unknown"
	}

	// Build status response
	status := map[string]interface{}{
		"status":          "operational",
		"version":         "0.1.0",
		"tools_available": "See tool catalog",
		"project_root":    projectRoot,
	}

	return framework.FormatResult(status, "")
}
