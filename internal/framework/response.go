// response.go — Framework wrappers for mcp-go-core response utilities.
package framework

import mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"

// FormatResult formats a result map as indented JSON and optionally writes it to a file.
func FormatResult(result map[string]interface{}, outputPath string) ([]TextContent, error) {
	return mcpresponse.FormatResult(result, outputPath)
}

// FormatResultCompact formats a result map as compact (non-indented) JSON and optionally writes it to a file.
func FormatResultCompact(result map[string]interface{}, outputPath string) ([]TextContent, error) {
	return mcpresponse.FormatResultCompact(result, outputPath)
}

// ConvertToMap converts any result to map[string]interface{} via JSON marshal/unmarshal.
func ConvertToMap(result interface{}) (map[string]interface{}, error) {
	return mcpresponse.ConvertToMap(result)
}
