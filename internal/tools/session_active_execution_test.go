package tools

import (
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
)

func TestBuildActiveExecutionSummary_singleLock(t *testing.T) {
	t.Parallel()

	now := time.Now().Add(30 * time.Minute)
	locks := []database.LockStatus{{
		TaskID:        "T-1",
		Assignee:      "general-host-999",
		LockUntil:     now,
		TimeRemaining: 30 * time.Minute,
	}}
	runs := []database.TaskExecutionRun{{
		RunID:     "R-1",
		TaskID:    "T-1",
		Status:    "running",
		AgentID:   "general-host-888",
		StartedAt: time.Now(),
	}}
	tasks := map[string]Todo2Task{"T-1": {ID: "T-1", Content: "Epic", Status: "In Progress", Priority: "high"}}

	summary := buildActiveExecutionSummary(locks, runs, tasks)
	if summary == nil {
		t.Fatal("expected summary for single lock")
	}

	if summary["task_id"] != "T-1" {
		t.Fatalf("task_id = %v", summary["task_id"])
	}

	if summary["content"] != "Epic" {
		t.Fatalf("content = %v", summary["content"])
	}

	if _, ok := summary["run"].(map[string]interface{}); !ok {
		t.Fatalf("expected run map, got %T", summary["run"])
	}
}
