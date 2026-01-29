package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
)

func TestHandleReportOverview(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "overview with text format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "text",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result should be text format
			},
		},
		{
			name: "overview with json format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "overview with markdown format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "markdown",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result should be markdown format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleReportOverview(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReportOverview() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleReportPRD(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
			},
		},
		{
			name: "prd with project_name",
			params: map[string]interface{}{
				"action":       "prd",
				"project_name": "test-project",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result should contain project name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleReportPRD(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReportPRD() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleReport(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "overview action",
			params: map[string]interface{}{
				"action": "overview",
			},
			wantError: false,
		},
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
			},
			wantError: false,
		},
		{
			name: "scorecard action",
			params: map[string]interface{}{
				"action": "scorecard",
			},
			// scorecard: no error only when IsGoProject(); otherwise expect "only supported for Go projects"
			wantError: !IsGoProject(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleReport(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReport() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

// TestAggregateProjectDataProto verifies proto-based overview aggregation (step 1).
func TestAggregateProjectDataProto(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	ctx := context.Background()
	pb, err := aggregateProjectDataProto(ctx, tmpDir, false)
	if err != nil {
		t.Fatalf("aggregateProjectDataProto() error = %v", err)
	}
	if pb == nil {
		t.Fatal("aggregateProjectDataProto() returned nil")
	}
	if pb.GeneratedAt == "" {
		t.Error("expected GeneratedAt to be set")
	}
	if pb.Project == nil && pb.Tasks == nil && pb.Codebase == nil {
		t.Error("expected at least one of project, tasks, or codebase to be set")
	}
}

// TestFormatOverviewTextProto verifies proto-based formatters (no map assertions).
func TestFormatOverviewTextProto(t *testing.T) {
	if got := formatOverviewTextProto(nil); got != "" {
		t.Errorf("formatOverviewTextProto(nil) = %q, want \"\"", got)
	}
	pb := &proto.ProjectOverviewData{
		GeneratedAt: "2026-01-29T00:00:00Z",
		Project:     &proto.ProjectInfo{Name: "test-module", Version: "0.1.0", Type: "MCP", Status: "Active"},
		Tasks:       &proto.TaskMetrics{Total: 10, Pending: 3, Completed: 7, CompletionRate: 70},
	}
	got := formatOverviewTextProto(pb)
	if got == "" {
		t.Error("formatOverviewTextProto(proto) returned empty")
	}
	if !strings.Contains(got, "test-module") || !strings.Contains(got, "PROJECT OVERVIEW") {
		t.Errorf("formatOverviewTextProto output missing expected content: %s", got)
	}
}

func TestFormatOverviewMarkdownProto(t *testing.T) {
	if got := formatOverviewMarkdownProto(nil); got != "" {
		t.Errorf("formatOverviewMarkdownProto(nil) = %q, want \"\"", got)
	}
	pb := &proto.ProjectOverviewData{Project: &proto.ProjectInfo{Name: "p"}}
	got := formatOverviewMarkdownProto(pb)
	if !strings.Contains(got, "# Project Overview") || !strings.Contains(got, "p") {
		t.Errorf("formatOverviewMarkdownProto output missing expected content: %s", got)
	}
}

func TestFormatOverviewHTMLProto(t *testing.T) {
	if got := formatOverviewHTMLProto(nil); got != "" {
		t.Errorf("formatOverviewHTMLProto(nil) = %q, want \"\"", got)
	}
	pb := &proto.ProjectOverviewData{Project: &proto.ProjectInfo{Name: "p"}}
	got := formatOverviewHTMLProto(pb)
	if !strings.Contains(got, "<h1>Project Overview</h1>") || !strings.Contains(got, "p") {
		t.Errorf("formatOverviewHTMLProto output missing expected content: %s", got)
	}
}

// TestGoScorecardResultToProtoAndMap verifies scorecard proto path (step 2).
func TestGoScorecardResultToProtoAndMap(t *testing.T) {
	scorecard := &GoScorecardResult{
		Score:          65.0,
		Recommendations: []string{"add tests"},
		Metrics:        GoProjectMetrics{GoFiles: 10, MCPTools: 24},
		Health:         GoHealthChecks{GoTestCoverage: 72.5},
	}
	pb := GoScorecardResultToProto(scorecard)
	if pb == nil {
		t.Fatal("GoScorecardResultToProto returned nil")
	}
	if pb.Score != 65.0 || len(pb.Recommendations) != 1 || pb.TestCoverage != 72.5 {
		t.Errorf("proto mismatch: score=%.1f recommendations=%d test_coverage=%.1f",
			pb.Score, len(pb.Recommendations), pb.TestCoverage)
	}
	m := ProtoToScorecardMap(pb)
	if m["overall_score"] != 65.0 {
		t.Errorf("ProtoToScorecardMap overall_score = %v, want 65.0", m["overall_score"])
	}
	metrics, _ := m["metrics"].(map[string]interface{})
	if metrics == nil || metrics["test_coverage"] != 72.5 {
		t.Errorf("ProtoToScorecardMap metrics.test_coverage = %v, want 72.5", metrics)
	}
}
