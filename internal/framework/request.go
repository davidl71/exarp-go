// request.go — Framework wrappers for mcp-go-core request utilities.
package framework

import (
	"encoding/json"

	mcprequest "github.com/davidl71/mcp-go-core/pkg/mcp/request"
	"google.golang.org/protobuf/proto"
)

// ApplyDefaults applies default values to a params map.
// Defaults are only applied if the key is missing or has an empty string value.
var ApplyDefaults = mcprequest.ApplyDefaults

// ParseRequest parses a protobuf or JSON request from raw MCP args.
// Returns the protobuf message if protobuf binary succeeds; otherwise returns a JSON params map.
func ParseRequest[T proto.Message](args json.RawMessage, newMsg func() T) (T, map[string]interface{}, error) {
	return mcprequest.ParseRequest(args, newMsg)
}

// ProtobufToParamsOptions configures the behavior of ProtobufToParams.
type ProtobufToParamsOptions = mcprequest.ProtobufToParamsOptions

// ProtobufToParams converts a protobuf message to a map[string]interface{}.
var ProtobufToParams = mcprequest.ProtobufToParams
