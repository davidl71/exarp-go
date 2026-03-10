// Package response provides utilities for formatting MCP tool responses.
//
// This package includes generic functions for formatting result maps as JSON
// and optionally writing them to files, eliminating repetitive formatting code
// in tool handlers.
//
// Example:
//
//	result := map[string]interface{}{
//		"success": true,
//		"data":    "example",
//	}
//	contents, err := response.FormatResult(result, "/path/to/output.json")
//	if err != nil {
//		return nil, err
//	}
//	return contents, nil
package response

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// FormatResult formats a result map as JSON and optionally writes it to a file.
//
// The function:
//   - Marshals the result map to indented JSON
//   - Optionally writes to a file if outputPath is provided
//   - Includes output_path in the result if file writing succeeds
//   - Returns the formatted JSON as TextContent for MCP protocol
//
// Parameters:
//   - result: The result map to format (will be modified if outputPath is provided)
//   - outputPath: Optional file path to write the JSON to (empty string to skip)
//
// Returns:
//   - []types.TextContent: Formatted JSON response for MCP protocol
//   - error: Error if JSON marshaling or file writing fails
//
// Example:
//
//	result := map[string]interface{}{
//		"success": true,
//		"method":  "native_go",
//	}
//	contents, err := FormatResult(result, "/tmp/output.json")
//	if err != nil {
//		return nil, err
//	}
//	// contents[0].Text contains the JSON string
//	// result["output_path"] is set if file was written successfully
func FormatResult(result map[string]interface{}, outputPath string) ([]types.TextContent, error) {
	// Marshal result to indented JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file if outputPath is provided
	if outputPath != "" {
		if err := os.WriteFile(outputPath, output, 0644); err == nil {
			// File written successfully - add output_path to result
			result["output_path"] = outputPath
			// Re-marshal with output_path included
			output, err = json.MarshalIndent(result, "", "  ")
			if err != nil {
				// If re-marshaling fails, return original output
				// (output_path was added but couldn't be included in JSON)
				return []types.TextContent{
					{Type: "text", Text: string(output)},
				}, nil
			}
		}
		// If file write fails, continue without output_path
		// (don't fail the entire operation)
	}

	return []types.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}
