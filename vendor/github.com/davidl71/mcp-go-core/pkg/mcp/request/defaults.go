// Package request provides generic utilities for parsing MCP tool requests.
//
// This package includes functions for applying default values to parameter maps,
// eliminating repetitive default-setting code in tool handlers.
//
// Example:
//
//	params := map[string]interface{}{
//		"action": "custom",
//	}
//	request.ApplyDefaults(params, map[string]interface{}{
//		"action":        "sync",
//		"sub_action":    "list",
//		"output_format": "text",
//	})
//	// params["action"] remains "custom" (not overridden)
//	// params["sub_action"] is set to "list" (was missing)
//	// params["output_format"] is set to "text" (was missing)
package request

// ApplyDefaults applies default values to a params map.
//
// Defaults are only applied if:
//   - The key doesn't exist in params, OR
//   - The existing value is an empty string
//
// This ensures that:
//   - Existing non-empty values are preserved
//   - Missing parameters get default values
//   - Empty strings are replaced with defaults
//
// Parameters:
//   - params: The parameter map to apply defaults to (modified in place)
//   - defaults: Map of default values to apply
//
// Example:
//
//	params := map[string]interface{}{
//		"action": "custom",        // Will NOT be overridden
//		"status": "",              // Will be set to "Review"
//		// "limit" is missing      // Will be set to 10
//	}
//	ApplyDefaults(params, map[string]interface{}{
//		"action": "sync",
//		"status": "Review",
//		"limit":  10,
//	})
//	// Result:
//	// params["action"] = "custom" (preserved)
//	// params["status"] = "Review" (replaced empty string)
//	// params["limit"] = 10 (added)
func ApplyDefaults(params map[string]interface{}, defaults map[string]interface{}) {
	// Ensure params map exists
	if params == nil {
		return
	}

	// Apply each default value
	for key, defaultValue := range defaults {
		// Check if key exists and has a non-empty value
		existingValue, exists := params[key]

		// Apply default if:
		// 1. Key doesn't exist, OR
		// 2. Existing value is empty string
		if !exists {
			// Key doesn't exist - apply default
			params[key] = defaultValue
		} else if strValue, ok := existingValue.(string); ok && strValue == "" {
			// Existing value is empty string - replace with default
			params[key] = defaultValue
		}
		// Otherwise, preserve existing non-empty value
	}
}
