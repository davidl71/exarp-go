package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

// ──�� ParseSyncConflictStrategy ───────────────────────────────────────────────

func TestParseSyncConflictStrategy(t *testing.T) {
	tests := []struct {
		input    string
		want     SyncConflictStrategy
		wantErr  bool
	}{
		{"", StrategyDBWins, false},
		{"db_wins", StrategyDBWins, false},
		{"DB_WINS", StrategyDBWins, false},
		{"db", StrategyDBWins, false},
		{"json_wins", StrategyJSONWins, false},
		{"json", StrategyJSONWins, false},
		{"merge_prefer_db", StrategyMergePreferDB, false},
		{"merge_db", StrategyMergePreferDB, false},
		{"merge_prefer_json", StrategyMergePreferJSON, false},
		{"merge_json", StrategyMergePreferJSON, false},
		{"unknown_strategy", StrategyDBWins, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSyncConflictStrategy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSyncConflictStrategy(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSyncConflictStrategy(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ─── mergeTask ───────────────────────────────────────────────────────────────

func makeTask(id, content, status, assignedTo string) Todo2Task {
	t := Todo2Task{
		ID:       id,
		Content:  content,
		Status:   status,
		Priority: "medium",
		Tags:     []string{"tag1"},
		Metadata: map[string]interface{}{"source": "test"},
	}
	if assignedTo != "" {
		t.AssignedTo = assignedTo
	}
	models.SetContentHash(&t)
	return t
}

func TestMergeTask_DBWins(t *testing.T) {
	db := makeTask("T-1", "DB content", "Done", "alice")
	json := makeTask("T-1", "JSON content", "Todo", "")

	result := mergeTask(db, json, StrategyDBWins)
	if result.Content != "DB content" {
		t.Errorf("StrategyDBWins: Content = %q, want %q", result.Content, "DB content")
	}
	if result.Status != "Done" {
		t.Errorf("StrategyDBWins: Status = %q, want %q", result.Status, "Done")
	}
}

func TestMergeTask_JSONWins(t *testing.T) {
	db := makeTask("T-1", "DB content", "Done", "alice")
	json := makeTask("T-1", "JSON content", "Todo", "")

	result := mergeTask(db, json, StrategyJSONWins)
	if result.Content != "JSON content" {
		t.Errorf("StrategyJSONWins: Content = %q, want %q", result.Content, "JSON content")
	}
	if result.Status != "Todo" {
		t.Errorf("StrategyJSONWins: Status = %q, want %q", result.Status, "Todo")
	}
}

func TestMergeTask_MergePreferDB_FillsEmptyFields(t *testing.T) {
	db := makeTask("T-1", "DB content", "Done", "alice")
	db.LongDescription = "" // empty — should be filled from JSON
	db.Tags = nil

	jt := makeTask("T-1", "JSON content", "Todo", "")
	jt.LongDescription = "JSON long desc"
	jt.Tags = []string{"from-json"}

	result := mergeTask(db, jt, StrategyMergePreferDB)
	// DB content wins (non-empty)
	if result.Content != "DB content" {
		t.Errorf("MergePreferDB: Content = %q, want %q", result.Content, "DB content")
	}
	// JSON fills empty LongDescription
	if result.LongDescription != "JSON long desc" {
		t.Errorf("MergePreferDB: LongDescription = %q, want %q", result.LongDescription, "JSON long desc")
	}
	// JSON fills empty Tags
	if len(result.Tags) != 1 || result.Tags[0] != "from-json" {
		t.Errorf("MergePreferDB: Tags = %v, want [from-json]", result.Tags)
	}
	// DB status preserved
	if result.Status != "Done" {
		t.Errorf("MergePreferDB: Status = %q, want %q", result.Status, "Done")
	}
}

func TestMergeTask_MergePreferDB_DoesNotOverwriteNonEmpty(t *testing.T) {
	db := makeTask("T-1", "DB content", "Done", "alice")
	db.LongDescription = "DB long desc"

	jt := makeTask("T-1", "JSON content", "Todo", "")
	jt.LongDescription = "JSON long desc"

	result := mergeTask(db, jt, StrategyMergePreferDB)
	// DB LongDescription must not be overwritten
	if result.LongDescription != "DB long desc" {
		t.Errorf("MergePreferDB: LongDescription should not be overwritten; got %q", result.LongDescription)
	}
}

func TestMergeTask_MergePreferJSON_PreservesDBRuntimeState(t *testing.T) {
	db := makeTask("T-1", "DB content", "Done", "alice")
	db.CompletedAt = "2025-01-01T00:00:00Z"
	db.LastModified = "2025-06-01T12:00:00Z"

	jt := makeTask("T-1", "JSON content", "Todo", "")

	result := mergeTask(db, jt, StrategyMergePreferJSON)
	// JSON content wins
	if result.Content != "JSON content" {
		t.Errorf("MergePreferJSON: Content = %q, want %q", result.Content, "JSON content")
	}
	// DB runtime state preserved
	if result.Status != "Done" {
		t.Errorf("MergePreferJSON: Status = %q, want %q", result.Status, "Done")
	}
	if result.CompletedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("MergePreferJSON: CompletedAt = %q, want %q", result.CompletedAt, "2025-01-01T00:00:00Z")
	}
	if result.AssignedTo != "alice" {
		t.Errorf("MergePreferJSON: AssignedTo = %q, want %q", result.AssignedTo, "alice")
	}
	if result.LastModified != "2025-06-01T12:00:00Z" {
		t.Errorf("MergePreferJSON: LastModified = %q, want %q", result.LastModified, "2025-06-01T12:00:00Z")
	}
}

// ─── mergeMetadata ───────────────────────────────────────────────────────────

func TestMergeMetadata(t *testing.T) {
	base := map[string]interface{}{"a": 1, "b": "base"}
	override := map[string]interface{}{"b": "override", "c": true}

	result := mergeMetadata(base, override)
	if result["a"] != 1 {
		t.Errorf("mergeMetadata: key a = %v, want 1", result["a"])
	}
	if result["b"] != "override" {
		t.Errorf("mergeMetadata: key b = %v, want 'override'", result["b"])
	}
	if result["c"] != true {
		t.Errorf("mergeMetadata: key c = %v, want true", result["c"])
	}
}

func TestMergeMetadata_BothNil(t *testing.T) {
	result := mergeMetadata(nil, nil)
	if result != nil {
		t.Errorf("mergeMetadata(nil, nil) = %v, want nil", result)
	}
}

func TestMergeMetadata_BaseNil(t *testing.T) {
	override := map[string]interface{}{"k": "v"}
	result := mergeMetadata(nil, override)
	if result["k"] != "v" {
		t.Errorf("mergeMetadata(nil, override): key k = %v, want v", result["k"])
	}
}

// ─── ImportFromJSON ───────────────────────────────────────────────────────────

func TestImportFromJSON_NewTasks(t *testing.T) {
	// Build a valid JSON task file.
	tasks := []Todo2Task{
		{
			ID:      "T-import-1",
			Content: "Imported task 1",
			Status:  "Todo",
		},
		{
			ID:      "T-import-2",
			Content: "Imported task 2",
			Status:  "Todo",
		},
	}
	state := models.Todo2State{Todos: tasks}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "import.json")
	if err := os.WriteFile(jsonFile, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Use a project root without a DB (JSON-only mode).
	projectRoot := tmpDir
	report, err := ImportFromJSON(projectRoot, jsonFile, StrategyDBWins)
	if err != nil {
		t.Fatalf("ImportFromJSON: %v", err)
	}

	// Both tasks should be reported as created/json-only.
	if report.Created+report.JSONOnly == 0 {
		t.Errorf("expected Created or JSONOnly > 0, got Created=%d JSONOnly=%d", report.Created, report.JSONOnly)
	}
	if len(report.Errors) > 0 {
		t.Logf("non-fatal errors: %v", report.Errors)
	}
}

func TestImportFromJSON_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := ImportFromJSON(tmpDir, filepath.Join(tmpDir, "nonexistent.json"), StrategyDBWins)
	if err == nil {
		t.Error("ImportFromJSON should fail for missing file")
	}
}

// ─── ParseSyncConflictStrategy round-trip ────────────────────────────────────

func TestSyncConflictStrategy_String(t *testing.T) {
	cases := []struct {
		s    SyncConflictStrategy
		want string
	}{
		{StrategyDBWins, "db_wins"},
		{StrategyJSONWins, "json_wins"},
		{StrategyMergePreferDB, "merge_prefer_db"},
		{StrategyMergePreferJSON, "merge_prefer_json"},
	}
	for _, c := range cases {
		if c.s.String() != c.want {
			t.Errorf("String() = %q, want %q", c.s.String(), c.want)
		}
	}
}
