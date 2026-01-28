package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestParseTasksFromJSON_ValidMetadata(t *testing.T) {
	data := []byte(`{"todos":[
		{"id":"T-1","content":"ok","status":"Todo","metadata":{"key":"value"}},
		{"id":"T-2","content":"ok2","status":"Todo"}
	]}`)
	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		t.Fatalf("ParseTasksFromJSON: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].Metadata == nil || tasks[0].Metadata["key"] != "value" {
		t.Errorf("task 0 metadata: got %v", tasks[0].Metadata)
	}
	if tasks[1].Metadata != nil && len(tasks[1].Metadata) > 0 {
		t.Errorf("task 1 metadata should be nil or empty, got %v", tasks[1].Metadata)
	}
}

func TestParseTasksFromJSON_InvalidMetadataCoercedToRaw(t *testing.T) {
	// metadata as plain string (invalid for map) -> coerced to {"raw": "..."}
	// RawMessage preserves the JSON encoding, so we get the quoted string
	data := []byte(`{"todos":[
		{"id":"T-1","content":"ok","status":"Todo","metadata":"Plan document goes here"}
	]}`)
	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		t.Fatalf("ParseTasksFromJSON: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	raw, ok := tasks[0].Metadata["raw"]
	if !ok {
		t.Fatalf("expected metadata.raw, got %v", tasks[0].Metadata)
	}
	s, ok := raw.(string)
	if !ok {
		t.Fatalf("metadata.raw not string: %T %v", raw, raw)
	}
	if s != `"Plan document goes here"` && s != "Plan document goes here" {
		t.Errorf("metadata.raw = %q", s)
	}
}

func TestParseTasksFromJSON_MalformedMetadataCoercedToRaw(t *testing.T) {
	// metadata malformed JSON -> coerced to {"raw": "..."}
	data := []byte(`{"todos":[
		{"id":"T-1","content":"ok","status":"Todo","metadata":"{invalid"}
	]}`)
	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		t.Fatalf("ParseTasksFromJSON: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	raw, ok := tasks[0].Metadata["raw"]
	if !ok {
		t.Fatalf("expected metadata.raw, got %v", tasks[0].Metadata)
	}
	s, ok := raw.(string)
	if !ok {
		t.Fatalf("metadata.raw not string: %T %v", raw, raw)
	}
	// RawMessage is the JSON value; string is stored as "{invalid" (quoted in JSON)
	if s != `"{invalid"` && s != "{invalid" {
		t.Errorf("metadata.raw = %q", s)
	}
}

func TestLoadJSONStateFromContent_InvalidMetadata(t *testing.T) {
	data := []byte(`{"todos":[
		{"id":"T-1","content":"x","status":"Todo","metadata":"plain text"}
	]}`)
	tasks, _, err := LoadJSONStateFromContent(data)
	if err != nil {
		t.Fatalf("LoadJSONStateFromContent: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	raw, ok := tasks[0].Metadata["raw"]
	if !ok {
		t.Fatalf("expected metadata.raw, got %v", tasks[0].Metadata)
	}
	s, ok := raw.(string)
	if !ok {
		t.Fatalf("metadata.raw not string: %T %v", raw, raw)
	}
	if s != `"plain text"` && s != "plain text" {
		t.Errorf("metadata.raw = %q", s)
	}
}

func TestParseTasksFromJSON_EmptyTodos(t *testing.T) {
	data := []byte(`{"todos":[]}`)
	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		t.Fatalf("ParseTasksFromJSON: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("len(tasks) = %d, want 0", len(tasks))
	}
}

func TestParseTasksFromJSON_InvalidJSONFails(t *testing.T) {
	data := []byte(`{invalid`)
	_, err := ParseTasksFromJSON(data)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestParseTasksFromJSON_NameAndDescriptionAliases verifies that "name", "title", and "description" are read as content/long_description.
func TestParseTasksFromJSON_NameAndDescriptionAliases(t *testing.T) {
	data := []byte(`{"todos":[
		{"id":"T-1","name":"Task title","status":"Todo"},
		{"id":"T-2","content":"Title","description":"Full description here","status":"Todo"},
		{"id":"T-3","title":"Title only no content","status":"Todo"}
	]}`)
	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		t.Fatalf("ParseTasksFromJSON: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(tasks))
	}
	if tasks[0].Content != "Task title" {
		t.Errorf("task 0 content (from name): got %q", tasks[0].Content)
	}
	if tasks[1].Content != "Title" {
		t.Errorf("task 1 content: got %q", tasks[1].Content)
	}
	if tasks[1].LongDescription != "Full description here" {
		t.Errorf("task 1 long_description (from description): got %q", tasks[1].LongDescription)
	}
	if tasks[2].Content != "Title only no content" {
		t.Errorf("task 2 content (from title): got %q", tasks[2].Content)
	}
}

// TestMarshalTasksToStateJSON_IncludesNameAndDescription verifies written JSON has "name" and "description" for extension compatibility.
func TestMarshalTasksToStateJSON_IncludesNameAndDescription(t *testing.T) {
	tasks := []models.Todo2Task{
		{ID: "T-1", Content: "Short title", LongDescription: "Long description", Status: "Todo"},
	}
	data, err := MarshalTasksToStateJSON(tasks)
	if err != nil {
		t.Fatalf("MarshalTasksToStateJSON: %v", err)
	}
	var decoded struct {
		Todos []map[string]interface{} `json:"todos"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(decoded.Todos) != 1 {
		t.Fatalf("len(todos) = %d, want 1", len(decoded.Todos))
	}
	first := decoded.Todos[0]
	if name, _ := first["name"].(string); name != "Short title" {
		t.Errorf("name in JSON: got %q", name)
	}
	if desc, _ := first["description"].(string); desc != "Long description" {
		t.Errorf("description in JSON: got %q", desc)
	}
	if !strings.Contains(string(data), `"name"`) || !strings.Contains(string(data), `"description"`) {
		t.Errorf("JSON should contain \"name\" and \"description\" keys: %s", string(data))
	}
}
