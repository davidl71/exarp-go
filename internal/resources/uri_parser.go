// uri_parser.go — URI parsing helpers for stdio:// resource path segments.
package resources

import (
	"fmt"
	"strings"
)

// parseURIVariableByIndex extracts a variable from a URI by path index.
// This helper eliminates repetitive URI parsing code in resource handlers.
//
// URI format: stdio://resource/path/to/{variable}
// Split by "/" gives: ["stdio:", "", "resource", "path", "to", "{variable}"]
//
// Examples:
//   - parseURIVariableByIndex("stdio://tasks/T-123", 3) → "T-123", nil
//   - parseURIVariableByIndex("stdio://tasks/status/Todo", 3) → "Todo", nil
//   - parseURIVariableByIndex("stdio://memories/category/debug", 3) → "debug", nil
//
// Parameters:
//   - uri: The actual URI to parse (e.g., "stdio://tasks/T-123")
//   - index: The path segment index (0-based after "stdio://")
//   - expectedFormat: Optional format description for error messages
//
// Returns:
//   - The extracted variable value
//   - An error if the URI format is invalid or variable is empty
func parseURIVariableByIndex(uri string, index int, expectedFormat string) (string, error) {
	parts := strings.Split(uri, "/")

	// Validate URI has enough parts
	// Minimum: ["stdio:", "", "resource"] = 3 parts (index 2)
	// For index 3, need at least 4 parts
	if len(parts) <= index {
		if expectedFormat != "" {
			return "", fmt.Errorf("invalid URI format: %s (expected %s)", uri, expectedFormat)
		}

		return "", fmt.Errorf("invalid URI format: %s (index %d out of range, got %d parts)", uri, index, len(parts))
	}

	value := parts[index]
	if value == "" {
		return "", fmt.Errorf("variable at index %d is empty in URI: %s", index, uri)
	}

	return value, nil
}

// parseURIVariableByIndexWithValidation extracts a variable and validates it's not empty.
// This is a convenience wrapper around parseURIVariableByIndex that adds explicit empty check.
//
// Returns an error if the variable is empty (useful for required parameters).
func parseURIVariableByIndexWithValidation(uri string, index int, variableName, expectedFormat string) (string, error) {
	value, err := parseURIVariableByIndex(uri, index, expectedFormat)
	if err != nil {
		return "", err
	}

	if value == "" {
		return "", fmt.Errorf("%s cannot be empty (URI: %s)", variableName, uri)
	}

	return value, nil
}
