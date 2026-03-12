// request.go — Shared request parsing helpers used by framework adapters and tools.
package framework

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ApplyDefaults applies default values to a params map.
// Defaults are applied if the key is missing, has an empty string, or has a zero value (0, 0.0, false, empty array).
func ApplyDefaults(params map[string]interface{}, defaults map[string]interface{}) {
	if params == nil {
		return
	}
	for key, defaultValue := range defaults {
		existingValue, exists := params[key]
		if !exists {
			params[key] = defaultValue
			continue
		}
		// Handle empty string
		if strValue, ok := existingValue.(string); ok && strValue == "" {
			params[key] = defaultValue
			continue
		}
		// Handle zero values: 0, 0.0, false, empty array
		if isZeroValue(existingValue) {
			params[key] = defaultValue
		}
	}
}

// isZeroValue returns true if the value is a Go zero value (0, 0.0, false, nil, empty slice).
func isZeroValue(v interface{}) bool {
	switch val := v.(type) {
	case int:
		return val == 0
	case int8:
		return val == 0
	case int16:
		return val == 0
	case int32:
		return val == 0
	case int64:
		return val == 0
	case float32:
		return val == 0
	case float64:
		return val == 0
	case bool:
		return !val
	case []interface{}:
		return len(val) == 0
	case nil:
		return true
	}
	return false
}

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

	req := newMsg()
	if err := proto.Unmarshal(args, req); err == nil {
		return req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return zero, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return zero, params, nil
}

// ProtobufToParamsOptions configures the behavior of ProtobufToParams.
type ProtobufToParamsOptions struct {
	FilterEmptyStrings  bool
	StringifyArrays     bool
	ConvertFloat64ToInt bool
	Float64ToIntFields  []string
}

// ProtobufToParams converts a protobuf message to a map[string]interface{}.
func ProtobufToParams(msg proto.Message, opts *ProtobufToParamsOptions) (map[string]interface{}, error) {
	if msg == nil {
		return make(map[string]interface{}), nil
	}

	jsonBytes, err := protojson.MarshalOptions{
		EmitDefaultValues: true,
		UseProtoNames:     true,
	}.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf to JSON: %w", err)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	if opts != nil {
		if opts.FilterEmptyStrings {
			params = filterEmptyStrings(params)
		}
		if opts.StringifyArrays {
			params = stringifyArrays(params)
		}
		if opts.ConvertFloat64ToInt {
			params = convertFloat64ToInt(params, opts.Float64ToIntFields)
		}
	}

	return params, nil
}

func filterEmptyStrings(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		if f, ok := v.(float64); ok && f == 0.0 {
			continue
		}
		result[k] = v
	}
	return result
}

func stringifyArrays(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		switch val := v.(type) {
		case []interface{}:
			if len(val) == 0 {
				continue
			}
			jsonBytes, err := json.Marshal(val)
			if err == nil {
				result[k] = string(jsonBytes)
			} else {
				result[k] = val
			}
		default:
			result[k] = v
		}
	}
	return result
}

func convertFloat64ToInt(params map[string]interface{}, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		return params
	}

	fieldSet := make(map[string]bool, len(fields))
	for _, field := range fields {
		fieldSet[field] = true
	}

	result := make(map[string]interface{})
	for k, v := range params {
		if fieldSet[k] {
			if f, ok := v.(float64); ok {
				result[k] = int(f)
				continue
			}
		}
		result[k] = v
	}
	return result
}
