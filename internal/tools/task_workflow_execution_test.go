package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

func TestHandleTaskWorkflowExecutionActions(t *testing.T) {
	cleanup := initSessionTestDB(t)
	defer cleanup()

	task := &models.Todo2Task{
		ID:      "T-3000101",
		Content: "Workflow execution test",
		Status:  models.StatusTodo,
	}
	if err := database.CreateTask(context.Background(), task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	startResult, err := handleTaskWorkflowNative(context.Background(), map[string]interface{}{
		"action":  "start_run",
		"task_id": task.ID,
		"summary": "starting work",
	})
	if err != nil {
		t.Fatalf("start_run error = %v", err)
	}

	var startPayload map[string]interface{}
	if err := json.Unmarshal([]byte(startResult[0].Text), &startPayload); err != nil {
		t.Fatalf("start_run unmarshal error = %v", err)
	}
	runMap, ok := startPayload["run"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected run payload, got %T", startPayload["run"])
	}
	runID, _ := runMap["run_id"].(string)
	if runID == "" {
		t.Fatal("expected run_id in start_run result")
	}

	if _, err := handleTaskWorkflowNative(context.Background(), map[string]interface{}{
		"action":  "verify",
		"task_id": task.ID,
		"run_id":  runID,
		"kind":    "compile",
		"result":  "passed",
		"details": "compile succeeded",
	}); err != nil {
		t.Fatalf("verify error = %v", err)
	}

	if _, err := handleTaskWorkflowNative(context.Background(), map[string]interface{}{
		"action":         "add_progress",
		"task_id":        task.ID,
		"run_id":         runID,
		"summary":        "implemented handlers",
		"remaining_work": "wire CLI output",
	}); err != nil {
		t.Fatalf("add_progress error = %v", err)
	}

	showResult, err := handleTaskWorkflowNative(context.Background(), map[string]interface{}{
		"action": "show_run",
		"run_id": runID,
	})
	if err != nil {
		t.Fatalf("show_run error = %v", err)
	}
	var showPayload map[string]interface{}
	if err := json.Unmarshal([]byte(showResult[0].Text), &showPayload); err != nil {
		t.Fatalf("show_run unmarshal error = %v", err)
	}
	if got := len(showPayload["verifications"].([]interface{})); got != 1 {
		t.Fatalf("expected 1 verification, got %d", got)
	}
	if got := len(showPayload["progress"].([]interface{})); got != 1 {
		t.Fatalf("expected 1 progress item, got %d", got)
	}

	primeResult, err := handleSessionPrime(context.Background(), map[string]interface{}{
		"include_tasks": false,
		"include_hints": false,
		"compact":       true,
	})
	if err != nil {
		t.Fatalf("handleSessionPrime error = %v", err)
	}
	var primePayload map[string]interface{}
	if err := json.Unmarshal([]byte(primeResult[0].Text), &primePayload); err != nil {
		t.Fatalf("prime unmarshal error = %v", err)
	}
	if _, ok := primePayload["active_runs"]; !ok {
		t.Fatal("expected active_runs in session prime result")
	}
}
