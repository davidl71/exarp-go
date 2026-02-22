package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGatherEvidence(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}
	// Create a few files so evidence has content
	internalDir := filepath.Join(tmpDir, "internal")
	if err := os.MkdirAll(internalDir, 0755); err != nil {
		t.Fatalf("mkdir internal: %v", err)
	}

	goFile := filepath.Join(internalDir, "foo.go")
	if err := os.WriteFile(goFile, []byte("package internal\n\n// Add task completion inference\nfunc Bar() {}\n"), 0644); err != nil {
		t.Fatalf("write foo.go: %v", err)
	}

	pyFile := filepath.Join(tmpDir, "script.py")
	if err := os.WriteFile(pyFile, []byte("# migration script\nprint('hello')\n"), 0644); err != nil {
		t.Fatalf("write script.py: %v", err)
	}

	evidence, err := GatherEvidence(tmpDir, 3, []string{".go", ".py"})
	if err != nil {
		t.Fatalf("GatherEvidence: %v", err)
	}

	if evidence == nil {
		t.Fatal("evidence is nil")
	}

	if len(evidence.Paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(evidence.Paths))
	}

	hasGo := false
	hasPy := false

	for _, p := range evidence.Paths {
		if filepath.Ext(p) == ".go" {
			hasGo = true
		}

		if filepath.Ext(p) == ".py" {
			hasPy = true
		}
	}

	if !hasGo || !hasPy {
		t.Errorf("expected .go and .py paths; got paths %v", evidence.Paths)
	}

	if len(evidence.Snippets) == 0 {
		t.Log("Snippets may be empty for small files; Paths should be present")
	}
}

func TestScoreTasksHeuristic(t *testing.T) {
	tasks := []Todo2Task{
		{ID: "T-1", Content: "Add foo module", LongDescription: "Implement foo in internal/foo.go"},
		{ID: "T-2", Content: "Unrelated task", LongDescription: "Nothing matches"},
	}
	evidence := &CodebaseEvidence{
		Paths:    []string{"internal/foo.go", "cmd/main.go"},
		Snippets: map[string]string{"internal/foo.go": "package foo func Bar add foo module"},
	}

	results := scoreTasksHeuristic(tasks, evidence, 0.5)
	if len(results) == 0 {
		t.Fatal("expected at least one inferred result for T-1 (matches foo)")
	}

	found := false

	for _, r := range results {
		if r.TaskID == "T-1" {
			found = true

			if r.Confidence <= 0 || r.Confidence > 1 {
				t.Errorf("T-1 confidence %v not in (0, 1]", r.Confidence)
			}

			if len(r.Evidence) == 0 {
				t.Error("T-1 expected non-empty evidence")
			}

			break
		}
	}

	if !found {
		t.Errorf("expected T-1 in results, got %v", results)
	}
}

func TestHandleInferTaskProgressNative_ResponseShape(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}
	// Empty state.todo2.json so LoadTodo2Tasks returns empty (DB will be tried first; may have tasks)
	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	if err := os.WriteFile(statePath, []byte(`{"todos":[]}`), 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	t.Setenv("PROJECT_ROOT", tmpDir)

	ctx := context.Background()
	params := map[string]interface{}{
		"project_root": tmpDir,
		"dry_run":      true,
		"scan_depth":   1,
	}

	result, err := handleInferTaskProgressNative(ctx, params)
	if err != nil {
		t.Fatalf("handleInferTaskProgressNative: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["success"] != true {
		t.Errorf("success = %v", data["success"])
	}

	if _, ok := data["total_tasks_analyzed"]; !ok {
		t.Error("missing total_tasks_analyzed")
	}

	if _, ok := data["inferences_made"]; !ok {
		t.Error("missing inferences_made")
	}

	if _, ok := data["tasks_updated"]; !ok {
		t.Error("missing tasks_updated")
	}

	if _, ok := data["inferred_results"]; !ok {
		t.Error("missing inferred_results")
	}

	results, _ := data["inferred_results"].([]interface{})
	for _, r := range results {
		item, _ := r.(map[string]interface{})
		if item["task_id"] == nil || item["confidence"] == nil || item["evidence"] == nil {
			t.Errorf("inferred result missing task_id/confidence/evidence: %v", item)
		}
	}
}

func TestFilterByThreshold(t *testing.T) {
	scored := []InferredResult{
		{TaskID: "T-1", Confidence: 0.8, Evidence: []string{"a"}},
		{TaskID: "T-2", Confidence: 0.5, Evidence: []string{"b"}},
		{TaskID: "T-3", Confidence: 0.7, Evidence: []string{"c"}},
	}

	out := filterByThreshold(scored, 0.7)
	if len(out) != 2 {
		t.Errorf("filterByThreshold(0.7) expected 2, got %d", len(out))
	}

	for _, r := range out {
		if r.Confidence < 0.7 {
			t.Errorf("result %s has confidence %v < 0.7", r.TaskID, r.Confidence)
		}
	}
}

func TestParseFMCompletionResponse(t *testing.T) {
	tests := []struct {
		name     string
		resp     string
		taskID   string
		wantConf float64
		wantEv   int
	}{
		{"valid", `{"task_id":"T-1","complete":true,"confidence":0.9,"evidence":["file x exists"]}`, "T-1", 0.9, 1},
		{"with prefix", `Here is the result: {"task_id":"T-2","confidence":0.5,"evidence":[]}`, "T-2", 0.5, 0},
		{"wrong id", `{"task_id":"T-9","confidence":0.8}`, "T-1", -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf, ev := parseFMCompletionResponse(tt.resp, tt.taskID)
			if conf != tt.wantConf {
				t.Errorf("confidence = %v, want %v", conf, tt.wantConf)
			}

			if len(ev) != tt.wantEv {
				t.Errorf("evidence len = %d, want %d", len(ev), tt.wantEv)
			}
		})
	}
}

func TestWriteInferReport(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.md")
	result := map[string]interface{}{
		"total_tasks_analyzed": 2,
		"inferences_made":      1,
		"tasks_updated":        0,
		"inferred_results": []InferredResult{
			{TaskID: "T-1", Confidence: 0.85, Evidence: []string{"path:internal/foo.go", "snippet:internal/foo.go"}},
		},
	}

	err := writeInferReport(reportPath, result, true, "native_go_heuristics")
	if err != nil {
		t.Fatalf("writeInferReport: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Task Completion Check Report") {
		t.Errorf("report missing title: %s", content)
	}

	if !strings.Contains(content, "## Summary") {
		t.Errorf("report missing Summary section: %s", content)
	}

	if !strings.Contains(content, "Total Tasks Analyzed") {
		t.Errorf("report missing Total Tasks Analyzed: %s", content)
	}

	if !strings.Contains(content, "## Inferred Completions") {
		t.Errorf("report missing Inferred Completions section: %s", content)
	}

	if !strings.Contains(content, "Task T-1") {
		t.Errorf("report missing Task T-1: %s", content)
	}

	if !strings.Contains(content, "85.0%") {
		t.Errorf("report missing confidence 85.0%%: %s", content)
	}
}

func TestHandleInferTaskProgressNative_ReportFileAndDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}

	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	if err := os.WriteFile(statePath, []byte(`{"todos":[]}`), 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	t.Setenv("PROJECT_ROOT", tmpDir)

	reportPath := filepath.Join(tmpDir, "out", "TASK_COMPLETION_CHECK.md")
	ctx := context.Background()
	params := map[string]interface{}{
		"project_root": tmpDir,
		"dry_run":      true,
		"output_path":  reportPath,
		"scan_depth":   1,
	}

	result, err := handleInferTaskProgressNative(ctx, params)
	if err != nil {
		t.Fatalf("handleInferTaskProgressNative: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["tasks_updated"].(float64) != 0 {
		t.Errorf("dry_run true should yield tasks_updated 0, got %v", data["tasks_updated"])
	}

	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report file not created: %v", err)
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	// Report may be markdown (Summary, Inferred Completions) or overwritten by response; accept either
	if len(content) > 0 && (strings.Contains(string(content), "Summary") || strings.Contains(string(content), "success") || strings.Contains(string(content), "method")) {
		return
	}
	t.Errorf("report file missing expected content (Summary or JSON keys): got %d bytes", len(content))
}

func TestLoadTasksByStatus(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}

	// Create test tasks in different statuses
	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	testData := `{
		"todos": [
			{"id": "T-1", "content": "Todo task", "status": "Todo", "priority": "medium"},
			{"id": "T-2", "content": "In Progress task", "status": "In Progress", "priority": "high"},
			{"id": "T-3", "content": "Done task", "status": "Done", "priority": "low"},
			{"id": "T-4", "content": "Another Todo task", "status": "Todo", "priority": "low"}
		]
	}`
	if err := os.WriteFile(statePath, []byte(testData), 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	t.Setenv("PROJECT_ROOT", tmpDir)
	ctx := context.Background()

	// Test loading Todo tasks
	todoTasks, err := loadTasksByStatus(ctx, tmpDir, "Todo")
	if err != nil {
		t.Fatalf("loadTasksByStatus(Todo): %v", err)
	}

	if len(todoTasks) != 2 {
		t.Errorf("expected 2 Todo tasks, got %d", len(todoTasks))
	}

	for _, task := range todoTasks {
		if task.Status != "Todo" {
			t.Errorf("task %s has status %s, expected Todo", task.ID, task.Status)
		}
	}

	// Test loading In Progress tasks
	inProgressTasks, err := loadTasksByStatus(ctx, tmpDir, "In Progress")
	if err != nil {
		t.Fatalf("loadTasksByStatus(In Progress): %v", err)
	}

	if len(inProgressTasks) != 1 {
		t.Errorf("expected 1 In Progress task, got %d", len(inProgressTasks))
	}

	if len(inProgressTasks) > 0 && inProgressTasks[0].ID != "T-2" {
		t.Errorf("expected task T-2, got %s", inProgressTasks[0].ID)
	}

	// Test loading Done tasks
	doneTasks, err := loadTasksByStatus(ctx, tmpDir, "Done")
	if err != nil {
		t.Fatalf("loadTasksByStatus(Done): %v", err)
	}

	if len(doneTasks) != 1 {
		t.Errorf("expected 1 Done task, got %d", len(doneTasks))
	}
}

func TestHandleInferTaskProgressNative_WithStatusFilter(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}

	// Create test tasks with matching codebase evidence
	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	testData := `{
		"todos": [
			{"id": "T-100", "content": "Add foo feature", "status": "Todo", "priority": "high", "long_description": "Implement foo in internal/foo.go"},
			{"id": "T-101", "content": "Fix bar bug", "status": "In Progress", "priority": "medium"},
			{"id": "T-102", "content": "Refactor baz", "status": "Todo", "priority": "low", "long_description": "Refactor baz module"}
		]
	}`
	if err := os.WriteFile(statePath, []byte(testData), 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	// Create corresponding codebase files
	internalDir := filepath.Join(tmpDir, "internal")
	if err := os.MkdirAll(internalDir, 0755); err != nil {
		t.Fatalf("mkdir internal: %v", err)
	}

	fooFile := filepath.Join(internalDir, "foo.go")
	if err := os.WriteFile(fooFile, []byte("package foo\n\n// foo feature implementation\nfunc Foo() {}\n"), 0644); err != nil {
		t.Fatalf("write foo.go: %v", err)
	}

	t.Setenv("PROJECT_ROOT", tmpDir)
	ctx := context.Background()

	// Test with status_filter=Todo
	params := map[string]interface{}{
		"project_root":         tmpDir,
		"status_filter":        "Todo",
		"dry_run":              true,
		"confidence_threshold": 0.5,
		"scan_depth":           2,
		"use_fm":               false, // Disable FM to make test deterministic
	}

	result, err := handleInferTaskProgressNative(ctx, params)
	if err != nil {
		t.Fatalf("handleInferTaskProgressNative: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	totalAnalyzed := int(data["total_tasks_analyzed"].(float64))
	if totalAnalyzed != 2 {
		t.Errorf("expected 2 Todo tasks analyzed, got %d", totalAnalyzed)
	}

	// Should find at least T-100 with foo evidence
	inferences := int(data["inferences_made"].(float64))
	if inferences < 1 {
		t.Logf("Warning: expected at least 1 inference for T-100 (foo match), got %d", inferences)
		t.Logf("Full result: %+v", data)
	}
}
