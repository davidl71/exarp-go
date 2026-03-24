package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

type taskCommandStubServer struct {
	lastTool string
	lastArgs map[string]interface{}
	result   []framework.TextContent
	err      error
}

func (s *taskCommandStubServer) RegisterTool(string, string, framework.ToolSchema, framework.ToolHandler) error {
	return nil
}

func (s *taskCommandStubServer) RegisterPrompt(string, string, framework.PromptHandler) error {
	return nil
}

func (s *taskCommandStubServer) RegisterResource(string, string, string, string, framework.ResourceHandler) error {
	return nil
}

func (s *taskCommandStubServer) RegisterResourceTemplate(string, string, string, string, framework.ResourceHandler) error {
	return nil
}

func (s *taskCommandStubServer) Run(context.Context, framework.Transport) error { return nil }

func (s *taskCommandStubServer) GetName() string { return "stub" }

func (s *taskCommandStubServer) CallTool(_ context.Context, name string, args json.RawMessage) ([]framework.TextContent, error) {
	s.lastTool = name
	_ = json.Unmarshal(args, &s.lastArgs)
	return s.result, s.err
}

func (s *taskCommandStubServer) ListTools() []framework.ToolInfo { return nil }

func TestHandleTaskStatusJSONWrapsSingleTask(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{
			Type: "text",
			Text: `{"success":true,"method":"list","tasks":[{"id":"T-123","status":"Done","content":"Fix CLI output","priority":"high"}]}`,
		}},
	}

	restore := setCLIOutputOptsForTest(true, false, false)
	defer restore()

	output := captureStdout(t, func() {
		if err := handleTaskStatus(server, []string{"T-123"}); err != nil {
			t.Fatalf("handleTaskStatus returned error: %v", err)
		}
	})

	if server.lastTool != "task_workflow" {
		t.Fatalf("tool = %q, want task_workflow", server.lastTool)
	}
	if got := server.lastArgs["action"]; got != "list" {
		t.Fatalf("action = %v, want list", got)
	}
	if got := server.lastArgs["task_id"]; got != "T-123" {
		t.Fatalf("task_id = %v, want T-123", got)
	}
	if got := server.lastArgs["output_format"]; got != "json" {
		t.Fatalf("output_format = %v, want json", got)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &data); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output)
	}
	if got := data["method"]; got != "status" {
		t.Fatalf("method = %v, want status", got)
	}

	task, _ := data["task"].(map[string]interface{})
	if got := task["id"]; got != "T-123" {
		t.Fatalf("task.id = %v, want T-123", got)
	}
	if got := task["status"]; got != "Done" {
		t.Fatalf("task.status = %v, want Done", got)
	}
}

func TestHandleTaskShowReturnsNotFound(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{Type: "text", Text: `{"success":true,"method":"list","tasks":[]}`}},
	}

	restore := setCLIOutputOptsForTest(false, false, false)
	defer restore()

	err := handleTaskShow(server, []string{"T-missing"})
	if err == nil {
		t.Fatal("expected error for missing task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %q, want contains not found", err.Error())
	}
}

func TestHandleTaskListParsedUsesListActionAndStatus(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{Type: "text", Text: `{"success":true,"method":"list","tasks":[]}`}},
	}

	restore := setCLIOutputOptsForTest(false, false, false)
	defer restore()

	parsed := mcpcli.ParseArgs([]string{"task", "list", "--status", "Todo", "--priority", "high", "--tag", "cli"})
	if err := handleTaskListParsed(server, parsed); err != nil {
		t.Fatalf("handleTaskListParsed() error = %v", err)
	}

	if server.lastTool != "task_workflow" {
		t.Fatalf("tool = %q, want task_workflow", server.lastTool)
	}
	if got := server.lastArgs["action"]; got != "list" {
		t.Fatalf("action = %v, want list", got)
	}
	if _, ok := server.lastArgs["sub_action"]; ok {
		t.Fatalf("sub_action = %v, want omitted", server.lastArgs["sub_action"])
	}
	if got := server.lastArgs["status"]; got != "Todo" {
		t.Fatalf("status = %v, want Todo", got)
	}
	if got := server.lastArgs["priority"]; got != "high" {
		t.Fatalf("priority = %v, want high", got)
	}
	if got := server.lastArgs["filter_tag"]; got != "cli" {
		t.Fatalf("filter_tag = %v, want cli", got)
	}
}

func TestHandleTaskShowTextPrintsStructuredFields(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{
			Type: "text",
			Text: `{"success":true,"method":"list","tasks":[{"id":"T-55","status":"In Progress","content":"Improve CLI","long_description":"Make task commands easier for agents","tags":["cli","agents"],"dependencies":["T-1"],"recommended_tools":["task_workflow","report"]}]}`,
		}},
	}

	restore := setCLIOutputOptsForTest(false, false, false)
	defer restore()

	output := captureStdout(t, func() {
		if err := handleTaskShow(server, []string{"T-55"}); err != nil {
			t.Fatalf("handleTaskShow returned error: %v", err)
		}
	})

	for _, want := range []string{
		"ID: T-55",
		"Status: In Progress",
		"Content: Improve CLI",
		"Description: Make task commands easier for agents",
		"Tags: cli, agents",
		"Dependencies: T-1",
		"Recommended Tools: task_workflow, report",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func setCLIOutputOptsForTest(jsonOut, quiet, concise bool) func() {
	prev := CLIOutputOpts
	CLIOutputOpts.JSON = jsonOut
	CLIOutputOpts.Quiet = quiet
	CLIOutputOpts.Concise = concise
	return func() {
		CLIOutputOpts = prev
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	return <-done
}

func TestHandleTaskUpdateParsedUsesTaskWorkflowForPositionalID(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{Type: "text", Text: `{"success":true,"updated_count":1}`}},
	}

	restore := setCLIOutputOptsForTest(false, false, false)
	defer restore()

	parsed := mcpcli.ParseArgs([]string{"task", "update", "T-3000001", "--new-status", "Done"})
	if err := handleTaskUpdateParsed(server, parsed); err != nil {
		t.Fatalf("handleTaskUpdateParsed() error = %v", err)
	}

	if server.lastTool != "task_workflow" {
		t.Fatalf("tool = %q, want task_workflow", server.lastTool)
	}
	if got := server.lastArgs["action"]; got != "update" {
		t.Fatalf("action = %v, want update", got)
	}
	if got := server.lastArgs["task_ids"]; got != "T-3000001" {
		t.Fatalf("task_ids = %v, want T-3000001", got)
	}
	if got := server.lastArgs["new_status"]; got != "Done" {
		t.Fatalf("new_status = %v, want Done", got)
	}
}

func TestHandleTaskUpdateParsedUsesTaskWorkflowForPriorityFlag(t *testing.T) {
	server := &taskCommandStubServer{
		result: []framework.TextContent{{Type: "text", Text: `{"success":true,"updated_count":1}`}},
	}

	restore := setCLIOutputOptsForTest(false, false, false)
	defer restore()

	parsed := mcpcli.ParseArgs([]string{"task", "update", "--ids", "T-3000002", "--new-priority", "high"})
	if err := handleTaskUpdateParsed(server, parsed); err != nil {
		t.Fatalf("handleTaskUpdateParsed() error = %v", err)
	}

	if server.lastTool != "task_workflow" {
		t.Fatalf("tool = %q, want task_workflow", server.lastTool)
	}
	if got := server.lastArgs["action"]; got != "update" {
		t.Fatalf("action = %v, want update", got)
	}
	if got := server.lastArgs["task_ids"]; got != "T-3000002" {
		t.Fatalf("task_ids = %v, want T-3000002", got)
	}
	if got := server.lastArgs["priority"]; got != "high" {
		t.Fatalf("priority = %v, want high", got)
	}
}
