// helpers.go — Shared utility functions for tools.
//
// This file contains common helper functions used across multiple tools to avoid duplication.
package tools

// GetString retrieves a string from a map, returns empty string if not found.
func GetString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// GetStringDefault retrieves a string from a map, returns default if not found.
func GetStringDefault(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

// GetStringSlice retrieves a string slice from a map.
func GetStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, s := range val {
			if str, ok := s.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

// GetFloat retrieves a float64 from a map.
func GetFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}
