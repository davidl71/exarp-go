package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseExecutionResponse(t *testing.T) {
	body := []byte(`{
  "changes": [
    {"file": "foo.go", "location": "10", "old_content": "a", "new_content": "b"}
  ],
  "explanation": "done",
  "confidence": 0.9
}`)
	result, err := ParseExecutionResponse(body)
	if err != nil {
		t.Fatalf("ParseExecutionResponse: %v", err)
	}
	if len(result.Changes) != 1 {
		t.Errorf("len(Changes) = %d, want 1", len(result.Changes))
	}
	if result.Changes[0].File != "foo.go" || result.Changes[0].OldContent != "a" || result.Changes[0].NewContent != "b" {
		t.Errorf("change = %+v", result.Changes[0])
	}
	if result.Explanation != "done" || result.Confidence != 0.9 {
		t.Errorf("explanation=%q confidence=%f", result.Explanation, result.Confidence)
	}
}

func TestParseExecutionResponse_WithMarkdownBlock(t *testing.T) {
	body := []byte("```json\n{\"changes\":[],\"explanation\":\"\",\"confidence\":0.5}\n```")
	result, err := ParseExecutionResponse(body)
	if err != nil {
		t.Fatalf("ParseExecutionResponse: %v", err)
	}
	if result.Confidence != 0.5 {
		t.Errorf("confidence = %f, want 0.5", result.Confidence)
	}
}

func TestApplyChanges_ReplaceInFile(t *testing.T) {
	dir := t.TempDir()
	projectRoot := filepath.Join(dir, "proj")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(projectRoot, "a.go")
	orig := "package main\n\nvar x = 1\n"
	if err := os.WriteFile(f, []byte(orig), 0644); err != nil {
		t.Fatal(err)
	}

	changes := []FileChange{
		{File: "a.go", OldContent: "var x = 1", NewContent: "var x = 2"},
	}
	applied, err := ApplyChanges(projectRoot, changes)
	if err != nil {
		t.Fatalf("ApplyChanges: %v", err)
	}
	if len(applied) != 1 || applied[0] != "a.go" {
		t.Errorf("applied = %v", applied)
	}
	after, _ := os.ReadFile(f)
	if string(after) != "package main\n\nvar x = 2\n" {
		t.Errorf("file content = %q", string(after))
	}
}

func TestApplyChanges_NewFile(t *testing.T) {
	dir := t.TempDir()
	projectRoot := filepath.Join(dir, "proj")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}

	changes := []FileChange{
		{File: "sub/new.go", NewContent: "package sub\n"},
	}
	applied, err := ApplyChanges(projectRoot, changes)
	if err != nil {
		t.Fatalf("ApplyChanges: %v", err)
	}
	if len(applied) != 1 {
		t.Errorf("applied = %v", applied)
	}
	full := filepath.Join(projectRoot, "sub", "new.go")
	after, _ := os.ReadFile(full)
	if string(after) != "package sub\n" {
		t.Errorf("file content = %q", string(after))
	}
}

func TestApplyChanges_RejectsPathOutsideRoot(t *testing.T) {
	dir := t.TempDir()
	projectRoot := filepath.Join(dir, "proj")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}

	changes := []FileChange{
		{File: "../../etc/passwd", NewContent: "x"},
	}
	_, err := ApplyChanges(projectRoot, changes)
	if err == nil {
		t.Error("ApplyChanges: expected error for path outside root")
	}
}

func TestApplyChanges_EmptyOldContent_Overwrites(t *testing.T) {
	dir := t.TempDir()
	projectRoot := filepath.Join(dir, "proj")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(projectRoot, "b.go")
	if err := os.WriteFile(f, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	changes := []FileChange{
		{File: "b.go", NewContent: "new"},
	}
	_, err := ApplyChanges(projectRoot, changes)
	if err != nil {
		t.Fatalf("ApplyChanges: %v", err)
	}
	after, _ := os.ReadFile(f)
	if string(after) != "new" {
		t.Errorf("file content = %q", string(after))
	}
}

func TestApplyChanges_EmptyProjectRoot_ReturnsError(t *testing.T) {
	_, err := ApplyChanges("", []FileChange{{File: "x.go", NewContent: "y"}})
	if err == nil {
		t.Error("expected error for empty project root")
	}
}

func TestExecutionResult_JSONRoundTrip(t *testing.T) {
	res := &ExecutionResult{
		Changes: []FileChange{
			{File: "a.go", Location: "1", OldContent: "x", NewContent: "y"},
		},
		Explanation: "test",
		Confidence:  0.85,
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseExecutionResponse(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Changes) != 1 || parsed.Changes[0].File != "a.go" || parsed.Confidence != 0.85 {
		t.Errorf("parsed = %+v", parsed)
	}
}
