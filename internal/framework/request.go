// request.go — Framework wrappers for mcp-go-core request utilities.
package framework

import (
	"bytes"
	"encoding/json"
	"fmt"

	mcprequest "github.com/davidl71/mcp-go-core/pkg/mcp/request"
	"google.golang.org/protobuf/proto"
)

// ApplyDefaults applies default values to a params map.
// Defaults are only applied if the key is missing or has an empty string value.
var ApplyDefaults = mcprequest.ApplyDefaults

// ParseRequest parses a protobuf or JSON request from raw MCP args.
// Returns the protobuf message if protobuf binary succeeds; otherwise returns a JSON params map.
func ParseRequest[T proto.Message](args json.RawMessage, newMsg func() T) (T, map[string]interface{}, error) {
	var zero T

	trimmed := bytes.TrimSpace(args)
	if len(trimmed) == 0 {
		return zero, map[string]interface{}{}, nil
	}

	// MCP tool arguments arrive as JSON objects. Detect valid JSON first so raw JSON bytes
	// are not misinterpreted as a successfully decoded zero-value protobuf message.
	if json.Valid(trimmed) {
		switch trimmed[0] {
		case '{':
			var params map[string]interface{}
			if err := json.Unmarshal(trimmed, &params); err != nil {
				return zero, nil, fmt.Errorf("failed to parse arguments: %w", err)
			}
			return zero, params, nil
		case '[':
			return zero, nil, fmt.Errorf("failed to parse arguments: expected JSON object")
		}
	}

	return mcprequest.ParseRequest(args, newMsg)
}

// ProtobufToParamsOptions configures the behavior of ProtobufToParams.
type ProtobufToParamsOptions = mcprequest.ProtobufToParamsOptions

// ProtobufToParams converts a protobuf message to a map[string]interface{}.
var ProtobufToParams = mcprequest.ProtobufToParams
