package gosdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// MapToolHandler is a handler that accepts a params map and returns an arbitrary result.
// WrapMapToolHandler converts it to framework.ToolHandler by unmarshaling args,
// calling the handler, then converting the result to map and formatting with response.FormatResult.
type MapToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// WrapMapToolHandler wraps a MapToolHandler into a framework.ToolHandler.
// It unmarshals raw args to map[string]interface{}, calls the handler, then uses
// response.ConvertToMap and response.FormatResult to produce []types.TextContent.
// If params contain "output_path" (string), it is passed to FormatResult for optional file write.
func WrapMapToolHandler(handler MapToolHandler) framework.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
		params := make(map[string]interface{})
		if len(args) > 0 {
			if err := json.Unmarshal(args, &params); err != nil {
				return nil, fmt.Errorf("failed to parse arguments: %w", err)
			}
		}

		result, err := handler(ctx, params)
		if err != nil {
			return nil, err
		}

		resultMap, err := response.ConvertToMap(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result to map: %w", err)
		}

		outputPath := ""
		if path, ok := params["output_path"].(string); ok && path != "" {
			outputPath = path
		}

		contents, err := response.FormatResult(resultMap, outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to format result: %w", err)
		}

		return contents, nil
	}
}
