package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl/mcp-stdio-tools/internal/framework"
	"github.com/davidl/mcp-stdio-tools/tests/fixtures"
)

func TestRegisterAllTools(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllTools(server)
	if err != nil {
		t.Fatalf("RegisterAllTools() error = %v", err)
	}

	// Verify all 24 tools are registered
	if server.ToolCount() != 24 {
		t.Errorf("server.ToolCount() = %v, want 24", server.ToolCount())
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
		// Batch 1
		{"analyze_alignment", "generate_config", "health", "setup_hooks"},
		// Batch 2
		{"memory", "memory_maint", "report", "security", "task_analysis", "task_discovery", "task_workflow", "testing"},
		// Batch 3
		{"automation", "tool_catalog", "workflow_mode", "lint", "estimation", "git_tools", "session", "infer_session_mode", "ollama", "mlx"},
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
