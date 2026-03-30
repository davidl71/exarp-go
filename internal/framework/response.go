// response.go — Shared response formatting helpers used by MCP tool handlers.
// FormatResult* delegate to mcp-go-core; ConvertToMap stays local (JSON round-trip semantics).
package framework

import (
	"encoding/json"
	"fmt"

	responsecore "github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// FormatResult formats a result map as indented JSON and optionally writes it to a file.
var FormatResult = responsecore.FormatResult

// FormatResultCompact formats a result map as compact (non-indented) JSON and optionally writes it to a file.
var FormatResultCompact = responsecore.FormatResultCompact

// ConvertToMap converts any result to map[string]interface{} via JSON marshal/unmarshal.
func ConvertToMap(result interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}
	return out, nil
}
