// params_decode.go — Typed decoding helpers for MCP tool params (map-shaped JSON args).
package tools

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// decodeArgs decodes MCP JSON tool arguments into a proto request using protojson (camelCase /
// JSON names for enum fields). Empty input yields a zero message from newMsg().
func decodeArgsToProto[T proto.Message](args json.RawMessage, newMsg func() T) (T, error) {
	var zero T

	trim := bytes.TrimSpace(args)
	if len(trim) == 0 {
		return newMsg(), nil
	}

	if !json.Valid(trim) || trim[0] != '{' {
		return zero, fmt.Errorf("tool args: expected JSON object")
	}

	req := newMsg()

	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	if err := opts.Unmarshal(trim, req); err != nil {
		return zero, fmt.Errorf("tool args: %w", err)
	}

	return req, nil
}

// MapToStructViaJSON round-trips params through encoding/json so map-shaped MCP args decode into
// a tagged struct. Unknown fields are ignored; missing fields keep zero values.
func MapToStructViaJSON(params map[string]interface{}, dst interface{}) error {
	if dst == nil {
		return fmt.Errorf("dst is nil")
	}

	if params == nil {
		return fmt.Errorf("params is nil")
	}

	b, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("params encode: %w", err)
	}

	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("params decode: %w", err)
	}

	return nil
}
