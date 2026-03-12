package request

import (
	"testing"
)

func TestApplyDefaults_MissingKeys(t *testing.T) {
	params := map[string]interface{}{
		"action": "custom",
	}

	defaults := map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"output_format": "text",
	}

	ApplyDefaults(params, defaults)

	// Existing value should be preserved
	if params["action"] != "custom" {
		t.Errorf("ApplyDefaults() overrode existing value: params[action] = %v, want %q", params["action"], "custom")
	}

	// Missing keys should get defaults
	if params["sub_action"] != "list" {
		t.Errorf("ApplyDefaults() did not set default: params[sub_action] = %v, want %q", params["sub_action"], "list")
	}

	if params["output_format"] != "text" {
		t.Errorf("ApplyDefaults() did not set default: params[output_format] = %v, want %q", params["output_format"], "text")
	}
}

func TestApplyDefaults_EmptyString(t *testing.T) {
	params := map[string]interface{}{
		"action": "",        // Empty string - should be replaced
		"status": "active",  // Non-empty - should be preserved
	}

	defaults := map[string]interface{}{
		"action": "sync",
		"status": "inactive",
	}

	ApplyDefaults(params, defaults)

	// Empty string should be replaced
	if params["action"] != "sync" {
		t.Errorf("ApplyDefaults() did not replace empty string: params[action] = %v, want %q", params["action"], "sync")
	}

	// Non-empty value should be preserved
	if params["status"] != "active" {
		t.Errorf("ApplyDefaults() overrode non-empty value: params[status] = %v, want %q", params["status"], "active")
	}
}

func TestApplyDefaults_PreserveNonEmpty(t *testing.T) {
	params := map[string]interface{}{
		"action": "custom_action",
		"limit":  42,
		"flag":   true,
	}

	defaults := map[string]interface{}{
		"action": "default_action",
		"limit":  10,
		"flag":   false,
	}

	ApplyDefaults(params, defaults)

	// All existing values should be preserved
	if params["action"] != "custom_action" {
		t.Errorf("ApplyDefaults() overrode string: params[action] = %v, want %q", params["action"], "custom_action")
	}

	if params["limit"] != 42 {
		t.Errorf("ApplyDefaults() overrode int: params[limit] = %v, want %d", params["limit"], 42)
	}

	if params["flag"] != true {
		t.Errorf("ApplyDefaults() overrode bool: params[flag] = %v, want %v", params["flag"], true)
	}
}

func TestApplyDefaults_VariousTypes(t *testing.T) {
	params := map[string]interface{}{}

	defaults := map[string]interface{}{
		"string": "value",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"slice":  []interface{}{1, 2, 3},
		"map":    map[string]interface{}{"key": "value"},
	}

	ApplyDefaults(params, defaults)

	// Verify all types are applied correctly
	if params["string"] != "value" {
		t.Errorf("ApplyDefaults() string = %v, want %q", params["string"], "value")
	}

	if params["int"] != 42 {
		t.Errorf("ApplyDefaults() int = %v, want %d", params["int"], 42)
	}

	if params["float"] != 3.14 {
		t.Errorf("ApplyDefaults() float = %v, want %f", params["float"], 3.14)
	}

	if params["bool"] != true {
		t.Errorf("ApplyDefaults() bool = %v, want %v", params["bool"], true)
	}

	slice, ok := params["slice"].([]interface{})
	if !ok || len(slice) != 3 {
		t.Errorf("ApplyDefaults() slice = %v, want [1, 2, 3]", params["slice"])
	}

	m, ok := params["map"].(map[string]interface{})
	if !ok || m["key"] != "value" {
		t.Errorf("ApplyDefaults() map = %v, want map with key=value", params["map"])
	}
}

func TestApplyDefaults_NilParams(t *testing.T) {
	// Should not panic with nil params
	defaults := map[string]interface{}{
		"action": "sync",
	}

	// This should not panic
	ApplyDefaults(nil, defaults)
}

func TestApplyDefaults_EmptyDefaults(t *testing.T) {
	params := map[string]interface{}{
		"action": "custom",
	}

	// Empty defaults should not modify params
	ApplyDefaults(params, map[string]interface{}{})

	if params["action"] != "custom" {
		t.Errorf("ApplyDefaults() modified params with empty defaults: params[action] = %v, want %q", params["action"], "custom")
	}
}

func TestApplyDefaults_MixedEmptyAndNonEmpty(t *testing.T) {
	params := map[string]interface{}{
		"action": "",           // Empty - should be replaced
		"status": "active",      // Non-empty - should be preserved
		"limit":  0,            // Zero int - should be preserved (not empty string)
		// "format" is missing  // Missing - should get default
	}

	defaults := map[string]interface{}{
		"action": "sync",
		"status": "inactive",
		"limit":  10,
		"format": "json",
	}

	ApplyDefaults(params, defaults)

	// Empty string replaced
	if params["action"] != "sync" {
		t.Errorf("ApplyDefaults() did not replace empty string: params[action] = %v, want %q", params["action"], "sync")
	}

	// Non-empty string preserved
	if params["status"] != "active" {
		t.Errorf("ApplyDefaults() overrode non-empty: params[status] = %v, want %q", params["status"], "active")
	}

	// Zero int preserved (not treated as empty)
	if params["limit"] != 0 {
		t.Errorf("ApplyDefaults() overrode zero int: params[limit] = %v, want %d", params["limit"], 0)
	}

	// Missing key gets default
	if params["format"] != "json" {
		t.Errorf("ApplyDefaults() did not set default: params[format] = %v, want %q", params["format"], "json")
	}
}
