// tui_mcp_adapter_test.go — Unit tests for TUI MCP adapter (listTasksByStatusViaMCP, loadScorecardForTUI).
// Uses a stub MCPServer that returns predefined JSON. See docs/TUI_MCP_ADAPTER_DESIGN.md "Testing Strategy".
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

// adapterStubServer is a minimal MCPServer that returns predefined CallTool results for adapter tests.
type adapterStubServer struct {
	callToolResult []framework.TextContent
	callToolErr    error
}

func (s *adapterStubServer) RegisterTool(string, string, framework.ToolSchema, framework.ToolHandler) error {
	return nil
}
func (s *adapterStubServer) RegisterPrompt(string, string, framework.PromptHandler) error {
	return nil
}
func (s *adapterStubServer) RegisterResource(string, string, string, string, framework.ResourceHandler) error {
	return nil
}
func (s *adapterStubServer) RegisterResourceTemplate(string, string, string, string, framework.ResourceHandler) error {
	return nil
}
func (s *adapterStubServer) Run(context.Context, framework.Transport) error { return nil }
func (s *adapterStubServer) GetName() string                                { return "adapter-stub" }
func (s *adapterStubServer) CallTool(_ context.Context, _ string, _ json.RawMessage) ([]framework.TextContent, error) {
	return s.callToolResult, s.callToolErr
}
func (s *adapterStubServer) ListTools() []framework.ToolInfo { return nil }

func TestTuiMcpAdapter_ListTasksByStatusViaMCP(t *testing.T) {
	ctx := context.Background()
	taskListJSON := `{"success":true,"method":"list","tasks":[{"id":"T-1","content":"Task One","status":"Todo","priority":"medium"},{"id":"T-2","content":"Task Two","status":"Todo"}]}`
	server := &adapterStubServer{
		callToolResult: []framework.TextContent{{Type: "text", Text: taskListJSON}},
	}

	tasks, err := listTasksByStatusViaMCP(ctx, server, "Todo")
	if err != nil {
		t.Fatalf("listTasksByStatusViaMCP: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].ID != "T-1" || tasks[0].Content != "Task One" || tasks[0].Status != "Todo" {
		t.Errorf("tasks[0] = %+v", tasks[0])
	}
	if tasks[1].ID != "T-2" || tasks[1].Content != "Task Two" {
		t.Errorf("tasks[1] = %+v", tasks[1])
	}
}

func TestTuiMcpAdapter_ListTasksByStatusViaMCP_EmptyList(t *testing.T) {
	ctx := context.Background()
	server := &adapterStubServer{
		callToolResult: []framework.TextContent{{Type: "text", Text: `{"success":true,"method":"list","tasks":[]}`}},
	}

	tasks, err := listTasksByStatusViaMCP(ctx, server, "Done")
	if err != nil {
		t.Fatalf("listTasksByStatusViaMCP: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("len(tasks) = %d, want 0", len(tasks))
	}
}

func TestTuiMcpAdapter_ListTasksByStatusViaMCP_Error(t *testing.T) {
	ctx := context.Background()
	server := &adapterStubServer{callToolErr: context.DeadlineExceeded}

	_, err := listTasksByStatusViaMCP(ctx, server, "Todo")
	if err == nil {
		t.Fatal("expected error from CallTool")
	}
}

func TestTuiMcpAdapter_LoadScorecardForTUI(t *testing.T) {
	ctx := context.Background()
	scorecardJSON := `{"formatted_text":"## Scorecard\nScore: 85","recommendations":["Run make test"],"overall_score":85,"blockers":[]}`
	server := &adapterStubServer{
		callToolResult: []framework.TextContent{{Type: "text", Text: scorecardJSON}},
	}

	text, recs, err := loadScorecardForTUI(ctx, server, "/tmp/proj", false)
	if err != nil {
		t.Fatalf("loadScorecardForTUI: %v", err)
	}
	if text != "## Scorecard\nScore: 85" {
		t.Errorf("text = %q", text)
	}
	if len(recs) != 1 || recs[0] != "Run make test" {
		t.Errorf("recommendations = %v", recs)
	}
}

func TestTuiMcpAdapter_LoadScorecardForTUI_FallbackNoFormattedText(t *testing.T) {
	ctx := context.Background()
	// Response without formatted_text triggers fallback built from overall_score/blockers/recommendations
	scorecardJSON := `{"recommendations":["Fix lint"],"overall_score":70,"blockers":["No tests"]}`
	server := &adapterStubServer{
		callToolResult: []framework.TextContent{{Type: "text", Text: scorecardJSON}},
	}

	text, recs, err := loadScorecardForTUI(ctx, server, "/tmp/proj", true)
	if err != nil {
		t.Fatalf("loadScorecardForTUI: %v", err)
	}
	if recs == nil || len(recs) != 1 || recs[0] != "Fix lint" {
		t.Errorf("recommendations = %v", recs)
	}
	// Fallback format: "Score: %.0f\n" then Blockers then Recommendations
	if text == "" {
		t.Error("expected non-empty fallback text")
	}
	if !strings.Contains(text, "Score: 70") {
		t.Errorf("text should contain 'Score: 70': %q", text)
	}
	if !strings.Contains(text, "Blockers:") || !strings.Contains(text, "No tests") {
		t.Errorf("text should contain Blockers section: %q", text)
	}
	if !strings.Contains(text, "Recommendations:") || !strings.Contains(text, "Fix lint") {
		t.Errorf("text should contain Recommendations section: %q", text)
	}
}

func TestTuiMcpAdapter_LoadScorecardForTUI_NilServer(t *testing.T) {
	ctx := context.Background()

	_, _, err := loadScorecardForTUI(ctx, nil, "/tmp/proj", false)
	if err == nil {
		t.Fatal("expected error for nil server")
	}
	if !strings.Contains(err.Error(), "no server") {
		t.Errorf("error = %q", err.Error())
	}
}
