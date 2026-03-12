package response

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatResult_Basic(t *testing.T) {
	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"data":    "test",
	}

	contents, err := FormatResult(result, "")
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil", err)
	}

	if len(contents) != 1 {
		t.Fatalf("FormatResult() returned %d contents, want 1", len(contents))
	}

	if contents[0].Type != "text" {
		t.Errorf("FormatResult() contents[0].Type = %q, want %q", contents[0].Type, "text")
	}

	if contents[0].Text == "" {
		t.Fatal("FormatResult() contents[0].Text is empty, want JSON string")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &parsed); err != nil {
		t.Fatalf("FormatResult() output is not valid JSON: %v", err)
	}

	if parsed["success"] != true {
		t.Errorf("FormatResult() parsed[success] = %v, want true", parsed["success"])
	}
}

func TestFormatResult_WithFile(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.json")

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
	}

	contents, err := FormatResult(result, outputPath)
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil", err)
	}

	// Verify file was written
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("FormatResult() did not create output file: %v", err)
	}

	// Verify file contents
	fileData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var fileResult map[string]interface{}
	if err := json.Unmarshal(fileData, &fileResult); err != nil {
		t.Fatalf("Output file is not valid JSON: %v", err)
	}

	if fileResult["success"] != true {
		t.Errorf("Output file result[success] = %v, want true", fileResult["success"])
	}

	// Verify output_path was added to result
	if result["output_path"] != outputPath {
		t.Errorf("FormatResult() result[output_path] = %q, want %q", result["output_path"], outputPath)
	}

	// Verify output_path is in the returned JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &parsed); err != nil {
		t.Fatalf("FormatResult() output is not valid JSON: %v", err)
	}

	if parsed["output_path"] != outputPath {
		t.Errorf("FormatResult() parsed[output_path] = %q, want %q", parsed["output_path"], outputPath)
	}
}

func TestFormatResult_FileWriteFailure(t *testing.T) {
	// Use invalid path (directory that doesn't exist)
	outputPath := "/nonexistent/directory/output.json"

	result := map[string]interface{}{
		"success": true,
	}

	// Should not fail even if file write fails
	contents, err := FormatResult(result, outputPath)
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil (file write failure should be ignored)", err)
	}

	if len(contents) != 1 {
		t.Fatalf("FormatResult() returned %d contents, want 1", len(contents))
	}

	// Verify result still formatted correctly
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &parsed); err != nil {
		t.Fatalf("FormatResult() output is not valid JSON: %v", err)
	}

	// output_path should not be in result if file write failed
	if _, exists := parsed["output_path"]; exists {
		t.Error("FormatResult() included output_path in result despite file write failure")
	}
}

func TestFormatResult_EmptyResult(t *testing.T) {
	result := map[string]interface{}{}

	contents, err := FormatResult(result, "")
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil", err)
	}

	if len(contents) != 1 {
		t.Fatalf("FormatResult() returned %d contents, want 1", len(contents))
	}

	// Verify it's valid JSON (empty object)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &parsed); err != nil {
		t.Fatalf("FormatResult() output is not valid JSON: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("FormatResult() parsed length = %d, want 0", len(parsed))
	}
}

func TestFormatResult_NestedStructures(t *testing.T) {
	result := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"nested": map[string]interface{}{
				"value": 42,
			},
		},
		"array": []interface{}{1, 2, 3},
	}

	contents, err := FormatResult(result, "")
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil", err)
	}

	// Verify nested structures are preserved
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(contents[0].Text), &parsed); err != nil {
		t.Fatalf("FormatResult() output is not valid JSON: %v", err)
	}

	nested, ok := parsed["data"].(map[string]interface{})
	if !ok {
		t.Fatal("FormatResult() did not preserve nested map structure")
	}

	inner, ok := nested["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("FormatResult() did not preserve deeply nested map structure")
	}

	if inner["value"].(float64) != 42 {
		t.Errorf("FormatResult() nested value = %v, want 42", inner["value"])
	}

	array, ok := parsed["array"].([]interface{})
	if !ok {
		t.Fatal("FormatResult() did not preserve array structure")
	}

	if len(array) != 3 {
		t.Errorf("FormatResult() array length = %d, want 3", len(array))
	}
}

func TestFormatResult_Indentation(t *testing.T) {
	result := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]interface{}{
			"nested": "value",
		},
	}

	contents, err := FormatResult(result, "")
	if err != nil {
		t.Fatalf("FormatResult() error = %v, want nil", err)
	}

	// Verify output is indented (contains newlines and spaces)
	text := contents[0].Text
	if !strings.Contains(text, "\n  ") {
		t.Error("FormatResult() output is not indented (should contain newlines and spaces)")
	}

	// Verify it's still valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("FormatResult() indented output is not valid JSON: %v", err)
	}
}
