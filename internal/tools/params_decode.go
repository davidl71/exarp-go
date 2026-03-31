// params_decode.go — Typed MCP args decoding helpers (proto binary or JSON).
package tools

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// decodeArgsToProto decodes MCP args into a protobuf message.
//
// - If args are JSON object bytes, we protojson-unmarshal into the message.
// - Otherwise we attempt protobuf binary unmarshal.
//
// This is intended to keep tool entrypoints typed and reduce map[string]any + cast usage
// at the MCP boundary.
func decodeArgsToProto[T proto.Message](args json.RawMessage, newMsg func() T) (T, error) {
	req := newMsg()

	trimmed := bytes.TrimSpace(args)
	if len(trimmed) == 0 {
		return req, nil
	}

	if json.Valid(trimmed) {
		if trimmed[0] != '{' {
			return req, fmt.Errorf("failed to parse arguments: expected JSON object")
		}

		if err := (protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}).Unmarshal(trimmed, req); err != nil {
			return req, fmt.Errorf("failed to parse JSON arguments: %w", err)
		}

		return req, nil
	}

	if err := proto.Unmarshal(args, req); err != nil {
		return req, fmt.Errorf("failed to parse protobuf arguments: %w", err)
	}

	return req, nil
}

