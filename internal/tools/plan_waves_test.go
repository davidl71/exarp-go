package tools

import (
	"path/filepath"
	"testing"
)

func TestParseWavesFromPlanMarkdown(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "PARALLEL_EXECUTION_PLAN_RESEARCH.md")
	waves, err := ParseWavesFromPlanMarkdown(path)
	if err != nil {
		t.Fatalf("ParseWavesFromPlanMarkdown error = %v", err)
	}
	if waves == nil || len(waves) == 0 {
		t.Skip("docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md not found or empty")
	}
	if ids, ok := waves[0]; !ok || len(ids) == 0 {
		t.Errorf("expected wave 0 to have task IDs, got waves=%v", waves)
	}
}
