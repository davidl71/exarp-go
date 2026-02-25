// response_compact.go — optional compact JSON and token_estimate for MCP responses to reduce context size.
package tools

import (
	"encoding/json"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// AddTokenEstimateToResult sets result["token_estimate"] to the estimated token count of the
// serialized result (using config.TokensPerChar). Call before FormatResultOptionalCompact when
// output_format=json so AI clients can budget context. Mutates result in place.
func AddTokenEstimateToResult(result map[string]interface{}) {
	if result == nil {
		return
	}
	b, err := json.Marshal(result)
	if err != nil {
		return
	}
	tokens := int(float64(len(b)) * config.TokensPerChar())
	result["token_estimate"] = tokens
}

// FormatResultOptionalCompact formats a result map as JSON and optionally writes to a file.
// When compact is true, JSON is emitted without indentation (smaller payload, faster parse).
// When compact is false, delegates to mcp-go-core FormatResult (indented JSON).
func FormatResultOptionalCompact(result map[string]interface{}, outputPath string, compact bool) ([]framework.TextContent, error) {
	if !compact {
		return mcpresponse.FormatResult(result, outputPath)
	}
	return mcpresponse.FormatResultCompact(result, outputPath)
}
