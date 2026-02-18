package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlanFile(t *testing.T) {
	planContent := `---
name: Test Plan
todos:
  - id: T-1
    content: First task
    status: pending
  - id: T-2
    content: Second task
    status: done
---
# Plan
- [ ] **First** (T-1)
- [x] **Second** (T-2)
`
	tmpDir := t.TempDir()

	planPath := filepath.Join(tmpDir, "test.plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	todos, checkboxState, err := parsePlanFile(planPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}

	if todos[0].ID != "T-1" || todos[0].Status != "pending" {
		t.Errorf("todos[0]: id=%s status=%s", todos[0].ID, todos[0].Status)
	}

	if todos[1].ID != "T-2" || todos[1].Status != "done" {
		t.Errorf("todos[1]: id=%s status=%s", todos[1].ID, todos[1].Status)
	}

	if !checkboxState["T-2"] {
		t.Error("T-2 should be checked")
	}

	if checkboxState["T-1"] {
		t.Error("T-1 should be unchecked")
	}
}

func TestHandleTaskWorkflowSyncFromPlanDryRun(t *testing.T) {
	planContent := `---
name: Sync Test
todos:
  - id: T-sync-test-1
    content: Plan task one
    status: pending
  - id: T-sync-test-2
    content: Plan task two
    status: done
---
# Plan
- [ ] **One** (T-sync-test-1)
- [x] **Two** (T-sync-test-2)
`
	tmpDir := t.TempDir()
	prevRoot := os.Getenv("PROJECT_ROOT")

	os.Setenv("PROJECT_ROOT", tmpDir)

	defer func() { os.Setenv("PROJECT_ROOT", prevRoot) }()
	// Ensure FindProjectRoot resolves so ValidatePlanningLink can use project root
	if _, err := FindProjectRoot(); err != nil {
		t.Skip("PROJECT_ROOT not set for test; skipping sync_from_plan integration test")
	}

	planPath := filepath.Join(tmpDir, "sync-test.plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}
	// .todo2 dir so LoadTodo2Tasks doesn't fail (can be empty)
	if err := os.MkdirAll(filepath.Join(tmpDir, ".todo2"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	// Use absolute path so plan file is found regardless of GetProjectRoot behavior
	absPlanPath, _ := filepath.Abs(planPath)
	params := map[string]interface{}{
		"action":       "sync_from_plan",
		"planning_doc": absPlanPath,
		"dry_run":      true,
		"write_plan":   false,
	}

	result, err := handleTaskWorkflowSyncFromPlan(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatal(err)
	}

	if data["success"] != true {
		t.Errorf("success: %v", data["success"])
	}

	if data["action"] != "sync_from_plan" {
		t.Errorf("action: %v", data["action"])
	}

	if data["dry_run"] != true {
		t.Errorf("dry_run: %v", data["dry_run"])
	}
}

func TestTodo2StatusToPlanStatus(t *testing.T) {
	tests := []struct {
		todo2 string
		plan  string
	}{
		{"Done", "done"},
		{"In Progress", "in_progress"},
		{"Review", "review"},
		{"Todo", "pending"},
		{"", "pending"},
	}
	for _, tt := range tests {
		got := todo2StatusToPlanStatus(tt.todo2)
		if got != tt.plan {
			t.Errorf("todo2StatusToPlanStatus(%q) = %q, want %q", tt.todo2, got, tt.plan)
		}
	}
}
