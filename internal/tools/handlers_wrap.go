// handlers_wrap.go — WrapHandler: generic protobuf-aware adapter for MCP tool handlers.
// Reduces boilerplate for handlers that follow the parse → convert → defaults → native pattern.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/davidl71/exarp-go/internal/framework"
)

// ProtoParser parses raw JSON into an optional protobuf message and a params map.
// If protobuf parsing succeeds, the first return value is non-nil.
// If only JSON parsing works, the second return value holds the map.
type ProtoParser func(json.RawMessage) (any, map[string]interface{}, error)

// ProtoConverter converts a protobuf message (from ProtoParser) into a params map.
type ProtoConverter func(any) map[string]interface{}

// NativeHandler is the native Go handler that processes params and returns results.
type NativeHandler func(context.Context, map[string]interface{}) ([]framework.TextContent, error)

// WrapHandler creates a handler function from a parse/convert/defaults/native pipeline.
// This is the common pattern used by most MCP tool handlers:
//  1. Parse protobuf (with JSON fallback)
//  2. Convert proto to params map
//  3. Apply default values
//  4. Call native handler
func WrapHandler(
	toolName string,
	parse ProtoParser,
	convert ProtoConverter,
	defaults map[string]interface{},
	handler NativeHandler,
) func(context.Context, json.RawMessage) ([]framework.TextContent, error) {
	return func(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
		req, params, err := parse(args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if !isNilAny(req) {
			params = convert(req)
		}

		if params == nil {
			params = make(map[string]interface{})
		}

		// Apply defaults for both protobuf and JSON paths
		if defaults != nil {
			framework.ApplyDefaults(params, defaults)
		}

		result, err := handler(ctx, params)
		if err != nil {
			return nil, &framework.ErrToolFailed{ToolName: toolName, Err: err}
		}

		return result, nil
	}
}

func isNilAny(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
