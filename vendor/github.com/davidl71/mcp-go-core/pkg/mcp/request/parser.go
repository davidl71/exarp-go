// Package request provides generic utilities for parsing MCP tool requests.
//
// This package includes generic functions for parsing protobuf or JSON requests,
// eliminating the need for repetitive parsing code in tool handlers.
//
// Example:
//
//	req, params, err := request.ParseRequest(args, func() *proto.MyRequest {
//		return &proto.MyRequest{}
//	})
//	if err != nil {
//		return nil, err
//	}
//	if req != nil {
//		// Use protobuf request
//	} else {
//		// Use JSON params map
//	}
package request

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// ParseRequest is a generic function for parsing protobuf or JSON requests.
//
// It attempts to parse the input as a protobuf message first. If that fails,
// it falls back to parsing as JSON into a map[string]interface{}.
//
// T must be a protobuf message type that implements proto.Message.
// newMessage is a function that returns a new zero-value instance of T.
//
// Returns:
//   - If protobuf parsing succeeds: the parsed protobuf message, nil params map, nil error
//   - If JSON parsing succeeds: zero-value of T, params map, nil error
//   - If both fail: zero-value of T, nil params map, error describing the failure
//
// Example:
//
//	type MyRequest struct {
//		proto.Message
//		Action string `protobuf:"bytes,1,opt,name=action"`
//	}
//
//	req, params, err := ParseRequest(args, func() *MyRequest {
//		return &MyRequest{}
//	})
func ParseRequest[T proto.Message](
	args json.RawMessage,
	newMessage func() T,
) (T, map[string]interface{}, error) {
	var zero T

	// Try protobuf binary first
	req := newMessage()
	if err := proto.Unmarshal(args, req); err == nil {
		// Successfully parsed as protobuf
		return req, nil, nil
	}

	// Fall back to JSON
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return zero, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Successfully parsed as JSON
	return zero, params, nil
}
