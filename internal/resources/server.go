package resources

import (
	"context"
	"encoding/json"
	"os"
)

// handleServerStatus handles the stdio://server/status resource
// Returns server operational status, version, and project root information
func handleServerStatus(ctx context.Context, uri string) ([]byte, string, error) {
	// Get project root from environment variable
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "unknown"
	}

	// Build status response (matching tool format)
	status := map[string]interface{}{
		"status":          "operational",
		"version":         "0.1.0",
		"tools_available": "See tool catalog",
		"project_root":    projectRoot,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, "", err
	}

	return jsonData, "application/json", nil
}
