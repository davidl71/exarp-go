package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/proto"
)

func TestHandleReportOverview(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

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
	t.Setenv("PROJECT_ROOT", tmpDir)

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

func TestPlanFilenameFromTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"github.com/davidl71/exarp-go", "exarp-go"},
		{"exarp-go", "exarp-go"},
		{"", "plan"},
		{"My Feature", "My-Feature"},
		{"path/to/project", "project"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			if got := planFilenameFromTitle(tt.title); got != tt.want {
				t.Errorf("planFilenameFromTitle(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestHandleReportPlan(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	planPath := tmpDir + "/test-project.plan.md"
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "plan action writes plan file with .plan.md suffix",
			params: map[string]interface{}{
				"action":      "plan",
				"output_path": planPath,
				"plan_title":  "Test Project",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}

				text := result[0].Text
				for _, s := range []string{"## Scope", "## 1. Technical Foundation", "## 2. Backlog Tasks", "## 3. Iterative Milestones", "## 4. Recommended Execution Order", "## 5. Open Questions", "## 6. Out-of-Scope"} {
					if !strings.Contains(text, s) {
						t.Errorf("plan output missing section %q", s)
					}
				}

				if !strings.Contains(text, "name:") || !strings.Contains(text, "overview:") {
					t.Error("plan output missing YAML frontmatter")
				}

				if !strings.Contains(text, "status: draft") {
					t.Error("plan output missing status: draft (Cursor Build/Built)")
				}

				if !strings.Contains(text, "Test Project") {
					t.Error("plan output missing plan title")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleReportPlan(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReportPlan() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestGetCodebaseMetrics_UsesCurrentRegistryCounts(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	metrics, err := getCodebaseMetrics(tmpDir)
	if err != nil {
		t.Fatalf("getCodebaseMetrics() error = %v", err)
	}

	if got := metrics["tools"]; got != ExpectedToolCountBase {
		t.Fatalf("tools = %v, want %d", got, ExpectedToolCountBase)
	}

	if got := metrics["prompts"]; got != len(prompts.ListAllPromptNames()) {
		t.Fatalf("prompts = %v, want %d", got, len(prompts.ListAllPromptNames()))
	}

	if got := metrics["resources"]; got != 27 {
		t.Fatalf("resources = %v, want 27", got)
	}
}

func TestHandleReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

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
			name: "plan action",
			params: map[string]interface{}{
				"action":      "plan",
				"output_path": tmpDir + "/exarp-go.plan.md",
			},
			wantError: false,
		},
		{
			name: "scorecard action",
			params: map[string]interface{}{
				"action": "scorecard",
			},
			// scorecard: always succeeds; Go project gets Go scorecard, non-Go gets generic scorecard
			wantError: false,
		},
		{
			name: "parallel_execution_plan action (empty project yields no waves)",
			params: map[string]interface{}{
				"action":      "parallel_execution_plan",
				"output_path": tmpDir + "/parallel-execution-subagents.plan.md",
			},
			// Empty tmpDir has no Todo2 backlog → "no waves" error
			wantError: true,
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

func TestHandleReportExecutionBriefingIncludesAgentRoleOrchestration(t *testing.T) {
	cleanup := initSessionTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tasks := []*models.Todo2Task{
		{
			ID:       "T-4200001",
			Content:  "Planning slice",
			Status:   models.StatusTodo,
			Priority: "high",
			Metadata: map[string]interface{}{metadataAgentRoleKey: AgentRolePlanner},
		},
		{
			ID:       "T-4200002",
			Content:  "Implementation slice",
			Status:   models.StatusTodo,
			Priority: "high",
			Metadata: map[string]interface{}{metadataAgentRoleKey: AgentRoleWorker},
		},
		{
			ID:       "T-4200003",
			Content:  "Review slice",
			Status:   models.StatusTodo,
			Priority: "medium",
			Metadata: map[string]interface{}{metadataAgentRoleKey: AgentRoleReviewer},
		},
	}
	for _, task := range tasks {
		if err := database.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask(%s): %v", task.ID, err)
		}
	}

	if _, err := database.ClaimTaskForAgent(ctx, "T-4200002", "worker-agent", 30*time.Minute); err != nil {
		t.Fatalf("ClaimTaskForAgent: %v", err)
	}
	run := &database.TaskExecutionRun{
		TaskID:  "T-4200002",
		AgentID: "worker-agent",
		Host:    "test-host",
		Status:  "running",
		Summary: "Implementing worker slice",
	}
	if err := database.StartTaskExecutionRun(ctx, run); err != nil {
		t.Fatalf("StartTaskExecutionRun: %v", err)
	}

	result, err := handleReportExecutionBriefing(ctx, map[string]interface{}{
		"action":  "execution_briefing",
		"limit":   10,
		"compact": true,
	})
	if err != nil {
		t.Fatalf("handleReportExecutionBriefing: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	summary, ok := payload["agent_role_summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agent_role_summary, got %T", payload["agent_role_summary"])
	}
	if got := summary["dominant_role"]; got != AgentRolePlanner && got != AgentRoleWorker && got != AgentRoleReviewer {
		t.Fatalf("unexpected dominant_role: %v", got)
	}
	distribution, ok := summary["distribution"].(map[string]interface{})
	if !ok || len(distribution) == 0 {
		t.Fatalf("expected distribution in agent_role_summary, got %v", summary["distribution"])
	}
	if distribution[AgentRolePlanner] == nil {
		t.Fatalf("expected planner in distribution, got %v", distribution)
	}

	lanes, ok := payload["orchestration_lanes"].([]interface{})
	if !ok || len(lanes) == 0 {
		t.Fatalf("expected orchestration_lanes, got %v", payload["orchestration_lanes"])
	}

	foundWorkerLane := false
	for _, laneRaw := range lanes {
		lane, ok := laneRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if lane["role"] != AgentRoleWorker {
			continue
		}
		foundWorkerLane = true
		if lane["active_claim_count"] != float64(1) {
			t.Fatalf("worker active_claim_count = %v, want 1", lane["active_claim_count"])
		}
		if lane["active_run_count"] != float64(1) {
			t.Fatalf("worker active_run_count = %v, want 1", lane["active_run_count"])
		}
	}
	if !foundWorkerLane {
		t.Fatalf("worker lane not found in %v", lanes)
	}

	suggestions, ok := payload["delegation_suggestions"].([]interface{})
	if !ok || len(suggestions) == 0 {
		t.Fatalf("expected delegation_suggestions, got %v", payload["delegation_suggestions"])
	}
}

// TestAggregateProjectDataProto verifies proto-based overview aggregation (step 1).
func TestAggregateProjectDataProto(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

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

func TestAggregateProjectDataProto_UsesExplicitProjectRootForGoDetection(t *testing.T) {
	outerRoot := t.TempDir()
	t.Setenv("PROJECT_ROOT", outerRoot)

	goRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(goRoot, "go.mod"), []byte("module example.com/test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}

	pb, err := aggregateProjectDataProto(context.Background(), goRoot, false)
	if err != nil {
		t.Fatalf("aggregateProjectDataProto() error = %v", err)
	}
	if pb.Health == nil && (pb.Project == nil || !strings.Contains(pb.Project.Description, "Health warning:")) {
		t.Fatalf("expected Go-root health attempt based on explicit project root, got %#v", pb)
	}
}

func TestHandleReportScorecardJSON_WritesOutputPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}
	outputPath := filepath.Join(tmpDir, "out", "scorecard.json")

	argsJSON, _ := json.Marshal(map[string]interface{}{
		"action":        "scorecard",
		"output_format": "json",
		"output_path":   outputPath,
		"fast_mode":     true,
	})

	result, err := handleReport(context.Background(), argsJSON)
	if err != nil {
		t.Fatalf("handleReport() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected scorecard output file at %s: %v", outputPath, err)
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

// TestBriefingDataProto verifies briefing uses proto internally (BriefingDataToMap, BuildBriefingDataProto).
func TestBriefingDataProto(t *testing.T) {
	// BriefingDataToMap(nil) returns empty map
	m := BriefingDataToMap(nil)
	if m == nil || len(m) != 0 {
		t.Errorf("BriefingDataToMap(nil) = %v, want non-nil empty map", m)
	}
	// Minimal proto produces expected keys
	pb := &proto.BriefingData{
		Date:    "2026-01-29",
		Score:   50,
		Sources: []string{"stoic"},
		Quotes:  []*proto.BriefingQuote{{Quote: "Test quote", Source: "stoic"}},
	}

	m = BriefingDataToMap(pb)
	if m["date"] != "2026-01-29" || m["score"] != 50.0 {
		t.Errorf("BriefingDataToMap: date=%v score=%v", m["date"], m["score"])
	}

	quotes, _ := m["quotes"].([]interface{})
	if len(quotes) != 1 {
		t.Errorf("BriefingDataToMap: quotes len = %d, want 1", len(quotes))
	}
}

// TestGoScorecardResultToProtoAndMap verifies scorecard proto path (step 2).
func TestGoScorecardResultToProtoAndMap(t *testing.T) {
	scorecard := &GoScorecardResult{
		Score:           65.0,
		Recommendations: []string{"add tests"},
		Metrics:         GoProjectMetrics{GoFiles: 10, MCPTools: 24},
		Health:          GoHealthChecks{GoTestCoverage: 72.5},
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
