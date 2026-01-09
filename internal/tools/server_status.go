package tools

import (
	"context"
	"encoding/json"
	"os"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleServerStatusNative handles the server_status tool with native Go implementation
// Returns server operational status, version, and project root information
func handleServerStatusNative(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Get project root from environment variable
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "unknown"
	}

	// Build status response
	status := map[string]interface{}{
		"status":          "operational",
		"version":         "0.1.0",
		"tools_available": "See tool catalog",
		"project_root":    projectRoot,
	}

	// Marshal to JSON with indentation
	result, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, err
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}
