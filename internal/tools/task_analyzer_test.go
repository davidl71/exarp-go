package tools

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type captureGenerator struct {
	supported bool
	output    string
	err       error
	prompt    string
}

func (g *captureGenerator) Supported() bool { return g.supported }

func (g *captureGenerator) Generate(_ context.Context, prompt string, _ int, _ float32) (string, error) {
	g.prompt = prompt
	return g.output, g.err
}

func TestAnalyzeTask(t *testing.T) {
	gen := &captureGenerator{
		supported: true,
		output:    `{"subtasks":[{"name":"Implement parser","description":"Add parser support","complexity":"medium"}]}`,
	}

	result, err := AnalyzeTask(context.Background(), "Build parser", "must pass tests", "existing CLI", gen)
	if err != nil {
		t.Fatalf("AnalyzeTask: %v", err)
	}
	if len(result.Subtasks) != 1 {
		t.Fatalf("len(Subtasks) = %d, want 1", len(result.Subtasks))
	}
	if result.Subtasks[0].Name != "Implement parser" {
		t.Fatalf("subtask name = %q", result.Subtasks[0].Name)
	}
	if !strings.Contains(gen.prompt, "Build parser") {
		t.Fatalf("prompt missing task description: %q", gen.prompt)
	}
	if !strings.Contains(gen.prompt, "must pass tests") {
		t.Fatalf("prompt missing acceptance criteria: %q", gen.prompt)
	}
	if !strings.Contains(gen.prompt, "existing CLI") {
		t.Fatalf("prompt missing context hint: %q", gen.prompt)
	}
}

func TestAnalyzeTask_Validation(t *testing.T) {
	if _, err := AnalyzeTask(context.Background(), "", "", "", &captureGenerator{supported: true}); err == nil {
		t.Fatal("expected error for empty task description")
	}

	if _, err := AnalyzeTask(context.Background(), "Task", "", "", nil); err == nil {
		t.Fatal("expected error for nil generator")
	}

	if _, err := AnalyzeTask(context.Background(), "Task", "", "", &captureGenerator{}); err == nil {
		t.Fatal("expected error for unsupported generator")
	}
}

func TestAnalyzeTask_GeneratorError(t *testing.T) {
	gen := &captureGenerator{
		supported: true,
		err:       errors.New("backend unavailable"),
	}

	_, err := AnalyzeTask(context.Background(), "Task", "", "", gen)
	if err == nil {
		t.Fatal("expected generator error")
	}
	if !strings.Contains(err.Error(), "generate breakdown") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseTaskBreakdown(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid json",
			text:    `{"subtasks":[{"name":"A","description":"Do A","complexity":"simple","dependencies":[]},{"name":"B","description":"Do B","complexity":"medium","dependencies":["A"]}]}`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "with prefix text",
			text:    `Here is the breakdown: {"subtasks":[{"name":"Step 1","description":"First","complexity":"simple"}]}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty",
			text:    ``,
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "fallback plain text",
			text:    "Write unit tests for parser\nUpdate docs for CLI usage",
			wantLen: 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaskBreakdown(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTaskBreakdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && len(got.Subtasks) != tt.wantLen {
				t.Errorf("parseTaskBreakdown() len(Subtasks) = %v, want %v", len(got.Subtasks), tt.wantLen)
			}
		})
	}
}
