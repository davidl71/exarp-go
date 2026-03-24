package resources

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

func initResourceTestDB(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".todo2"), 0o755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}
	if err := database.Init(dir); err != nil {
		t.Fatalf("database.Init: %v", err)
	}
	oldRoot := os.Getenv("PROJECT_ROOT")
	_ = os.Setenv("PROJECT_ROOT", dir)
	return func() {
		_ = os.Setenv("PROJECT_ROOT", oldRoot)
		_ = database.Close()
	}
}

func TestHandleAgentBriefing(t *testing.T) {
	cleanup := initResourceTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := &models.Todo2Task{
		ID:       "T-5100001",
		Content:  "Planner briefing task",
		Status:   models.StatusTodo,
		Priority: "high",
		Metadata: map[string]interface{}{"agent_role": "planner", "recommended_tools": []interface{}{"task_workflow", "report"}},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	data, mimeType, err := handleAgentBriefing(ctx, "stdio://agent/briefing")
	if err != nil {
		t.Fatalf("handleAgentBriefing: %v", err)
	}
	if mimeType != "application/json" {
		t.Fatalf("mimeType = %q, want application/json", mimeType)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["resource_kind"] != "agent_briefing" {
		t.Fatalf("resource_kind = %v, want agent_briefing", payload["resource_kind"])
	}
	if _, ok := payload["suggested_next"]; !ok {
		t.Fatalf("missing suggested_next: %v", payload)
	}
	if _, ok := payload["delegation_suggestions"]; !ok {
		t.Fatalf("missing delegation_suggestions: %v", payload)
	}
}

func TestHandleAgentExecutionPack(t *testing.T) {
	cleanup := initResourceTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := &models.Todo2Task{
		ID:              "T-5100002",
		Content:         "Reviewer execution task",
		LongDescription: "Review the latest execution evidence",
		Status:          models.StatusReview,
		Priority:        "medium",
		Metadata:        map[string]interface{}{"agent_role": "reviewer", "recommended_tools": []interface{}{"task_workflow", "report"}},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	run := &database.TaskExecutionRun{
		TaskID:  task.ID,
		AgentID: "reviewer-agent",
		Host:    "test-host",
		Status:  "running",
		Summary: "Reviewing evidence",
	}
	if err := database.StartTaskExecutionRun(ctx, run); err != nil {
		t.Fatalf("StartTaskExecutionRun: %v", err)
	}

	data, _, err := handleAgentExecutionPack(ctx, "stdio://agent/task/"+task.ID+"/execution-pack")
	if err != nil {
		t.Fatalf("handleAgentExecutionPack: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["resource_kind"] != "task_execution_pack" {
		t.Fatalf("resource_kind = %v, want task_execution_pack", payload["resource_kind"])
	}
	if _, ok := payload["recommended_workflow"].([]interface{}); !ok {
		t.Fatalf("recommended_workflow missing or wrong type: %T", payload["recommended_workflow"])
	}
	if _, ok := payload["safe_next_actions"].([]interface{}); !ok {
		t.Fatalf("safe_next_actions missing or wrong type: %T", payload["safe_next_actions"])
	}
	if payload["agent_role"] != "reviewer" {
		t.Fatalf("agent_role = %v, want reviewer", payload["agent_role"])
	}
}

func TestHandleAgentAlerts(t *testing.T) {
	cleanup := initResourceTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := &models.Todo2Task{
		ID:       "T-5100003",
		Content:  "Review-ready task",
		Status:   models.StatusReview,
		Priority: "high",
		Metadata: map[string]interface{}{"agent_role": "reviewer"},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	expiringTask := &models.Todo2Task{
		ID:       "T-5100004",
		Content:  "Stale run task",
		Status:   models.StatusInProgress,
		Priority: "medium",
		Metadata: map[string]interface{}{"agent_role": "worker"},
	}
	if err := database.CreateTask(ctx, expiringTask); err != nil {
		t.Fatalf("CreateTask stale task: %v", err)
	}
	if _, err := database.ClaimTaskForAgent(ctx, expiringTask.ID, "worker-agent", time.Minute); err != nil {
		t.Fatalf("ClaimTaskForAgent: %v", err)
	}
	run := &database.TaskExecutionRun{
		TaskID:    expiringTask.ID,
		AgentID:   "worker-agent",
		Host:      "test-host",
		Status:    "running",
		Summary:   "Long-running worker task",
		StartedAt: time.Now().Add(-15 * time.Minute),
	}
	if err := database.StartTaskExecutionRun(ctx, run); err != nil {
		t.Fatalf("StartTaskExecutionRun: %v", err)
	}

	data, _, err := handleAgentAlerts(ctx, "stdio://agent/alerts")
	if err != nil {
		t.Fatalf("handleAgentAlerts: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["resource_kind"] != "agent_alerts" {
		t.Fatalf("resource_kind = %v, want agent_alerts", payload["resource_kind"])
	}
	if got := payload["review_ready_count"]; got != float64(1) {
		t.Fatalf("review_ready_count = %v, want 1", got)
	}
	if staleRuns, ok := payload["stale_runs"].([]interface{}); !ok || len(staleRuns) == 0 {
		t.Fatalf("expected stale_runs, got %v", payload["stale_runs"])
	}
}
