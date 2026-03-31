// Unit tests for task_execute with mocked model (Phase 6 MODEL_ASSISTED_WORKFLOW).
package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

// mockExecutionRouter returns a fixed execution JSON so RunTaskExecutionFlow can be tested without a real LLM.
type mockExecutionRouter struct {
	response string
}

func (m *mockExecutionRouter) SelectModel(taskType string, _ ModelRequirements) ModelType {
	return ModelFM
}

func (m *mockExecutionRouter) Generate(_ context.Context, _ ModelType, _ string, _ int, _ float32) (string, error) {
	return m.response, nil
}

func TestRunTaskExecutionFlow_WithMockRouter(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:         "Unit test task",
		LongDescription: "Description for mock execution.",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"test"},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	mockJSON := `{"changes":[],"explanation":"Mock execution result.","confidence":0.85}`
	router := &mockExecutionRouter{response: mockJSON}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:      task.ID,
		ProjectRoot: projectRoot,
		Apply:       false,
		ModelRouter: router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow with mock: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil RunTaskExecutionFlowResult")
	}
	if result.TaskID != task.ID {
		t.Errorf("TaskID = %q, want %q", result.TaskID, task.ID)
	}
	if result.Explanation != "Mock execution result." {
		t.Errorf("Explanation = %q", result.Explanation)
	}
	if result.Confidence != 0.85 {
		t.Errorf("Confidence = %f, want 0.85", result.Confidence)
	}
	if result.ParseError != "" {
		t.Errorf("ParseError = %q", result.ParseError)
	}
}

func TestRunTaskExecutionFlow_WithMockRouter_InvalidJSON(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:  "Task",
		Status:   "Todo",
		Priority: "medium",
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	router := &mockExecutionRouter{response: "not valid json"}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:      task.ID,
		ProjectRoot: projectRoot,
		Apply:       false,
		ModelRouter: router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow: %v", err)
	}

	if result.ParseError == "" {
		t.Error("expected ParseError when model returns invalid JSON")
	}
}

func TestRunTaskExecutionFlow_ApplyTrue(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:         "Apply test task",
		LongDescription: "Task that should apply file changes.",
		Status:          "Todo",
		Priority:        "medium",
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	targetFile := filepath.Join(projectRoot, "output.txt")

	mockJSON := `{
		"changes":[{"file":"output.txt","location":"","old_content":"","new_content":"hello world"}],
		"explanation":"Created output file.",
		"confidence":0.9
	}`
	router := &mockExecutionRouter{response: mockJSON}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:        task.ID,
		ProjectRoot:   projectRoot,
		Apply:         true,
		MinConfidence: 0.5,
		ModelRouter:   router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow: %v", err)
	}

	if len(result.Applied) != 1 {
		t.Fatalf("Applied = %v, want 1 file", result.Applied)
	}
	if result.Applied[0] != "output.txt" {
		t.Errorf("Applied[0] = %q, want output.txt", result.Applied[0])
	}
	if result.ApplyError != "" {
		t.Errorf("unexpected ApplyError = %q", result.ApplyError)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("file content = %q, want %q", string(data), "hello world")
	}
}

func TestRunTaskExecutionFlow_LowConfidence(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:  "Low confidence task",
		Status:   "Todo",
		Priority: "medium",
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	mockJSON := `{"changes":[{"file":"should_not_exist.txt","location":"","old_content":"","new_content":"data"}],"explanation":"Low confidence.","confidence":0.2}`
	router := &mockExecutionRouter{response: mockJSON}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:        task.ID,
		ProjectRoot:   projectRoot,
		Apply:         true,
		MinConfidence: 0.5,
		ModelRouter:   router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow: %v", err)
	}

	if len(result.Applied) != 0 {
		t.Errorf("Applied = %v, want empty (confidence too low)", result.Applied)
	}
	if result.Confidence != 0.2 {
		t.Errorf("Confidence = %f, want 0.2", result.Confidence)
	}

	targetFile := filepath.Join(projectRoot, "should_not_exist.txt")
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Error("file should NOT have been created when confidence < MinConfidence")
	}
}

func TestRunTaskExecutionFlow_EmptyTaskID(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	router := &mockExecutionRouter{response: `{"changes":[],"explanation":"noop","confidence":1.0}`}

	_, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:      "",
		ProjectRoot: projectRoot,
		Apply:       false,
		ModelRouter: router,
	})
	if err == nil {
		t.Fatal("expected error for empty TaskID")
	}
}

func TestRunTaskExecutionFlow_DefaultMinConfidencePreventsApply(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:  "Default confidence task",
		Status:   "Todo",
		Priority: "medium",
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	router := &mockExecutionRouter{response: `{"changes":[{"file":"should_not_exist.txt","new_content":"data"}],"explanation":"Low confidence.","confidence":0.2}`}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:      task.ID,
		ProjectRoot: projectRoot,
		Apply:       true,
		ModelRouter: router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow: %v", err)
	}

	if len(result.Applied) != 0 {
		t.Fatalf("Applied = %v, want no files applied", result.Applied)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "should_not_exist.txt")); !os.IsNotExist(err) {
		t.Fatal("expected low-confidence change to be skipped by default threshold")
	}
}

func TestRunTaskExecutionFlow_AddsResultComment(t *testing.T) {
	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	sessionTestDBMu.Lock()
	if err := database.Init(projectRoot); err != nil {
		sessionTestDBMu.Unlock()
		t.Fatalf("database.Init: %v", err)
	}
	defer func() {
		_ = database.Close()
		sessionTestDBMu.Unlock()
	}()

	task := &database.Todo2Task{
		Content:  "Commented task",
		Status:   "Todo",
		Priority: "medium",
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	router := &mockExecutionRouter{response: `{"changes":[{"file":"result.txt","new_content":"done"}],"explanation":"Created a result file.","confidence":0.9}`}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:        task.ID,
		ProjectRoot:   projectRoot,
		Apply:         true,
		MinConfidence: 0.5,
		ModelRouter:   router,
	})
	if err != nil {
		t.Fatalf("RunTaskExecutionFlow: %v", err)
	}
	if !result.CommentAdded {
		t.Fatal("expected execution result comment to be added")
	}

	comments, err := database.GetComments(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("len(comments) = %d, want 1", len(comments))
	}
	if comments[0].Type != database.CommentTypeResult {
		t.Fatalf("comment type = %q, want %q", comments[0].Type, database.CommentTypeResult)
	}
	if !strings.Contains(comments[0].Content, "Created a result file.") {
		t.Fatalf("comment content missing explanation: %q", comments[0].Content)
	}
	if !strings.Contains(comments[0].Content, "result.txt") {
		t.Fatalf("comment content missing applied file list: %q", comments[0].Content)
	}
}
