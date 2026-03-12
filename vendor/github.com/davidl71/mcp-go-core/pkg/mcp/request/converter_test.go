package request

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

// testMessage wraps structpb.Value to create a test protobuf message
type testMessage struct {
	*structpb.Value
}

func (t *testMessage) ProtoMessage() {}

func (t *testMessage) Reset() {
	t.Value = nil
}

func (t *testMessage) String() string {
	if t.Value == nil {
		return ""
	}
	return t.Value.String()
}

func TestProtobufToParams_Basic(t *testing.T) {
	// Create a simple protobuf message using structpb.Value
	msg, err := structpb.NewStruct(map[string]interface{}{
		"action": "test",
		"limit":  10,
		"enabled": true,
	})
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	// Convert to params
	params, err := ProtobufToParams(msg, &ProtobufToParamsOptions{
		FilterEmptyStrings: false,
		StringifyArrays:    false,
	})
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Verify basic fields are present
	if params["action"] != "test" {
		t.Errorf("params[\"action\"] = %v, want \"test\"", params["action"])
	}
}

func TestProtobufToParams_FilterEmptyStrings(t *testing.T) {
	// Create a message with empty strings
	msg, err := structpb.NewStruct(map[string]interface{}{
		"action":  "test",
		"empty":   "",
		"nonempty": "value",
	})
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	// Convert with FilterEmptyStrings=true
	params, err := ProtobufToParams(msg, &ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    false,
	})
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Verify empty string is filtered out
	if _, ok := params["empty"]; ok {
		t.Errorf("params[\"empty\"] should be filtered out, but was present")
	}

	// Verify non-empty values are present
	if params["action"] != "test" {
		t.Errorf("params[\"action\"] = %v, want \"test\"", params["action"])
	}
	if params["nonempty"] != "value" {
		t.Errorf("params[\"nonempty\"] = %v, want \"value\"", params["nonempty"])
	}
}

func TestProtobufToParams_StringifyArrays(t *testing.T) {
	// Create a message with arrays
	msg, err := structpb.NewStruct(map[string]interface{}{
		"tags": []interface{}{"tag1", "tag2"},
		"name": "test",
	})
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	// Convert with StringifyArrays=true
	params, err := ProtobufToParams(msg, &ProtobufToParamsOptions{
		FilterEmptyStrings: false,
		StringifyArrays:    true,
	})
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Verify array is stringified
	tagsVal, ok := params["tags"]
	if !ok {
		t.Fatalf("params[\"tags\"] should be present")
	}

	tagsStr, ok := tagsVal.(string)
	if !ok {
		t.Fatalf("params[\"tags\"] should be a string, got %T", tagsVal)
	}

	// Verify it's valid JSON
	var tagsArray []string
	if err := json.Unmarshal([]byte(tagsStr), &tagsArray); err != nil {
		t.Fatalf("params[\"tags\"] should be valid JSON: %v", err)
	}

	if len(tagsArray) != 2 || tagsArray[0] != "tag1" || tagsArray[1] != "tag2" {
		t.Errorf("params[\"tags\"] = %v, want [\"tag1\", \"tag2\"]", tagsArray)
	}

	// Verify non-array values are unchanged
	if params["name"] != "test" {
		t.Errorf("params[\"name\"] = %v, want \"test\"", params["name"])
	}
}

func TestProtobufToParams_ConvertFloat64ToInt(t *testing.T) {
	// Create a message with float64 values
	msg, err := structpb.NewStruct(map[string]interface{}{
		"stale_threshold_hours": 2.0,
		"other_float":          3.5,
		"name":                 "test",
	})
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	// Convert with ConvertFloat64ToInt=true
	params, err := ProtobufToParams(msg, &ProtobufToParamsOptions{
		FilterEmptyStrings:    false,
		StringifyArrays:       false,
		ConvertFloat64ToInt:    true,
		Float64ToIntFields:    []string{"stale_threshold_hours"},
	})
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Verify stale_threshold_hours is converted to int
	thresholdVal, ok := params["stale_threshold_hours"]
	if !ok {
		t.Fatalf("params[\"stale_threshold_hours\"] should be present")
	}

	thresholdInt, ok := thresholdVal.(int)
	if !ok {
		t.Fatalf("params[\"stale_threshold_hours\"] should be int, got %T", thresholdVal)
	}

	if thresholdInt != 2 {
		t.Errorf("params[\"stale_threshold_hours\"] = %v, want 2", thresholdInt)
	}

	// Verify other_float is not converted (not in the list)
	otherFloatVal, ok := params["other_float"]
	if !ok {
		t.Fatalf("params[\"other_float\"] should be present")
	}

	if _, ok := otherFloatVal.(float64); !ok {
		t.Errorf("params[\"other_float\"] should remain float64, got %T", otherFloatVal)
	}

	// Verify non-numeric values are unchanged
	if params["name"] != "test" {
		t.Errorf("params[\"name\"] = %v, want \"test\"", params["name"])
	}
}

func TestProtobufToParams_EmptyMessage(t *testing.T) {
	// Create an empty message
	msg, err := structpb.NewStruct(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	// Convert to params
	params, err := ProtobufToParams(msg, nil)
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Should return empty map, not nil
	if params == nil {
		t.Errorf("ProtobufToParams() returned nil, want empty map")
	}
	if len(params) != 0 {
		t.Errorf("ProtobufToParams() returned map with %d entries, want 0", len(params))
	}
}

func TestProtobufToParams_NilMessage(t *testing.T) {
	// Convert nil message
	params, err := ProtobufToParams(nil, nil)
	if err != nil {
		t.Fatalf("ProtobufToParams() error = %v", err)
	}

	// Should return empty map, not nil
	if params == nil {
		t.Errorf("ProtobufToParams() returned nil, want empty map")
	}
	if len(params) != 0 {
		t.Errorf("ProtobufToParams() returned map with %d entries, want 0", len(params))
	}
}
