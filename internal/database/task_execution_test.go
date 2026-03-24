package database

import (
	"context"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestTaskExecutionLifecycle(t *testing.T) {
	initCommentsTestDB(t)

	task := &models.Todo2Task{
		ID:      "T-3000001",
		Content: "Execution lifecycle test",
		Status:  models.StatusTodo,
	}
	if err := CreateTask(context.Background(), task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	run := &TaskExecutionRun{
		TaskID:  task.ID,
		AgentID: "agent-test",
		Status:  "running",
		Summary: "started",
	}
	if err := StartTaskExecutionRun(context.Background(), run); err != nil {
		t.Fatalf("StartTaskExecutionRun() error = %v", err)
	}
	if run.RunID == "" {
		t.Fatal("expected run ID to be generated")
	}

	if err := AddTaskVerification(context.Background(), &TaskVerification{
		TaskID:  task.ID,
		RunID:   run.RunID,
		Kind:    "compile",
		Result:  "passed",
		Details: "go build succeeded",
	}); err != nil {
		t.Fatalf("AddTaskVerification() error = %v", err)
	}

	if err := AddTaskProgressEntry(context.Background(), &TaskProgressEntry{
		TaskID:        task.ID,
		RunID:         run.RunID,
		Summary:       "wired handlers",
		Files:         []string{"internal/tools/task_workflow_execution.go"},
		RemainingWork: "add report surface",
	}); err != nil {
		t.Fatalf("AddTaskProgressEntry() error = %v", err)
	}

	if err := EndTaskExecutionRun(
		context.Background(),
		run.RunID,
		"completed",
		"finished",
		[]string{"internal/tools/task_workflow_execution.go"},
		[]string{"go build ./internal/tools ./internal/database"},
		"",
	); err != nil {
		t.Fatalf("EndTaskExecutionRun() error = %v", err)
	}

	loadedRun, err := GetTaskExecutionRun(context.Background(), run.RunID)
	if err != nil {
		t.Fatalf("GetTaskExecutionRun() error = %v", err)
	}
	if loadedRun.Status != "completed" {
		t.Fatalf("loaded run status = %q, want completed", loadedRun.Status)
	}
	if loadedRun.EndedAt.IsZero() {
		t.Fatal("expected ended_at to be set")
	}

	runs, err := ListTaskExecutionRuns(context.Background(), task.ID, "", 10)
	if err != nil {
		t.Fatalf("ListTaskExecutionRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	verifications, err := ListTaskVerifications(context.Background(), task.ID, run.RunID, 10)
	if err != nil {
		t.Fatalf("ListTaskVerifications() error = %v", err)
	}
	if len(verifications) != 1 {
		t.Fatalf("expected 1 verification, got %d", len(verifications))
	}

	progressEntries, err := ListTaskProgressEntries(context.Background(), task.ID, run.RunID, 10)
	if err != nil {
		t.Fatalf("ListTaskProgressEntries() error = %v", err)
	}
	if len(progressEntries) != 1 {
		t.Fatalf("expected 1 progress entry, got %d", len(progressEntries))
	}
}
