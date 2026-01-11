package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/tools"
)

// handleModels handles the stdio://models resource
// Returns all available AI models with capabilities
func handleModels(ctx context.Context, uri string) ([]byte, string, error) {
	// Build result (matching tool format)
	result := map[string]interface{}{
		"models": tools.MODEL_CATALOG,
		"count":  len(tools.MODEL_CATALOG),
		"tip":    "Use recommend_model for task-specific recommendations",
	}

	// Wrap in success response format (matching tool format)
	response := map[string]interface{}{
		"success":   true,
		"data":      result,
		"timestamp": 0, // No timestamp needed for resource
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal models: %w", err)
	}

	return jsonData, "application/json", nil
}
