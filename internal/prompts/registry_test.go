package prompts

import (
	"context"
	"testing"

	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestRegisterAllPrompts(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllPrompts(server)
	if err != nil {
		t.Fatalf("RegisterAllPrompts() error = %v", err)
	}

	// Verify all 34 prompts are registered (18 original + 16 migrated from Python)
	if server.PromptCount() != 34 {
		t.Errorf("server.PromptCount() = %v, want 34", server.PromptCount())
	}

	// Verify specific prompts are registered
	expectedPrompts := []string{
		"align", "discover", "config", "scan",
		"scorecard", "overview", "dashboard", "remember",
		"daily_checkin", "sprint_start", "sprint_end",
		"pre_sprint", "post_impl", "sync", "dups",
		"context", "mode", "task_update",
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health", "automation_setup",
		"advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
	}

	for _, promptName := range expectedPrompts {
		prompt, exists := server.GetPrompt(promptName)
		if !exists {
			t.Errorf("prompt %q not registered", promptName)
			continue
		}
		if prompt.Name != promptName {
			t.Errorf("prompt.Name = %v, want %v", prompt.Name, promptName)
		}
		if prompt.Description == "" {
			t.Errorf("prompt %q description is empty", promptName)
		}
	}
}

func TestRegisterAllPrompts_HandlerCreation(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllPrompts(server)
	if err != nil {
		t.Fatalf("RegisterAllPrompts() error = %v", err)
	}

	// Verify handlers are created for each prompt
	expectedPrompts := []string{
		"align", "discover", "config", "scan",
		"scorecard", "overview", "dashboard", "remember",
		"daily_checkin", "sprint_start", "sprint_end",
		"pre_sprint", "post_impl", "sync", "dups",
		"context", "mode", "task_update",
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health", "automation_setup",
		"advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
	}

	for _, promptName := range expectedPrompts {
		prompt, exists := server.GetPrompt(promptName)
		if !exists {
			continue
		}
		if prompt.Handler == nil {
			t.Errorf("prompt %q handler is nil", promptName)
		}
	}
}

func TestRegisterAllPrompts_DuplicatePrompt(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	// Register a prompt manually first
	err := server.RegisterPrompt(
		"align",
		"test description",
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("pre-register prompt error = %v", err)
	}

	// Now try to register all prompts - should fail on duplicate
	err = RegisterAllPrompts(server)
	// This will fail because align is already registered
	if err == nil {
		t.Error("RegisterAllPrompts() error = nil, want error for duplicate prompt")
	}
}
