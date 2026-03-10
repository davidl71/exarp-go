// response.go — Shared response formatting helpers used by MCP tool handlers.
package framework

import (
	"encoding/json"
	"fmt"
	"os"
)

// FormatResult formats a result map as indented JSON and optionally writes it to a file.
func FormatResult(result map[string]interface{}, outputPath string) ([]TextContent, error) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	if outputPath != "" {
		if err := os.WriteFile(outputPath, output, 0o644); err == nil {
			result["output_path"] = outputPath
			output, err = json.MarshalIndent(result, "", "  ")
			if err != nil {
				return []TextContent{{Type: "text", Text: string(output)}}, nil
			}
		}
	}
	return []TextContent{{Type: "text", Text: string(output)}}, nil
}

// FormatResultCompact formats a result map as compact (non-indented) JSON and optionally writes it to a file.
func FormatResultCompact(result map[string]interface{}, outputPath string) ([]TextContent, error) {
	output, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	if outputPath != "" {
		if err := os.WriteFile(outputPath, output, 0o644); err == nil {
			result["output_path"] = outputPath
			output, err = json.Marshal(result)
			if err != nil {
				return []TextContent{{Type: "text", Text: string(output)}}, nil
			}
		}
	}
	return []TextContent{{Type: "text", Text: string(output)}}, nil
}

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
