package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestGetPromptTemplate(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
	}{
		{
			name:    "valid prompt - align",
			prompt:  "align",
			wantErr: false,
		},
		{
			name:    "valid prompt - discover",
			prompt:  "discover",
			wantErr: false,
		},
		{
			name:    "valid prompt - context",
			prompt:  "context",
			wantErr: false,
		},
		{
			name:    "valid prompt - mode",
			prompt:  "mode",
			wantErr: false,
		},
		{
			name:    "unknown prompt",
			prompt:  "unknown_prompt",
			wantErr: true,
		},
		{
			name:    "empty prompt name",
			prompt:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetPromptTemplate(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPromptTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == "" {
					t.Errorf("GetPromptTemplate() returned empty string for valid prompt")
				}
				if !strings.Contains(result, tt.prompt) && tt.prompt != "" {
					// Most prompts contain their key concept, but not all
					// Just verify it's not empty
				}
			}
		})
	}
}

func TestGetPromptTemplate_AllPrompts(t *testing.T) {
	// Test all templates (36 MCP prompts + 3 internal prompt optimization templates)
	allPrompts := []string{
		"align", "discover", "config", "scan",
		"scorecard", "overview", "plan", "dashboard", "remember",
		"daily_checkin", "sprint_start", "sprint_end",
		"pre_sprint", "post_impl", "sync", "dups",
		"context", "mode", "task_update",
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health", "automation_setup",
		"advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
		"tractatus_decompose",
		"prompt_optimization_analysis",
		"prompt_optimization_suggestions",
		"prompt_optimization_refinement",
		"task_breakdown",
	}

	for _, promptName := range allPrompts {
		t.Run(promptName, func(t *testing.T) {
			result, err := GetPromptTemplate(promptName)
			if err != nil {
				t.Errorf("GetPromptTemplate(%q) error = %v", promptName, err)
				return
			}
			if result == "" {
				t.Errorf("GetPromptTemplate(%q) returned empty string", promptName)
			}
			// Verify it's a reasonable length (at least 50 chars)
			if len(result) < 50 {
				t.Errorf("GetPromptTemplate(%q) returned suspiciously short result: %d chars", promptName, len(result))
			}
		})
	}
}

func TestGetPromptTemplate_PromptOptimizationAnalysis(t *testing.T) {
	template, err := GetPromptTemplate("prompt_optimization_analysis")
	if err != nil {
		t.Fatalf("GetPromptTemplate(prompt_optimization_analysis) error = %v", err)
	}
	if template == "" {
		t.Fatal("GetPromptTemplate returned empty string")
	}
	// Verify required content
	if !strings.Contains(template, "{prompt}") {
		t.Error("template missing {prompt} placeholder")
	}
	if !strings.Contains(template, "clarity") || !strings.Contains(template, "specificity") {
		t.Error("template missing dimension labels")
	}
	// Test substitution
	substituted := substituteTemplate(template, map[string]interface{}{
		"prompt":    "Fix the bug",
		"context":   "MCP server",
		"task_type": "code",
	})
	if !strings.Contains(substituted, "Fix the bug") {
		t.Error("substitution failed: prompt not replaced")
	}
	if !strings.Contains(substituted, "MCP server") {
		t.Error("substitution failed: context not replaced")
	}
}

func TestGetPromptTemplate_PromptOptimizationSuggestions(t *testing.T) {
	template, err := GetPromptTemplate("prompt_optimization_suggestions")
	if err != nil {
		t.Fatalf("GetPromptTemplate(prompt_optimization_suggestions) error = %v", err)
	}
	if template == "" {
		t.Fatal("GetPromptTemplate returned empty string")
	}
	for _, ph := range []string{"{prompt}", "{analysis}"} {
		if !strings.Contains(template, ph) {
			t.Errorf("template missing %q placeholder", ph)
		}
	}
	if !strings.Contains(template, "suggestions") {
		t.Error("template missing suggestions output format")
	}
	substituted := substituteTemplate(template, map[string]interface{}{
		"prompt":    "Do the thing",
		"analysis":  `{"clarity":0.6,"specificity":0.5}`,
		"context":   "Go project",
		"task_type": "code",
	})
	if !strings.Contains(substituted, "Do the thing") {
		t.Error("substitution failed: prompt not replaced")
	}
	if !strings.Contains(substituted, `{"clarity":0.6,"specificity":0.5}`) {
		t.Error("substitution failed: analysis not replaced")
	}
}

func TestGetPromptTemplate_TaskTypeVariants(t *testing.T) {
	// Verify all three optimization templates include task-type guidance
	for _, name := range []string{"prompt_optimization_analysis", "prompt_optimization_suggestions", "prompt_optimization_refinement"} {
		template, err := GetPromptTemplate(name)
		if err != nil {
			t.Errorf("GetPromptTemplate(%q) error = %v", name, err)
			continue
		}
		for _, want := range []string{"code", "docs", "general"} {
			if !strings.Contains(template, want) {
				t.Errorf("template %q missing task-type variant %q", name, want)
			}
		}
	}
}

func TestGetPromptTemplate_PromptOptimizationRefinement(t *testing.T) {
	template, err := GetPromptTemplate("prompt_optimization_refinement")
	if err != nil {
		t.Fatalf("GetPromptTemplate(prompt_optimization_refinement) error = %v", err)
	}
	if template == "" {
		t.Fatal("GetPromptTemplate returned empty string")
	}
	for _, ph := range []string{"{prompt}", "{suggestions}"} {
		if !strings.Contains(template, ph) {
			t.Errorf("template missing %q placeholder", ph)
		}
	}
	if !strings.Contains(template, "Preserve the original intent") {
		t.Error("template missing refinement instructions")
	}
	substituted := substituteTemplate(template, map[string]interface{}{
		"prompt":      "Fix it",
		"suggestions": `[{"dimension":"specificity","issue":"vague","recommendation":"Be concrete"}]`,
		"context":     "Go project",
		"task_type":   "code",
	})
	if !strings.Contains(substituted, "Fix it") {
		t.Error("substitution failed: prompt not replaced")
	}
	if !strings.Contains(substituted, "Be concrete") {
		t.Error("substitution failed: suggestions not replaced")
	}
}

func TestSubstituteTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     map[string]interface{}
		want     string
	}{
		{
			name:     "single string variable",
			template: "Hello {name}, welcome!",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Hello Alice, welcome!",
		},
		{
			name:     "multiple variables",
			template: "Hello {name}, you have {count} tasks.",
			args: map[string]interface{}{
				"name":  "Bob",
				"count": 5,
			},
			want: "Hello Bob, you have 5 tasks.",
		},
		{
			name:     "integer variable",
			template: "Task {id} is complete.",
			args:     map[string]interface{}{"id": 123},
			want:     "Task 123 is complete.",
		},
		{
			name:     "float variable",
			template: "Score: {score}%",
			args:     map[string]interface{}{"score": 85.5},
			want:     "Score: 85.5%",
		},
		{
			name:     "boolean variable",
			template: "Status: {enabled}",
			args:     map[string]interface{}{"enabled": true},
			want:     "Status: true",
		},
		{
			name:     "empty args - no substitution",
			template: "Hello {name}, welcome!",
			args:     map[string]interface{}{},
			want:     "Hello {name}, welcome!",
		},
		{
			name:     "nil args - no substitution",
			template: "Hello {name}, welcome!",
			args:     nil,
			want:     "Hello {name}, welcome!",
		},
		{
			name:     "missing variable - left as-is",
			template: "Hello {name}, you have {count} tasks.",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Hello Alice, you have {count} tasks.",
		},
		{
			name:     "unused args - ignored",
			template: "Hello {name}!",
			args: map[string]interface{}{
				"name":  "Alice",
				"extra": "ignored",
			},
			want: "Hello Alice!",
		},
		{
			name:     "special characters in value",
			template: "Message: {text}",
			args:     map[string]interface{}{"text": "Hello\nWorld & <tags>"},
			want:     "Message: Hello\nWorld & <tags>",
		},
		{
			name:     "multiple occurrences",
			template: "{name} says {name} is great!",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Alice says Alice is great!",
		},
		{
			name:     "no placeholders - no change",
			template: "This is a plain template.",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "This is a plain template.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteTemplate(tt.template, tt.args)
			if got != tt.want {
				t.Errorf("substituteTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreatePromptHandler_WithSubstitution(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllPrompts(server)
	if err != nil {
		t.Fatalf("RegisterAllPrompts() error = %v", err)
	}

	// Test handler with template substitution
	t.Run("handler with args", func(t *testing.T) {
		prompt, exists := server.GetPrompt("align")
		if !exists {
			t.Fatal("prompt 'align' not registered")
		}

		ctx := context.Background()
		args := map[string]interface{}{
			"task_id": "T-123",
			"count":   5,
		}

		result, err := prompt.Handler(ctx, args)
		if err != nil {
			t.Errorf("handler error = %v", err)
			return
		}

		// Verify result is not empty
		if result == "" {
			t.Error("handler returned empty string")
		}

		// If template had placeholders, they should be substituted
		// For now, just verify we get a result (align prompt doesn't have placeholders)
		if !strings.Contains(result, "Todo2") {
			t.Errorf("handler result doesn't contain expected content: %q", result)
		}
	})

	t.Run("handler without args", func(t *testing.T) {
		prompt, exists := server.GetPrompt("discover")
		if !exists {
			t.Fatal("prompt 'discover' not registered")
		}

		ctx := context.Background()
		args := map[string]interface{}{}

		result, err := prompt.Handler(ctx, args)
		if err != nil {
			t.Errorf("handler error = %v", err)
			return
		}

		if result == "" {
			t.Error("handler returned empty string")
		}

		if !strings.Contains(result, "task_discovery") {
			t.Errorf("handler result doesn't contain expected content: %q", result)
		}
	})
}

func TestCreatePromptHandler_AllPrompts(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllPrompts(server)
	if err != nil {
		t.Fatalf("RegisterAllPrompts() error = %v", err)
	}

	allPrompts := []string{
		"align", "discover", "config", "scan",
		"scorecard", "overview", "plan", "dashboard", "remember",
		"daily_checkin", "sprint_start", "sprint_end",
		"pre_sprint", "post_impl", "sync", "dups",
		"context", "mode", "task_update",
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health", "automation_setup",
		"advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
		"tractatus_decompose",
	}

	ctx := context.Background()

	for _, promptName := range allPrompts {
		t.Run(promptName, func(t *testing.T) {
			prompt, exists := server.GetPrompt(promptName)
			if !exists {
				t.Fatalf("prompt %q not registered", promptName)
			}

			if prompt.Handler == nil {
				t.Fatalf("prompt %q handler is nil", promptName)
			}

			// Test with empty args
			result, err := prompt.Handler(ctx, map[string]interface{}{})
			if err != nil {
				t.Errorf("handler error = %v", err)
				return
			}

			if result == "" {
				t.Errorf("handler returned empty string for prompt %q", promptName)
			}

			// Verify minimum length
			if len(result) < 50 {
				t.Errorf("handler returned suspiciously short result for prompt %q: %d chars", promptName, len(result))
			}
		})
	}
}

func TestSubstituteTemplate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     map[string]interface{}
		want     string
	}{
		{
			name:     "empty template",
			template: "",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "",
		},
		{
			name:     "template with only placeholders",
			template: "{a}{b}{c}",
			args: map[string]interface{}{
				"a": "1",
				"b": "2",
				"c": "3",
			},
			want: "123",
		},
		{
			name:     "nested braces - substitutes inner",
			template: "{{name}}",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "{Alice}", // Substitutes inner placeholder, leaves outer brace
		},
		{
			name:     "placeholder at start",
			template: "{name} is here",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Alice is here",
		},
		{
			name:     "placeholder at end",
			template: "Hello {name}",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Hello Alice",
		},
		{
			name:     "placeholder with spaces",
			template: "Hello { name }",
			args:     map[string]interface{}{"name": "Alice"},
			want:     "Hello { name }", // Spaces in placeholder name won't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteTemplate(tt.template, tt.args)
			if got != tt.want {
				t.Errorf("substituteTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
