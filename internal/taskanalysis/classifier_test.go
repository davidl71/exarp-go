package taskanalysis

import (
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestAnalyzeTask(t *testing.T) {
	tests := []struct {
		name           string
		task           *models.Todo2Task
		wantComplexity TaskComplexity
		wantAutoExec   bool
		wantBreakdown  bool
	}{
		{
			name:           "nil task defaults to medium",
			task:           nil,
			wantComplexity: ComplexityMedium,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "simple: add comment, no deps, short",
			task: &models.Todo2Task{
				ID:              "T-1",
				Content:         "Add comment to function X",
				LongDescription: "Brief docstring.",
				Dependencies:    nil,
			},
			wantComplexity: ComplexitySimple,
			wantAutoExec:   true,
			wantBreakdown:  false,
		},
		{
			name: "simple: fix typo",
			task: &models.Todo2Task{
				ID:           "T-2",
				Content:      "Fix typo in README",
				Dependencies: nil,
			},
			wantComplexity: ComplexitySimple,
			wantAutoExec:   true,
			wantBreakdown:  false,
		},
		{
			name: "complex: security keyword",
			task: &models.Todo2Task{
				ID:           "T-3",
				Content:      "Implement security validation for paths",
				Dependencies: nil,
			},
			wantComplexity: ComplexityComplex,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "complex: migration keyword",
			task: &models.Todo2Task{
				ID:           "T-4",
				Content:      "Migrate Python tools to Go",
				Dependencies: nil,
			},
			wantComplexity: ComplexityComplex,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "complex: many dependencies",
			task: &models.Todo2Task{
				ID:           "T-5",
				Content:      "Integrate all components",
				Dependencies: []string{"T-1", "T-2", "T-3", "T-4", "T-5"},
			},
			wantComplexity: ComplexityComplex,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "medium: moderate dependencies",
			task: &models.Todo2Task{
				ID:           "T-6",
				Content:      "Wire up handlers",
				Dependencies: []string{"T-1", "T-2", "T-3"},
			},
			wantComplexity: ComplexityMedium,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "medium: empty description default",
			task: &models.Todo2Task{
				ID:              "T-7",
				Content:         "",
				LongDescription: "",
				Dependencies:    nil,
			},
			wantComplexity: ComplexityMedium,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
		{
			name: "complex: refactor keyword",
			task: &models.Todo2Task{
				ID:           "T-8",
				Content:      "Refactor task analyzer",
				Dependencies: nil,
			},
			wantComplexity: ComplexityComplex,
			wantAutoExec:   false,
			wantBreakdown:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := AnalyzeTask(tt.task)
			if r.Complexity != tt.wantComplexity {
				t.Errorf("Complexity = %v, want %v", r.Complexity, tt.wantComplexity)
			}

			if r.CanAutoExecute != tt.wantAutoExec {
				t.Errorf("CanAutoExecute = %v, want %v", r.CanAutoExecute, tt.wantAutoExec)
			}

			if r.NeedsBreakdown != tt.wantBreakdown {
				t.Errorf("NeedsBreakdown = %v, want %v", r.NeedsBreakdown, tt.wantBreakdown)
			}

			if r.Reason == "" && tt.task != nil {
				t.Error("Reason should not be empty for non-nil task")
			}
		})
	}
}
