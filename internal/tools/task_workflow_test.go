package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
)

func TestHandleTaskWorkflowNative(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "sync action",
			params: map[string]interface{}{
				"action": "sync",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "approve action",
			params: map[string]interface{}{
				"action": "approve",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "create action",
			params: map[string]interface{}{
				"action":           "create",
				"name":             "Test Task",
				"long_description": "Test description",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "sync_from_plan action requires planning_doc",
			params: map[string]interface{}{
				"action": "sync_from_plan",
			},
			wantError: true,
		},
		{
			name: "request_approval requires task_id",
			params: map[string]interface{}{
				"action": "request_approval",
			},
			wantError: true,
		},
		{
			name: "apply_approval_result requires task_id",
			params: map[string]interface{}{
				"action": "apply_approval_result",
				"result": "approved",
			},
			wantError: true,
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: true,
		},
		// sync with sub_action list returns task list (no sync); request JSON for parseable response
		{
			name: "sync sub_action list",
			params: map[string]interface{}{
				"action":        "sync",
				"sub_action":    "list",
				"output_format": "json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result for sync list")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if _, ok := data["tasks"]; !ok {
					t.Error("expected tasks field in list result")
				}
			},
		},
		// cleanup with dry_run returns without error
		{
			name: "cleanup dry_run",
			params: map[string]interface{}{
				"action":  "cleanup",
				"dry_run": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result for cleanup dry_run")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if dr, ok := data["dry_run"].(bool); !ok || !dr {
					t.Error("expected dry_run=true in cleanup result")
				}
			},
		},
		// clarity analyzes tasks (no task_id required); request JSON for parseable response
		{
			name: "clarity action",
			params: map[string]interface{}{
				"action":        "clarity",
				"output_format": "json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result for clarity")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if _, ok := data["clarity_issues"]; !ok {
					t.Error("expected clarity_issues in clarity result")
				}
			},
		},
		// delete requires task_id or task_ids
		{
			name: "delete without task_ids",
			params: map[string]interface{}{
				"action": "delete",
			},
			wantError: true,
		},
		// update requires task_ids
		{
			name: "update without task_ids",
			params: map[string]interface{}{
				"action": "update",
			},
			wantError: true,
		},
		// update requires at least one of new_status or priority
		{
			name: "update with task_ids but no new_status or priority",
			params: map[string]interface{}{
				"action":   "update",
				"task_ids": []interface{}{"T-1"},
			},
			wantError: true,
		},
		// link_planning requires planning_doc or epic_id
		{
			name: "link_planning without planning_doc and epic_id",
			params: map[string]interface{}{
				"action": "link_planning",
			},
			wantError: true,
		},
		// link_planning requires task_id or task_ids
		{
			name: "link_planning without task_ids",
			params: map[string]interface{}{
				"action":       "link_planning",
				"planning_doc": "docs/plan.md",
			},
			wantError: true,
		},
	}

	// Test planning link with valid doc (T-1768320725711): create task, then link_planning
	t.Run("link_planning with valid planning_doc", func(t *testing.T) {
		planDir := filepath.Join(tmpDir, ".cursor", "plans")
		if err := os.MkdirAll(planDir, 0755); err != nil {
			t.Fatalf("mkdir plans: %v", err)
		}
		planPath := filepath.Join(planDir, "test.plan.md")
		if err := os.WriteFile(planPath, []byte("# Test Plan\n"), 0644); err != nil {
			t.Fatalf("write plan: %v", err)
		}
		planDoc := ".cursor/plans/test.plan.md"

		ctx := context.Background()
		// Create a task first (link_planning only applies to Todo/In Progress)
		createResult, err := handleTaskWorkflowNative(ctx, map[string]interface{}{
			"action":           "create",
			"name":             "Plan link test task",
			"long_description": "For link_planning test",
		})
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
		if len(createResult) == 0 {
			t.Fatal("create returned no result")
		}
		var createData map[string]interface{}
		if err := json.Unmarshal([]byte(createResult[0].Text), &createData); err != nil {
			t.Fatalf("create result JSON: %v", err)
		}
		taskObj, _ := createData["task"].(map[string]interface{})
		taskID, _ := taskObj["id"].(string)
		if taskID == "" {
			t.Fatal("create did not return task id")
		}

		// Link planning doc to the task
		linkResult, err := handleTaskWorkflowNative(ctx, map[string]interface{}{
			"action":       "link_planning",
			"planning_doc": planDoc,
			"task_id":      taskID,
		})
		if err != nil {
			t.Fatalf("link_planning: %v", err)
		}
		if len(linkResult) == 0 {
			t.Error("link_planning returned no result")
			return
		}
		var linkData map[string]interface{}
		if err := json.Unmarshal([]byte(linkResult[0].Text), &linkData); err != nil {
			t.Errorf("link_planning result JSON: %v", err)
			return
		}
		if updated, _ := linkData["updated_ids"].([]interface{}); len(updated) != 1 {
			t.Errorf("link_planning updated_ids = %v, want 1 item", updated)
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleTaskWorkflowNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskWorkflowNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTaskWorkflowApproveConfirmViaElicitation(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	t.Run("confirm_via_elicitation with mock decline returns cancelled", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), mcpframework.EliciterContextKey, &mockEliciter{Action: "decline"})
		params := map[string]interface{}{"action": "approve", "confirm_via_elicitation": true}
		result, err := handleTaskWorkflowApprove(ctx, params)
		if err != nil {
			t.Fatalf("handleTaskWorkflowApprove() err = %v", err)
		}
		if len(result) == 0 {
			t.Fatal("expected non-empty result")
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if cancelled, _ := data["cancelled"].(bool); !cancelled {
			t.Errorf("expected cancelled=true when user declines, got %v", data)
		}
	})

	t.Run("confirm_via_elicitation with no eliciter proceeds", func(t *testing.T) {
		ctx := context.Background()
		params := map[string]interface{}{"action": "approve", "confirm_via_elicitation": true}
		result, err := handleTaskWorkflowApprove(ctx, params)
		if err != nil {
			t.Fatalf("handleTaskWorkflowApprove() err = %v", err)
		}
		if len(result) == 0 {
			t.Fatal("expected non-empty result")
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if cancelled, _ := data["cancelled"].(bool); cancelled {
			t.Errorf("expected no cancellation when eliciter is nil, got cancelled=true")
		}
	})
}

func TestApprovalWorkflowActions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}
	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	stateJSON := []byte(`{"todos":[{"id":"T-test-approval","content":"Test Approval Task","long_description":"For approval flow test","status":"Review","priority":"medium"}]}`)
	if err := os.WriteFile(statePath, stateJSON, 0644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	ctx := context.Background()

	t.Run("sync_approvals returns list", func(t *testing.T) {
		params := map[string]interface{}{"action": "sync_approvals"}
		result, err := handleTaskWorkflowNative(ctx, params)
		if err != nil {
			t.Fatalf("sync_approvals: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("expected non-empty result")
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if _, ok := data["approval_requests"]; !ok {
			t.Errorf("expected approval_requests in result")
		}
	})

	t.Run("request_approval returns payload for task", func(t *testing.T) {
		params := map[string]interface{}{"action": "request_approval", "task_id": "T-test-approval"}
		result, err := handleTaskWorkflowNative(ctx, params)
		if err != nil {
			t.Fatalf("request_approval: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("expected non-empty result")
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if _, ok := data["approval_request"]; !ok {
			t.Errorf("expected approval_request in result")
		}
	})

	t.Run("apply_approval_result updates task", func(t *testing.T) {
		params := map[string]interface{}{"action": "apply_approval_result", "task_id": "T-test-approval", "result": "approved"}
		_, err := handleTaskWorkflowNative(ctx, params)
		if err != nil {
			t.Fatalf("apply_approval_result: %v", err)
		}
		tasks, _ := LoadTodo2Tasks(tmpDir)
		for _, tsk := range tasks {
			if tsk.ID == "T-test-approval" {
				if tsk.Status != "Done" {
					t.Errorf("expected task status Done after approve, got %s", tsk.Status)
				}
				return
			}
		}
		t.Error("task T-test-approval not found after apply_approval_result")
	})
}

func TestHandleTaskWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "sync action",
			params: map[string]interface{}{
				"action": "sync",
			},
			wantError: false,
		},
		{
			name: "approve action",
			params: map[string]interface{}{
				"action": "approve",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleTaskWorkflow(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskWorkflow() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestParseTaskIDsFromParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		want   []string
	}{
		{"empty", map[string]interface{}{}, nil},
		{"task_id single", map[string]interface{}{"task_id": "T-1"}, []string{"T-1"}},
		{"task_id trimmed", map[string]interface{}{"task_id": "  T-2  "}, []string{"T-2"}},
		{"task_ids comma", map[string]interface{}{"task_ids": "T-1,T-2,T-3"}, []string{"T-1", "T-2", "T-3"}},
		{"task_ids JSON array", map[string]interface{}{"task_ids": `["T-a","T-b"]`}, []string{"T-a", "T-b"}},
		{"task_ids slice", map[string]interface{}{"task_ids": []interface{}{"T-x", "T-y"}}, []string{"T-x", "T-y"}},
		{"task_id and task_ids dedupe", map[string]interface{}{"task_id": "T-1", "task_ids": "T-1,T-2"}, []string{"T-1", "T-2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTaskIDsFromParams(tt.params)
			if len(got) != len(tt.want) {
				t.Errorf("ParseTaskIDsFromParams() len = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseTaskIDsFromParams()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
