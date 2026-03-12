// Package request provides generic utilities for parsing MCP tool requests.
//
// This package includes generic functions for parsing protobuf or JSON requests,
// eliminating the need for repetitive parsing code in tool handlers.
package request

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtobufToParams converts a protobuf message to a map[string]interface{}.
//
// This function uses protojson to convert the protobuf message to JSON,
// then unmarshals it into a map. This eliminates the need for manual
// field-by-field conversion in *RequestToParams functions.
//
// The function applies post-processing to match the behavior of existing
// *RequestToParams functions:
//   - Filters out empty string values (optional, configurable)
//   - Converts arrays to JSON strings (for backward compatibility)
//   - Handles field name mapping (camelCase to snake_case)
//
// Example:
//
//	req := &proto.TaskWorkflowRequest{
//		Action: "sync",
//		TaskId: "T-123",
//	}
//	params, err := ProtobufToParams(req, &ProtobufToParamsOptions{
//		FilterEmptyStrings: true,
//		StringifyArrays:   true,
//	})
func ProtobufToParams(msg proto.Message, opts *ProtobufToParamsOptions) (map[string]interface{}, error) {
	if msg == nil {
		return make(map[string]interface{}), nil
	}

	// Use protojson to convert protobuf to JSON
	// Use EmitDefaultValues=true to include all fields (including false booleans)
	// Use UseProtoNames=true to use snake_case field names (matching existing behavior)
	// This ensures field names like "task_id" instead of "taskId"
	// We use EmitDefaultValues=true because some *RequestToParams functions always include booleans
	marshalOpts := protojson.MarshalOptions{
		EmitDefaultValues: true,  // Include all fields (including false booleans and zero values)
		UseProtoNames:     true,  // Use snake_case field names from proto (e.g., "task_id" not "taskId")
	}

	jsonBytes, err := marshalOpts.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf to JSON: %w", err)
	}

	// Unmarshal JSON into map
	var params map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	// Apply post-processing options
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

// ProtobufToParamsOptions configures the behavior of ProtobufToParams.
type ProtobufToParamsOptions struct {
	// FilterEmptyStrings removes empty string values from the params map.
	// This matches the behavior of existing *RequestToParams functions that
	// only include fields if they are non-empty.
	FilterEmptyStrings bool

	// StringifyArrays converts array/slice values to JSON strings.
	// This matches the behavior of existing *RequestToParams functions that
	// marshal arrays to JSON strings for compatibility.
	StringifyArrays bool

	// ConvertFloat64ToInt converts specific float64 fields to int.
	// This is needed because protojson converts int32/int64 to float64 in JSON,
	// but some *RequestToParams functions convert them back to int.
	ConvertFloat64ToInt bool

	// Float64ToIntFields is a list of field names that should be converted from float64 to int.
	// Only used if ConvertFloat64ToInt is true.
	Float64ToIntFields []string
}

// filterEmptyStrings removes empty string values and zero numeric values from the params map.
// This matches the behavior of existing *RequestToParams functions that
// only include string fields if they are non-empty, and numeric fields if they are > 0.
// Booleans are always kept (even if false), matching existing behavior.
func filterEmptyStrings(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		// Skip empty strings
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		// Skip zero numeric values (but keep booleans, even if false)
		if f, ok := v.(float64); ok && f == 0.0 {
			continue
		}
		// Keep all other values (booleans, non-zero numbers, arrays, etc.)
		result[k] = v
	}
	return result
}

// stringifyArrays converts array/slice values to JSON strings.
// This matches the behavior of existing *RequestToParams functions that
// marshal arrays to JSON strings for compatibility with the params map format.
func stringifyArrays(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		// Check if value is an array/slice
		switch val := v.(type) {
		case []interface{}:
			// Convert array to JSON string (only if non-empty, matching existing behavior)
			if len(val) > 0 {
				jsonBytes, err := json.Marshal(val)
				if err == nil {
					result[k] = string(jsonBytes)
				} else {
					// If marshaling fails, keep original value
					result[k] = val
				}
			} else {
				// Empty arrays are not included (matching existing behavior)
				// This matches: if len(req.Tags) > 0 { ... }
				continue
			}
		default:
			// Keep non-array values as-is
			result[k] = v
		}
	}
	return result
}

// convertFloat64ToInt converts specific float64 fields to int.
// This is needed because protojson converts numeric fields to float64 in JSON,
// but some *RequestToParams functions convert them back to int.
func convertFloat64ToInt(params map[string]interface{}, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		return params
	}

	// Create a set of fields to convert for fast lookup
	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[field] = true
	}

	result := make(map[string]interface{})
	for k, v := range params {
		// Convert if field is in the list and value is float64
		if fieldSet[k] {
			if f, ok := v.(float64); ok {
				// Convert to int (truncates, matching existing behavior)
				result[k] = int(f)
			} else {
				// Keep original value if not float64
				result[k] = v
			}
		} else {
			// Keep non-converted fields as-is
			result[k] = v
		}
	}
	return result
}
