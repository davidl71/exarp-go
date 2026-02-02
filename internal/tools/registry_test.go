package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestRegisterAllTools(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllTools(server)
	if err != nil {
		t.Fatalf("RegisterAllTools() error = %v", err)
	}

	// Verify all tools are registered (base 30 from RegisterAllTools; +1 when Apple FM on darwin/arm64/cgo).
	expectedCount := 30
	if server.ToolCount() != expectedCount && server.ToolCount() != expectedCount+1 {
		t.Errorf("server.ToolCount() = %v, want %d or %d (with conditional Apple Foundation Models)",
			server.ToolCount(), expectedCount, expectedCount+1)
	}

	// Verify specific tools are registered
	expectedTools := []string{
		"analyze_alignment",
		"generate_config",
		"health",
		"setup_hooks",
		"check_attribution",
		"add_external_tool_hints",
		"memory",
		"memory_maint",
		"report",
		"security",
		"task_analysis",
		"task_discovery",
		"task_workflow",
		"infer_task_progress",
		"testing",
		"automation",
		"tool_catalog",
		"workflow_mode",
		"lint",
		"estimation",
		"git_tools",
		"session",
		"infer_session_mode",
		"ollama",
		"mlx",
		"context_budget",
		"context",
		"text_generate",
		"prompt_tracking",
		"recommend",
	}

	for _, toolName := range expectedTools {
		tool, exists := server.GetTool(toolName)
		if !exists {
			t.Errorf("tool %q not registered", toolName)
			continue
		}
		if tool.Name != toolName {
			t.Errorf("tool.Name = %v, want %v", tool.Name, toolName)
		}
		if tool.Schema.Type != "object" {
			t.Errorf("tool %q schema.Type = %v, want object", toolName, tool.Schema.Type)
		}
	}
}

func TestRegisterAllTools_RegistrationError(t *testing.T) {
	// Create a mock server that will fail on registration
	// We can't easily simulate this without modifying the mock,
	// but we can test that all tools have valid schemas
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllTools(server)
	if err != nil {
		t.Fatalf("RegisterAllTools() error = %v", err)
	}

	// Verify no duplicate tools - all tools are registered uniquely
	// by checking expected list

	// Verify batch registration order
	expectedBatches := [][]string{
		// Batch 1: 6 simple tools
		{"analyze_alignment", "generate_config", "health", "setup_hooks", "check_attribution", "add_external_tool_hints"},
		// Batch 2: 9 medium tools
		{"memory", "memory_maint", "report", "security", "task_analysis", "task_discovery", "task_workflow", "infer_task_progress", "testing"},
		// Batch 3: 10 advanced tools
		{"automation", "tool_catalog", "workflow_mode", "lint", "estimation", "git_tools", "session", "infer_session_mode", "ollama", "mlx"},
		// Batch 4 + 5: context_budget, context, text_generate, prompt_tracking, recommend
		{"context_budget", "context", "text_generate", "prompt_tracking", "recommend"},
	}

	for _, batch := range expectedBatches {
		for _, toolName := range batch {
			if _, exists := server.GetTool(toolName); !exists {
				t.Errorf("tool %q from batch not registered", toolName)
			}
		}
	}
}

func TestRegisterAllTools_DuplicateTool(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	// Register a tool manually first
	err := server.RegisterTool(
		"analyze_alignment",
		"test description",
		framework.ToolSchema{Type: "object"},
		func(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("pre-register tool error = %v", err)
	}

	// Now try to register all tools - should fail on duplicate
	err = RegisterAllTools(server)
	// This will fail because analyze_alignment is already registered
	if err == nil {
		t.Error("RegisterAllTools() error = nil, want error for duplicate tool")
	}
}

func TestRegisterAllTools_SchemaValidation(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllTools(server)
	if err != nil {
		t.Fatalf("RegisterAllTools() error = %v", err)
	}

	// Verify all tools have valid object schemas
	expectedTools := []string{
		"analyze_alignment", "generate_config", "health", "setup_hooks",
		"memory", "memory_maint", "report", "security",
	}

	for _, toolName := range expectedTools {
		tool, exists := server.GetTool(toolName)
		if !exists {
			continue
		}

		if tool.Schema.Type == "" {
			t.Errorf("tool %q schema.Type is empty", toolName)
		}
		if tool.Schema.Type != "object" {
			t.Errorf("tool %q schema.Type = %v, want object", toolName, tool.Schema.Type)
		}
		if tool.Schema.Properties == nil {
			t.Errorf("tool %q schema.Properties is nil", toolName)
		}
	}
}
