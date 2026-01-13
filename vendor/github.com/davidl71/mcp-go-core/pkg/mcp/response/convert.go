// Package response provides utilities for formatting MCP tool responses.
package response

import (
	"encoding/json"
	"fmt"
)

// ConvertToMap converts any result to map[string]interface{}.
// Handles both maps and structs by marshaling/unmarshaling through JSON.
// This is useful for standardizing tool responses before formatting.
//
// The function:
//   - Returns the input unchanged if it's already a map[string]interface{}
//   - Converts structs and other types to map by marshaling to JSON and unmarshaling
//   - Returns an error if JSON marshaling or unmarshaling fails
//
// Parameters:
//   - result: The result to convert (can be map, struct, or any JSON-serializable type)
//
// Returns:
//   - map[string]interface{}: The converted result as a map
//   - error: Error if conversion fails
//
// Example:
//
//	// Convert a struct to map
//	type Result struct {
//		Success bool   `json:"success"`
//		Message string `json:"message"`
//	}
//	result := Result{Success: true, Message: "done"}
//	resultMap, err := ConvertToMap(result)
//	if err != nil {
//		return nil, err
//	}
//	// resultMap is now map[string]interface{}{"success": true, "message": "done"}
//
//	// Already a map - returned unchanged
//	resultMap := map[string]interface{}{"key": "value"}
//	converted, err := ConvertToMap(resultMap)
//	// converted == resultMap (same reference)
func ConvertToMap(result interface{}) (map[string]interface{}, error) {
	// If already a map, return it
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}

	// Marshal to JSON and unmarshal to map
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &resultMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result to map: %w", err)
	}

	return resultMap, nil
}
