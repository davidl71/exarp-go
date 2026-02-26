// report_plan.go — Report plan: plan handler, wave formatting, and scorecard plans.
// See also: report_plan_generate.go
package tools

import (
	"context"
	"fmt"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleReportPlan
//   generateParallelExecutionSubagentsMarkdown — generateParallelExecutionSubagentsMarkdown builds the parallel-execution-subagents.plan.md content
//   FormatWavesAsSubagentsPlanMarkdown — FormatWavesAsSubagentsPlanMarkdown formats waves as parallel-execution-subagents.plan.md content.
//   handleReportParallelExecutionPlan — handleReportParallelExecutionPlan generates .cursor/plans/parallel-execution-subagents.plan.md from current Todo2 waves.
//   handleReportUpdateWavesFromPlan — handleReportUpdateWavesFromPlan parses docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md and updates Todo2
//   scorecardDimensionConfig — scorecardDimensionConfig holds display name and threshold for a scorecard dimension.
//   recommendationsByDimension — recommendationsByDimension classifies scorecard recommendations by dimension (testing, security, documentation, completion).
//   handleReportScorecardPlans — handleReportScorecardPlans runs the scorecard and writes one Cursor-style plan per dimension that is below threshold.
//   generateScorecardDimensionPlan — generateScorecardDimensionPlan returns markdown for a single dimension improvement plan (Cursor-buildable).
// ────────────────────────────────────────────────────────────────────────────

// ─── handleReportPlan ───────────────────────────────────────────────────────
func handleReportPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	planTitle, _ := params["plan_title"].(string)
	if planTitle == "" {
		if info, err := getProjectInfo(projectRoot); err == nil {
			if name, ok := info["name"].(string); ok && name != "" {
				planTitle = name
			}
		}

		if planTitle == "" {
			planTitle = filepath.Base(projectRoot)
		}
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath == "" {
		slug := planFilenameFromTitle(planTitle) + ".plan.md"
		outputPath = filepath.Join(projectRoot, ".cursor", "plans", slug)
	} else {
		outputPath = ensurePlanMdSuffix(outputPath)
	}

	repair, _ := params["repair"].(bool)
	planPath, _ := params["plan_path"].(string)
	if planPath == "" {
		planPath = outputPath
	} else {
		planPath = ensurePlanMdSuffix(planPath)
	}
	if repair {
		repaired, err := repairPlanFile(ctx, projectRoot, planPath, planTitle)
		if err != nil {
			return nil, fmt.Errorf("repair plan: %w", err)
		}
		return []framework.TextContent{{Type: "text", Text: repaired}}, nil
	}

	planMD, err := generatePlanMarkdown(ctx, projectRoot, planTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create plan directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, []byte(planMD), 0644); err != nil {
		return nil, fmt.Errorf("failed to write plan file: %w", err)
	}

	// Optionally update parallel-execution-subagents.plan.md when include_subagents is true
	includeSubagents, _ := params["include_subagents"].(bool)
	if includeSubagents {
		subagentsMD, subErr := generateParallelExecutionSubagentsMarkdown(ctx, projectRoot, planTitle)
		if subErr != nil {
			planMD += fmt.Sprintf("\n\n[Warning: failed to update subagents plan: %v]", subErr)
		} else {
			subagentsPath := filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
			if dir := filepath.Dir(subagentsPath); dir != "." {
				_ = os.MkdirAll(dir, 0755)
			}

			if err := os.WriteFile(subagentsPath, []byte(subagentsMD), 0644); err == nil {
				planMD += fmt.Sprintf("\n\n[Subagents plan updated: %s]", subagentsPath)
			} else {
				planMD += fmt.Sprintf("\n\n[Warning: failed to write subagents plan: %v]", err)
			}
		}
	}

	planMD += fmt.Sprintf("\n\n[Plan saved to: %s]", outputPath)

	return []framework.TextContent{
		{Type: "text", Text: planMD},
	}, nil
}

// ─── generateParallelExecutionSubagentsMarkdown ─────────────────────────────
// generateParallelExecutionSubagentsMarkdown builds the parallel-execution-subagents.plan.md content
// from current Todo2 backlog and dependency waves. Uses BacklogExecutionOrder for wave assignment.
// Only active tasks (Todo and In Progress) are included; Done tasks are excluded.
func generateParallelExecutionSubagentsMarkdown(ctx context.Context, projectRoot, planTitle string) (string, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("list tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	_, waves, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		return "", fmt.Errorf("backlog execution order: %w", err)
	}

	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}

	if len(waves) == 0 {
		return "", fmt.Errorf("no waves (empty backlog or no Todo/In Progress tasks)")
	}

	return FormatWavesAsSubagentsPlanMarkdown(waves, planTitle), nil
}

// ─── FormatWavesAsSubagentsPlanMarkdown ─────────────────────────────────────
// FormatWavesAsSubagentsPlanMarkdown formats waves as parallel-execution-subagents.plan.md content.
// Used by report parallel_execution_plan and task_analysis execution_plan (output_format=subagents_plan).
func FormatWavesAsSubagentsPlanMarkdown(waves map[int][]string, planTitle string) string {
	mainPlanSlug := planFilenameFromTitle(planTitle)
	mainPlanFile := mainPlanSlug + ".plan.md"

	levelOrder := make([]int, 0, len(waves))
	for k := range waves {
		levelOrder = append(levelOrder, k)
	}

	sort.Ints(levelOrder)

	var sb strings.Builder

	waveCount := len(levelOrder)
	waveLabel := fmt.Sprintf("Waves 0–%d", waveCount-1)

	if waveCount == 1 {
		waveLabel = "Wave 0"
	}

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: Parallel Execution (Subagents) — %s\n", waveLabel))
	sb.WriteString("overview: Run project plan waves using Cursor subagents; one subagent per task within each wave, waves run sequentially.\n")
	sb.WriteString("isProject: false\n")
	sb.WriteString("status: draft\n")
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# Parallel Execution Plan: Subagents (%s)\n\n", waveLabel))
	sb.WriteString(fmt.Sprintf("**Source plan:** [.cursor/plans/%s](.cursor/plans/%s)\n\n", mainPlanFile, mainPlanFile))
	sb.WriteString("**Strategy:** Execute waves sequentially (Wave 0, then Wave 1, …). Within each wave, launch **one subagent per task** in a single message so they run in parallel. Use the `wave-task-runner` subagent; optionally run `wave-verifier` after each wave.\n\n")
	sb.WriteString("---\n\n## Execution order\n\n")

	for i, level := range levelOrder {
		sb.WriteString(fmt.Sprintf("%d. **Wave %d** — Run all Wave %d tasks in parallel. Optionally run `wave-verifier` when complete.\n", i+1, level, level))
	}

	sb.WriteString("\n---\n\n")

	// One section per wave
	for _, level := range levelOrder {
		ids := waves[level]
		if len(ids) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## Wave %d (parallel)\n\n", level))
		sb.WriteString("Launch one `/wave-task-runner` per task in **one** message so Cursor runs them in parallel.\n\n")

		countLabel := "tasks"
		if len(ids) == 1 {
			countLabel = "task"
		}

		sb.WriteString(fmt.Sprintf("**Task IDs (%d %s):**\n\n", len(ids), countLabel))
		sb.WriteString(strings.Join(ids, ", "))
		sb.WriteString("\n\n**Prompt (Wave ")
		sb.WriteString(fmt.Sprint(level))
		sb.WriteString("):**\n\n> Execute Wave ")
		sb.WriteString(fmt.Sprint(level))
		sb.WriteString(" in parallel: run **wave-task-runner** once per task for the Wave ")
		sb.WriteString(fmt.Sprint(level))
		sb.WriteString(" list above. Give each subagent its task_id, the task description from the plan (or Todo2), and plan context from `.cursor/plans/")
		sb.WriteString(mainPlanFile)
		sb.WriteString("`. Wait for all to complete, then summarize. Optionally run **wave-verifier** for Wave ")
		sb.WriteString(fmt.Sprint(level))
		sb.WriteString(".\n\n")

		if len(ids) > 20 {
			sb.WriteString("**Practical note:** Consider batching (e.g. first 10–15 tasks) to avoid UI or token limits.\n\n")
		}

		sb.WriteString("---\n\n")
	}

	sb.WriteString("## Subagents used\n\n")
	sb.WriteString("| Subagent | Role |\n")
	sb.WriteString("|----------|------|\n")
	sb.WriteString("| **wave-task-runner** | One instance per task; implements/completes that task and returns a short summary. |\n")
	sb.WriteString("| **wave-verifier** | Optional; run once per wave after all wave-task-runner instances return; validates outcomes. |\n\n")
	sb.WriteString("Defined in [.cursor/agents/wave-task-runner.md](.cursor/agents/wave-task-runner.md) and [.cursor/agents/wave-verifier.md](.cursor/agents/wave-verifier.md).\n\n")
	sb.WriteString("---\n\n## Quick reference\n\n")
	sb.WriteString("| Wave | Task count | Run |\n")
	sb.WriteString("|------|------------|-----|\n")

	for _, level := range levelOrder {
		ids := waves[level]
		if len(ids) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("| Wave %d | %d | One subagent per task in one message (or batches of ~10–15 if large). |\n", level, len(ids)))
	}

	sb.WriteString("\n**Order:** Execute waves sequentially (do not start Wave N+1 until Wave N is complete).\n\n")
	sb.WriteString("*Generated by exarp-go task_analysis(action=execution_plan, output_format=subagents_plan) or report(action=parallel_execution_plan).*\n")

	return sb.String()
}

// ─── handleReportParallelExecutionPlan ──────────────────────────────────────
// handleReportParallelExecutionPlan generates .cursor/plans/parallel-execution-subagents.plan.md from current Todo2 waves.
func handleReportParallelExecutionPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	planTitle, _ := params["plan_title"].(string)
	if planTitle == "" {
		if info, err := getProjectInfo(projectRoot); err == nil {
			if name, ok := info["name"].(string); ok && name != "" {
				planTitle = name
			}
		}

		if planTitle == "" {
			planTitle = filepath.Base(projectRoot)
		}
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath == "" {
		outputPath = filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
	}

	subagentsMD, err := generateParallelExecutionSubagentsMarkdown(ctx, projectRoot, planTitle)
	if err != nil {
		return nil, fmt.Errorf("generate parallel execution plan: %w", err)
	}

	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create plan directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, []byte(subagentsMD), 0644); err != nil {
		return nil, fmt.Errorf("write subagents plan: %w", err)
	}

	msg := fmt.Sprintf("Parallel execution subagents plan saved to: %s", outputPath)

	return []framework.TextContent{
		{Type: "text", Text: msg},
	}, nil
}

// ─── handleReportUpdateWavesFromPlan ────────────────────────────────────────
// handleReportUpdateWavesFromPlan parses docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md and updates Todo2
// task dependencies so BacklogExecutionOrder produces those waves (wave 0 = no deps, wave N = depend on one task from wave N-1).
func handleReportUpdateWavesFromPlan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	planPath, _ := params["plan_path"].(string)
	if planPath == "" {
		planPath = filepath.Join(projectRoot, DefaultPlanWavesPath)
	}

	waves, err := ParseWavesFromPlanMarkdown(planPath)
	if err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}

	if len(waves) == 0 {
		return nil, fmt.Errorf("no waves in plan %s (empty or file not found)", planPath)
	}

	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}

	sort.Ints(levels)

	updated := 0

	for _, level := range levels {
		ids := waves[level]

		var deps []string

		if level > 0 {
			prevIDs := waves[levels[level-1]]
			if len(prevIDs) == 0 {
				return nil, fmt.Errorf("wave %d has no predecessor wave", level)
			}

			deps = []string{prevIDs[0]}
		}

		for _, taskID := range ids {
			task, err := database.GetTask(ctx, taskID)
			if err != nil || task == nil {
				continue // skip missing tasks
			}
			// Update only if deps differ
			same := len(task.Dependencies) == len(deps)
			if same && len(deps) > 0 {
				same = task.Dependencies[0] == deps[0]
			}

			if !same {
				updated++
				task.Dependencies = make([]string, len(deps))
				copy(task.Dependencies, deps)

				if err := database.UpdateTask(ctx, task); err != nil {
					return nil, fmt.Errorf("update task %s: %w", taskID, err)
				}
			}
		}
	}

	msg := fmt.Sprintf("Updated %d task dependencies from %s (%d waves)", updated, planPath, len(levels))

	return []framework.TextContent{
		{Type: "text", Text: msg},
	}, nil
}

// ─── scorecardDimensionConfig ───────────────────────────────────────────────
// scorecardDimensionConfig holds display name and threshold for a scorecard dimension.
var scorecardDimensionConfig = map[string]struct {
	displayName string
	threshold   float64
}{
	"testing":       {"Testing", 100},
	"security":      {"Security", 100},
	"documentation": {"Documentation", 100},
	"completion":    {"Completion & quality", 100},
}

// ─── recommendationsByDimension ─────────────────────────────────────────────
// recommendationsByDimension classifies scorecard recommendations by dimension (testing, security, documentation, completion).
func recommendationsByDimension(recs []string) map[string][]string {
	out := map[string][]string{
		"testing":       nil,
		"security":      nil,
		"documentation": nil,
		"completion":    nil,
	}

	lower := func(s string) string { return strings.ToLower(s) }
	for _, r := range recs {
		l := lower(r)

		switch {
		case strings.Contains(l, "test") || strings.Contains(l, "coverage"):
			out["testing"] = append(out["testing"], r)
		case strings.Contains(l, "govulncheck") || strings.Contains(l, "vuln") || strings.Contains(l, "path boundary") || strings.Contains(l, "rate limit") || strings.Contains(l, "access control"):
			out["security"] = append(out["security"], r)
		case strings.Contains(l, "doc") || strings.Contains(l, "readme"):
			out["documentation"] = append(out["documentation"], r)
		default:
			// go.mod, go.sum, tidy, build, vet, fmt, lint
			out["completion"] = append(out["completion"], r)
		}
	}

	return out
}

// ─── handleReportScorecardPlans ─────────────────────────────────────────────
// handleReportScorecardPlans runs the scorecard and writes one Cursor-style plan per dimension that is below threshold.
// Plans are written to .cursor/plans/improve-<dimension>.plan.md. Use from Makefile (make scorecard-plans) or git hooks.
func handleReportScorecardPlans(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	if !IsGoProject() {
		return nil, fmt.Errorf("scorecard_plans is only supported for Go projects (go.mod)")
	}

	fastMode := true
	if v, ok := params["fast_mode"].(bool); ok {
		fastMode = v
	}

	opts := &ScorecardOptions{FastMode: fastMode}

	scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
	if err != nil {
		return nil, fmt.Errorf("scorecard for plans: %w", err)
	}

	scores := map[string]float64{
		"testing":       calculateTestingScore(scorecard),
		"security":      calculateSecurityScore(scorecard),
		"documentation": calculateDocumentationScore(scorecard),
		"completion":    calculateCompletionScore(scorecard),
	}
	byDim := recommendationsByDimension(scorecard.Recommendations)

	// Add a generic documentation recommendation when doc score is below threshold and none exist
	if scores["documentation"] < 100 && len(byDim["documentation"]) == 0 {
		byDim["documentation"] = append(byDim["documentation"], "Improve documentation (README, godoc, .cursor/docs)")
	}

	plansDir := filepath.Join(projectRoot, ".cursor", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		return nil, fmt.Errorf("create plans dir: %w", err)
	}

	var created []string

	for _, dim := range []string{"testing", "security", "documentation", "completion"} {
		cfg := scorecardDimensionConfig[dim]
		if scores[dim] >= cfg.threshold {
			continue
		}

		recs := byDim[dim]
		if len(recs) == 0 && dim != "documentation" {
			recs = []string{fmt.Sprintf("Review and improve %s (score: %.0f%%)", cfg.displayName, scores[dim])}
		}

		planPath := filepath.Join(plansDir, "improve-"+dim+".plan.md")
		planMD := generateScorecardDimensionPlan(cfg.displayName, dim, scores[dim], cfg.threshold, recs)

		if err := os.WriteFile(planPath, []byte(planMD), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", planPath, err)
		}

		created = append(created, planPath)
	}

	var sb strings.Builder

	sb.WriteString("Scorecard improvement plans created:\n")

	if len(created) == 0 {
		sb.WriteString("  (none — all dimensions at or above threshold)\n")
	} else {
		for _, p := range created {
			sb.WriteString("  - " + p + "\n")
		}
	}

	return []framework.TextContent{{Type: "text", Text: sb.String()}}, nil
}

// ─── generateScorecardDimensionPlan ─────────────────────────────────────────
// generateScorecardDimensionPlan returns markdown for a single dimension improvement plan (Cursor-buildable).
func generateScorecardDimensionPlan(displayName, dimension string, currentScore, targetScore float64, recommendations []string) string {
	var sb strings.Builder

	scorecardDate := time.Now().Format("2006-01-02")
	tagHints := dimension + ", planning"

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: Improve %s\n", displayName))
	sb.WriteString(fmt.Sprintf("overview: \"Current %s score: %.0f%%. Target: %.0f%%. Address the items below to improve this metric.\"\n", displayName, currentScore, targetScore))
	sb.WriteString("todos:\n")

	for i, r := range recommendations {
		id := fmt.Sprintf("rec-%s-%d", dimension, i+1)

		content := r
		if len(content) > 120 {
			content = content[:117] + "..."
		}

		sb.WriteString(fmt.Sprintf("  - id: %s\n", id))
		sb.WriteString(fmt.Sprintf("    content: %q\n", escapeYAMLString(content)))
		sb.WriteString("    status: pending\n")
	}

	sb.WriteString("isProject: false\n")
	sb.WriteString("status: draft\n")
	sb.WriteString(fmt.Sprintf("last_updated: %q\n", scorecardDate))
	sb.WriteString(fmt.Sprintf("tag_hints: [%s]\n", tagHints))
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# Improve %s\n\n", displayName))
	sb.WriteString(fmt.Sprintf("**Current score:** %.0f%% | **Target:** %.0f%%\n\n", currentScore, targetScore))
	sb.WriteString("**Status:** draft\n\n")
	sb.WriteString(fmt.Sprintf("**Last updated:** %s\n\n", scorecardDate))
	sb.WriteString("**Referenced by:** (none by default; add agents in `.cursor/agents/` or rules in `.cursor/rules/` to reference this plan)\n\n")
	sb.WriteString(fmt.Sprintf("**Tag hints:** `#%s` `#planning`\n\n", dimension))
	sb.WriteString("## Agents\n\n")
	sb.WriteString("| Agent | Role |\n")
	sb.WriteString("|-------|------|\n")
	sb.WriteString("| *(none)* | Add [scorecard-improvement-runner](.cursor/agents/scorecard-improvement-runner.md) or a custom agent to execute this plan |\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Objective\n\n")
	sb.WriteString(fmt.Sprintf("Improve the %s dimension from %.0f%% to %.0f%% by addressing the checklist below.\n\n", displayName, currentScore, targetScore))
	sb.WriteString("## Success criteria\n\n")
	sb.WriteString("- Score reaches or exceeds target\n")
	sb.WriteString("- All checklist items addressed or explicitly deferred\n\n")
	sb.WriteString("## Checklist\n\n")

	for _, r := range recommendations {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", r))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("## Open Questions\n\n")
	sb.WriteString("- *(Add open questions or decisions needed.)*\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## References\n\n")
	sb.WriteString("- *(See main plan or scorecard output; add links as needed.)*\n\n")
	sb.WriteString("---\n\n*Generated from scorecard by exarp-go report(action=scorecard_plans).*\n")

	return sb.String()
}

// generatePlanMarkdown builds a Cursor-buildable plan aligned with reference format: frontmatter, Scope, Technical Foundation, Backlog table, Milestones, Execution Order, Open Questions, Out-of-Scope.
