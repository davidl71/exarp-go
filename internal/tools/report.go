package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// handleReportOverview handles the overview action for report tool.
func handleReportOverview(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	outputFormat := "text"
	if format := strings.TrimSpace(cast.ToString(params["output_format"])); format != "" {
		outputFormat = format
	}

	outputPath := cast.ToString(params["output_path"])
	includePlanning := cast.ToBool(params["include_planning"])

	// Aggregate project data (proto-based internally)
	overviewProto, err := aggregateProjectDataProto(ctx, projectRoot, includePlanning)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate project data: %w", err)
	}

	// Format output based on requested format (use proto for type-safe formatting)
	var formattedOutput string

	switch outputFormat {
	case "json":
		overviewMap := ProtoToProjectOverviewData(overviewProto)
		compact := cast.ToBool(params["compact"])
		contents, err := FormatResultOptionalCompact(overviewMap, outputPath, compact)
		if err != nil {
			return nil, fmt.Errorf("failed to format JSON: %w", err)
		}

		return contents, nil
	case "markdown":
		formattedOutput = formatOverviewMarkdownProto(overviewProto)
	case "html":
		formattedOutput = formatOverviewHTMLProto(overviewProto)
	default:
		formattedOutput = formatOverviewTextProto(overviewProto)
	}

	// Save to file if requested
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(formattedOutput), 0644); err != nil {
			return nil, fmt.Errorf("failed to write output file: %w", err)
		}

		formattedOutput += fmt.Sprintf("\n\n[Report saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: formattedOutput},
	}, nil
}

// handleReportBriefing handles the briefing action for report tool.
// Uses proto internally (BuildBriefingDataProto) for type-safe briefing data.
func handleReportBriefing(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var score float64

	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	} else {
		score = 50.0 // Default score
	}

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Build briefing from proto (type-safe)
	briefingProto := BuildBriefingDataProto(engine, score)
	briefingMap := BriefingDataToMap(briefingProto)
	compact, _ := params["compact"].(bool)
	return FormatResultOptionalCompact(briefingMap, "", compact)
}

// handleReportPRD handles the prd action for report tool.
func handleReportPRD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	projectName, _ := params["project_name"].(string)
	if projectName == "" {
		// Try to get from go.mod or default
		projectName = filepath.Base(projectRoot)
	}

	includeArchitecture := true
	if arch, ok := params["include_architecture"].(bool); ok {
		includeArchitecture = arch
	}

	includeMetrics := true
	if metrics, ok := params["include_metrics"].(bool); ok {
		includeMetrics = metrics
	}

	includeTasks := true
	if tasks, ok := params["include_tasks"].(bool); ok {
		includeTasks = tasks
	}

	outputPath, _ := params["output_path"].(string)

	// Generate PRD
	prd, err := generatePRD(ctx, projectRoot, projectName, includeArchitecture, includeMetrics, includeTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PRD: %w", err)
	}

	// Save to file if requested
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(prd), 0644); err != nil {
			return nil, fmt.Errorf("failed to write PRD file: %w", err)
		}

		prd += fmt.Sprintf("\n\n[PRD saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: prd},
	}, nil
}

// handleReportPlan generates a Cursor-style plan file with .plan.md suffix (Purpose, Technical Foundation, Iterative Milestones, Open Questions).
// See https://cursor.com/learn/creating-plans. Default output is .cursor/plans/<project-slug>.plan.md so Cursor discovers it and shows Build.
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
func generatePlanMarkdown(ctx context.Context, projectRoot, planTitle string) (string, error) {
	displayName := planDisplayName(planTitle)
	info, _ := getProjectInfo(projectRoot)

	overview := getStr(info, "description")
	if overview == "" {
		overview = fmt.Sprintf("Deliver and maintain %s with clear milestones and quality gates.", displayName)
	}

	store := NewDefaultTaskStore(projectRoot)
	list, _ := store.ListTasks(ctx, nil)
	tasks := tasksFromPtrs(list)
	taskByID := make(map[string]Todo2Task)

	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	snippet := getPlanningSnippet(projectRoot)
	actions, _ := getNextActions(projectRoot)
	risks, _ := getRisksAndBlockers(projectRoot)

	// Todos for frontmatter (Cursor-buildable: id, content, status so Cursor displays them as Todos)
	todoEntries := make([]struct{ id, content, status string }, 0, len(actions))

	for _, a := range actions {
		id := getStr(a, "task_id")
		if id == "" {
			continue
		}

		content := getStr(a, "name")
		status := "pending"

		if t, ok := taskByID[id]; ok {
			if content == "" {
				content = t.Content
			}

			status = todo2StatusToCursorStatus(t.Status)
		}

		if content == "" {
			content = id
		}

		content = strings.TrimSpace(content)
		if len(content) > 120 {
			content = content[:117] + "..."
		}

		todoEntries = append(todoEntries, struct{ id, content, status string }{id, content, status})
	}

	// Execution order and waves (same wave = can run in parallel); computed once for frontmatter and section 4
	orderedIDs, waves, _, _ := BacklogExecutionOrder(tasks, nil)
	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}

	var sb strings.Builder

	// YAML frontmatter (Cursor plan format; todos = [{ id, content, status }]; waves = parallelizable groups)
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s Plan\n", displayName))
	sb.WriteString(fmt.Sprintf("overview: %q\n", escapeYAMLString(overview)))

	if len(todoEntries) > 0 {
		sb.WriteString("todos:\n")

		for _, e := range todoEntries {
			sb.WriteString(fmt.Sprintf("  - id: %s\n", e.id))
			sb.WriteString(fmt.Sprintf("    content: %q\n", escapeYAMLString(e.content)))
			sb.WriteString(fmt.Sprintf("    status: %s\n", e.status))
		}
	} else {
		sb.WriteString("todos: []\n")
	}

	if len(waves) > 0 {
		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)
		sb.WriteString("waves:\n")

		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}

			quoted := make([]string, len(ids))
			for i, id := range ids {
				quoted[i] = fmt.Sprintf("%q", id)
			}

			sb.WriteString("  - [" + strings.Join(quoted, ", ") + "]\n")
		}
	}

	planDate := time.Now().Format("2006-01-02")

	sb.WriteString("isProject: true\n")
	sb.WriteString("status: draft\n")
	sb.WriteString(fmt.Sprintf("last_updated: %q\n", planDate))
	sb.WriteString("tag_hints: [planning]\n")
	// Referenced by / Agents in frontmatter so Cursor UI can show "Referenced by N" (same as body block below)
	sb.WriteString("referenced_by:\n")
	sb.WriteString("  - .cursor/agents/wave-task-runner.md\n")
	sb.WriteString("  - .cursor/agents/wave-verifier.md\n")
	sb.WriteString("  - .cursor/rules/plan-execution.mdc\n")
	sb.WriteString("agents:\n")
	sb.WriteString("  - name: wave-task-runner\n")
	sb.WriteString("    path: .cursor/agents/wave-task-runner.md\n")
	sb.WriteString("    role: Run one task per wave from this plan\n")
	sb.WriteString("  - name: wave-verifier\n")
	sb.WriteString("    path: .cursor/agents/wave-verifier.md\n")
	sb.WriteString("    role: Verify wave outcomes and update status\n")
	sb.WriteString("---\n\n")
	sb.WriteString("*Regenerate with: `exarp-go -tool report -args '{\"action\":\"plan\"}'`. Do not edit frontmatter in Cursor; it may strip fields and break Build.*\n\n")
	sb.WriteString(fmt.Sprintf("# %s Plan\n\n", displayName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", planDate))
	sb.WriteString("**Status:** draft\n\n")
	sb.WriteString(fmt.Sprintf("**Last updated:** %s\n\n", planDate))
	sb.WriteString("**Referenced by:** [wave-task-runner](.cursor/agents/wave-task-runner.md), [wave-verifier](.cursor/agents/wave-verifier.md), [plan-execution](.cursor/rules/plan-execution.mdc)\n\n")
	sb.WriteString("**Tag hints:** `#planning`\n\n")
	sb.WriteString("## Agents\n\n")
	sb.WriteString("| Agent | Role |\n")
	sb.WriteString("|-------|------|\n")
	sb.WriteString("| [wave-task-runner](.cursor/agents/wave-task-runner.md) | Run one task per wave from this plan |\n")
	sb.WriteString("| [wave-verifier](.cursor/agents/wave-verifier.md) | Verify wave outcomes and update status |\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Scope\n\n")
	sb.WriteString("**Purpose:** " + overview + "\n\n")
	sb.WriteString("**Success criteria:** Clear milestones and quality gates; backlog aligned with execution order.\n\n")
	sb.WriteString("---\n\n")

	// 1. Technical Foundation
	sb.WriteString("## 1. Technical Foundation\n\n")

	metrics, err := getCodebaseMetrics(projectRoot)
	if err == nil {
		sb.WriteString(fmt.Sprintf("- **Project type:** %s\n", getStr(info, "type")))
		sb.WriteString(fmt.Sprintf("- **Tools:** %v | **Prompts:** %v | **Resources:** %v\n", metrics["tools"], metrics["prompts"], metrics["resources"]))
		sb.WriteString(fmt.Sprintf("- **Codebase:** %v files (Go: %v)\n", metrics["total_files"], metrics["go_files"]))
	}

	sb.WriteString("- **Storage:** Todo2 (SQLite primary, JSON fallback)\n")
	sb.WriteString("- **Invariants:** Use Makefile targets; prefer report/task_workflow over direct file edits\n\n")

	if path, ok := snippet["critical_path"].([]string); ok && len(path) > 0 {
		sb.WriteString("**Critical path (longest dependency chain):** ")
		sb.WriteString(strings.Join(path, " → "))
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n\n")

	// 2. Backlog Tasks (table: Task | Priority | Description)
	sb.WriteString("## 2. Backlog Tasks\n\n")
	sb.WriteString("| Task | Priority | Description |\n")
	sb.WriteString("|------|----------|-------------|\n")

	if len(actions) > 0 {
		for _, a := range actions {
			taskID := getStr(a, "task_id")
			priority := getStr(a, "priority")
			name := getStr(a, "name")

			desc := name
			if t, ok := taskByID[taskID]; ok && t.LongDescription != "" {
				desc = tableCellSafe(t.LongDescription)
			} else if name != "" {
				desc = tableCellSafe(name)
			} else {
				desc = taskID
			}

			sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", taskID, priority, desc))
		}
	} else {
		if ids, ok := snippet["suggested_backlog_order"].([]string); ok && len(ids) > 0 {
			for _, id := range ids {
				t := taskByID[id]
				desc := t.Content

				if t.LongDescription != "" {
					desc = tableCellSafe(t.LongDescription)
				} else {
					desc = tableCellSafe(t.Content)
				}

				sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", id, t.Priority, desc))
			}
		} else {
			sb.WriteString("| — | — | *(Add Todo2 tasks; run report with include_planning to populate.)* |\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 3. Iterative Milestones (checkboxes — Cursor buildable)
	sb.WriteString("## 3. Iterative Milestones\n\n")
	sb.WriteString("Each milestone is independently valuable. Check off as done.\n\n")

	if len(actions) > 0 {
		for _, a := range actions {
			name := getStr(a, "name")
			taskID := getStr(a, "task_id")

			if name == "" {
				name = taskID
			}

			line := fmt.Sprintf("- [ ] **%s** (%s)", name, taskID)

			if t, ok := taskByID[taskID]; ok {
				if refs := getTaskFileRefs(&t); len(refs) > 0 {
					line += " — " + strings.Join(refs, ", ")
				}
			}

			sb.WriteString(line + "\n")
		}
	} else {
		if ids, ok := snippet["suggested_backlog_order"].([]string); ok && len(ids) > 0 {
			for _, id := range ids {
				line := fmt.Sprintf("- [ ] %s", id)

				if t, ok := taskByID[id]; ok {
					if refs := getTaskFileRefs(&t); len(refs) > 0 {
						line += " — " + strings.Join(refs, ", ")
					}
				}

				sb.WriteString(line + "\n")
			}
		} else {
			sb.WriteString("- [ ] *(Add Todo2 tasks to populate milestones)*\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 4. Recommended Execution Order (critical path + waves/parallelizable + optional mermaid)
	cp, _ := snippet["critical_path"].([]string)

	sb.WriteString("## 4. Recommended Execution Order\n\n")

	if len(cp) > 0 {
		sb.WriteString("```mermaid\nflowchart TD\n")

		for i := 0; i < len(cp); i++ {
			node := fmt.Sprintf("CP%d", i+1)
			label := cp[i]
			sb.WriteString(fmt.Sprintf("    %s[%s]\n", node, label))
		}

		for i := 1; i < len(cp); i++ {
			sb.WriteString(fmt.Sprintf("    CP%d --> CP%d\n", i, i+1))
		}

		sb.WriteString("```\n\n")

		for i, taskID := range cp {
			sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, taskID))
		}
		// Parallel: next actions not on critical path
		cpSet := make(map[string]bool)
		for _, id := range cp {
			cpSet[id] = true
		}

		var parallel []string

		for _, a := range actions {
			id := getStr(a, "task_id")
			if !cpSet[id] {
				parallel = append(parallel, id)
			}
		}

		if len(parallel) > 0 {
			sb.WriteString("\n**Parallel:** " + strings.Join(parallel, ", ") + "\n")
		}
	} else {
		sb.WriteString("1. Complete backlog in dependency order (see Milestones).\n")

		if len(actions) > 0 {
			for i, a := range actions {
				sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, getStr(a, "name"), getStr(a, "task_id")))
			}
		}
	}

	// Waves: same wave = can run in parallel (from BacklogExecutionOrder dependency levels)
	if len(waves) > 0 {
		sb.WriteString("\n### Waves (same wave = can run in parallel)\n\n")

		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}

		sort.Ints(levelOrder)

		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}

			if len(ids) >= 2 {
				sb.WriteString(fmt.Sprintf("- **Wave %d** (parallel): %s\n", level, strings.Join(ids, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf("- **Wave %d:** %s\n", level, ids[0]))
			}
		}

		sb.WriteString("\n")

		if len(orderedIDs) > 0 {
			sb.WriteString("**Full order:** " + strings.Join(orderedIDs, ", ") + "\n")
		}
	}

	sb.WriteString("\n---\n\n")

	// 5. Open Questions
	sb.WriteString("## 5. Open Questions\n\n")

	if len(risks) > 0 {
		for _, r := range risks {
			sb.WriteString(fmt.Sprintf("- %s", getStr(r, "description")))

			if id := getStr(r, "task_id"); id != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", id))
			}

			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("- *(Add open questions or decisions needed during implementation.)*\n")
	}

	sb.WriteString("\n---\n\n")

	// 6. Out-of-Scope / Deferred
	sb.WriteString("## 6. Out-of-Scope / Deferred\n\n")
	// Low-priority backlog tasks not in next actions, or placeholder
	var deferred []string

	actionIDs := make(map[string]bool)
	for _, a := range actions {
		actionIDs[getStr(a, "task_id")] = true
	}

	for _, t := range tasks {
		if IsPendingStatus(t.Status) && t.Priority == models.PriorityLow && !actionIDs[t.ID] {
			deferred = append(deferred, fmt.Sprintf("**%s** (%s) — low priority", t.ID, t.Content))
		}
	}

	if len(deferred) > 0 {
		for _, line := range deferred {
			sb.WriteString("- " + line + "\n")
		}
	} else {
		sb.WriteString("- *(Add deferred or out-of-scope items as needed.)*\n")
	}

	sb.WriteString("\n---\n\n")

	// 7. Key Files / Implementation Notes
	sb.WriteString("## 7. Key Files / Implementation Notes\n\n")
	sb.WriteString("- *(Add key files or implementation notes; or leave as placeholder.)*\n")
	sb.WriteString("\n---\n\n")

	// 8. References
	sb.WriteString("## 8. References\n\n")
	sb.WriteString("- [docs/BACKLOG_EXECUTION_PLAN.md](docs/BACKLOG_EXECUTION_PLAN.md) — full wave breakdown\n")
	sb.WriteString("- *(Add other plan or doc links as needed.)*\n")

	return sb.String(), nil
}

// repairPlanFile reads an existing .plan.md, recomputes frontmatter and the "## 3. Iterative Milestones" section from Todo2, and writes the file back so Cursor stripping is undone without losing other body content.
func repairPlanFile(ctx context.Context, projectRoot, planPath, planTitle string) (string, error) {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return "", fmt.Errorf("read plan file: %w", err)
	}
	content := string(data)

	// Locate frontmatter end (after second "---")
	firstDash := strings.Index(content, "---\n")
	if firstDash < 0 {
		return "", fmt.Errorf("plan file has no frontmatter")
	}
	secondDash := strings.Index(content[firstDash+4:], "\n---\n")
	if secondDash < 0 {
		return "", fmt.Errorf("plan file frontmatter not closed")
	}
	frontmatterEnd := firstDash + 4 + secondDash + len("\n---\n")

	// Locate "## 3. Iterative Milestones" and "## 4. Recommended Execution Order"
	marker3 := "## 3. Iterative Milestones"
	marker4 := "## 4. Recommended Execution Order"
	idx3 := strings.Index(content, marker3)
	idx4 := strings.Index(content, marker4)
	if idx3 < 0 || idx4 < 0 || idx4 <= idx3 {
		return "", fmt.Errorf("plan file missing ## 3 or ## 4 section")
	}

	bodyBefore := content[frontmatterEnd:idx3]
	bodyAfter := content[idx4:]

	// Recompute todoEntries and waves from Todo2 (same as generatePlanMarkdown)
	displayName := planDisplayName(planTitle)
	info, _ := getProjectInfo(projectRoot)
	overview := getStr(info, "description")
	if overview == "" {
		overview = fmt.Sprintf("Deliver and maintain %s with clear milestones and quality gates.", displayName)
	}

	store := NewDefaultTaskStore(projectRoot)
	list, _ := store.ListTasks(ctx, nil)
	tasks := tasksFromPtrs(list)
	taskByID := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	actions, _ := getNextActions(projectRoot)
	todoEntries := make([]struct{ id, content, status string }, 0, len(actions))
	for _, a := range actions {
		id := getStr(a, "task_id")
		if id == "" {
			continue
		}
		contentStr := getStr(a, "name")
		status := "pending"
		if t, ok := taskByID[id]; ok {
			if contentStr == "" {
				contentStr = t.Content
			}
			status = todo2StatusToCursorStatus(t.Status)
		}
		if contentStr == "" {
			contentStr = id
		}
		contentStr = strings.TrimSpace(contentStr)
		if len(contentStr) > 120 {
			contentStr = contentStr[:117] + "..."
		}
		todoEntries = append(todoEntries, struct{ id, content, status string }{id, contentStr, status})
	}

	_, waves, _, _ := BacklogExecutionOrder(tasks, nil)
	if max := config.MaxTasksPerWave(); max > 0 {
		waves = LimitWavesByMaxTasks(waves, max)
	}

	planDate := time.Now().Format("2006-01-02")

	// Build new frontmatter (same as generatePlanMarkdown)
	var fm strings.Builder
	fm.WriteString("---\n")
	fm.WriteString(fmt.Sprintf("name: %s Plan\n", displayName))
	fm.WriteString(fmt.Sprintf("overview: %q\n", escapeYAMLString(overview)))
	if len(todoEntries) > 0 {
		fm.WriteString("todos:\n")
		for _, e := range todoEntries {
			fm.WriteString(fmt.Sprintf("  - id: %s\n", e.id))
			fm.WriteString(fmt.Sprintf("    content: %q\n", escapeYAMLString(e.content)))
			fm.WriteString(fmt.Sprintf("    status: %s\n", e.status))
		}
	} else {
		fm.WriteString("todos: []\n")
	}
	if len(waves) > 0 {
		levelOrder := make([]int, 0, len(waves))
		for k := range waves {
			levelOrder = append(levelOrder, k)
		}
		sort.Ints(levelOrder)
		fm.WriteString("waves:\n")
		for _, level := range levelOrder {
			ids := waves[level]
			if len(ids) == 0 {
				continue
			}
			quoted := make([]string, len(ids))
			for i, id := range ids {
				quoted[i] = fmt.Sprintf("%q", id)
			}
			fm.WriteString("  - [" + strings.Join(quoted, ", ") + "]\n")
		}
	}
	fm.WriteString("isProject: true\n")
	fm.WriteString("status: draft\n")
	fm.WriteString(fmt.Sprintf("last_updated: %q\n", planDate))
	fm.WriteString("tag_hints: [planning]\n")
	fm.WriteString("referenced_by:\n")
	fm.WriteString("  - .cursor/agents/wave-task-runner.md\n")
	fm.WriteString("  - .cursor/agents/wave-verifier.md\n")
	fm.WriteString("  - .cursor/rules/plan-execution.mdc\n")
	fm.WriteString("agents:\n")
	fm.WriteString("  - name: wave-task-runner\n")
	fm.WriteString("    path: .cursor/agents/wave-task-runner.md\n")
	fm.WriteString("    role: Run one task per wave from this plan\n")
	fm.WriteString("  - name: wave-verifier\n")
	fm.WriteString("    path: .cursor/agents/wave-verifier.md\n")
	fm.WriteString("    role: Verify wave outcomes and update status\n")
	fm.WriteString("---\n\n")

	// Build new "## 3. Iterative Milestones" section with checkboxes
	reminder := "*Regenerate with: `exarp-go -tool report -args '{\"action\":\"plan\"}'`. Do not edit frontmatter in Cursor; it may strip fields and break Build.*\n\n"
	var milestones strings.Builder
	milestones.WriteString("## 3. Iterative Milestones\n\n")
	milestones.WriteString("Each milestone is independently valuable. Check off as done.\n\n")
	if len(actions) > 0 {
		for _, a := range actions {
			name := getStr(a, "name")
			taskID := getStr(a, "task_id")
			if name == "" {
				name = taskID
			}
			status := "pending"
			if t, ok := taskByID[taskID]; ok {
				status = todo2StatusToCursorStatus(t.Status)
			}
			box := "[ ]"
			if status == "completed" {
				box = "[x]"
			}
			line := fmt.Sprintf("- %s **%s** (%s)", box, name, taskID)
			if t, ok := taskByID[taskID]; ok {
				if refs := getTaskFileRefs(&t); len(refs) > 0 {
					line += " — " + strings.Join(refs, ", ")
				}
			}
			milestones.WriteString(line + "\n")
		}
	} else {
		milestones.WriteString("- [ ] *(Add Todo2 tasks to populate milestones)*\n")
	}
	milestones.WriteString("\n---\n\n")

	repaired := fm.String() + reminder + bodyBefore + milestones.String() + bodyAfter
	if err := os.WriteFile(planPath, []byte(repaired), 0644); err != nil {
		return "", fmt.Errorf("write repaired plan: %w", err)
	}
	return fmt.Sprintf("Plan format repaired: %s (frontmatter and ## 3. Iterative Milestones restored)", planPath), nil
}

// getTaskFileRefs extracts file/code references from task metadata (T-1769980664971).
// Returns paths from discovered_from, planning_doc, and planning_links.RelatedDocs.
func getTaskFileRefs(task *Todo2Task) []string {
	var refs []string

	seen := make(map[string]bool)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true

			refs = append(refs, s)
		}
	}

	if task.Metadata != nil {
		if f, ok := task.Metadata["discovered_from"].(string); ok {
			add(f)
		}

		if pd, ok := task.Metadata["planning_doc"].(string); ok {
			add(pd)
		}

		if linkMeta := GetPlanningLinkMetadata(task); linkMeta != nil {
			if linkMeta.PlanningDoc != "" {
				add(linkMeta.PlanningDoc)
			}

			for _, r := range linkMeta.RelatedDocs {
				add(r)
			}
		}
	}

	return refs
}

// planDisplayName returns a short display name from plan title (e.g. "github.com/davidl71/exarp-go" -> "exarp-go").
func planDisplayName(title string) string {
	if title == "" {
		return "Project"
	}

	if strings.Contains(title, "/") {
		return filepath.Base(title)
	}

	return title
}

// escapeYAMLString escapes a string for YAML double-quoted value.
func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")

	return strings.TrimSpace(s)
}

// todo2StatusToCursorStatus maps Todo2 status to Cursor plan todo status (pending, in_progress, completed).
func todo2StatusToCursorStatus(todo2Status string) string {
	switch strings.TrimSpace(strings.ToLower(todo2Status)) {
	case strings.ToLower(models.StatusDone):
		return "completed"
	case strings.ToLower(models.StatusInProgress), "inprogress":
		return "in_progress"
	case strings.ToLower(models.StatusReview):
		return "in_progress"
	default:
		return "pending"
	}
}

// tableCellSafe makes a string safe for a markdown table cell (no |, newlines -> space).
func tableCellSafe(s string) string {
	s = strings.ReplaceAll(s, "|", "-")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	if len(s) > 200 {
		s = s[:197] + "..."
	}

	return s
}

// aggregateProjectData aggregates all project data for overview.
func aggregateProjectData(ctx context.Context, projectRoot string, includePlanning bool) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Project info
	projectInfo, err := getProjectInfo(projectRoot)
	if err == nil {
		data["project"] = projectInfo
	}

	// Health metrics (try Go scorecard first, fallback to Python)
	if IsGoProject() {
		opts := &ScorecardOptions{FastMode: true}

		scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
		if err == nil {
			// Calculate component scores using helper functions
			scores := map[string]float64{
				"security":      calculateSecurityScore(scorecard),
				"testing":       calculateTestingScore(scorecard),
				"documentation": calculateDocumentationScore(scorecard),
				"completion":    calculateCompletionScore(scorecard),
			}
			// Alignment score would need to be calculated separately
			scores["alignment"] = 50.0 // Default, could calculate from tasks

			data["health"] = map[string]interface{}{
				"overall_score":    scorecard.Score,
				"production_ready": scorecard.Score >= float64(config.MinCoverage()), // Using coverage threshold as production ready indicator
				"scores":           scores,
			}
		}
	}

	// Codebase metrics
	codebaseMetrics, err := getCodebaseMetrics(projectRoot)
	if err == nil {
		data["codebase"] = codebaseMetrics
	}

	// Task metrics
	taskMetrics, err := getTaskMetrics(projectRoot)
	if err == nil {
		data["tasks"] = taskMetrics
	}

	// Project phases
	data["phases"] = getProjectPhases()

	// Risks and blockers
	risks, err := getRisksAndBlockers(projectRoot)
	if err == nil {
		data["risks"] = risks
	}

	// Next actions
	nextActions, err := getNextActions(projectRoot)
	if err == nil {
		data["next_actions"] = nextActions
	}

	// Optional: planning snippet (critical path + first N backlog order)
	if includePlanning {
		planning := getPlanningSnippet(projectRoot)
		if planning != nil {
			data["planning"] = planning
		}
	}

	data["generated_at"] = time.Now().Format(time.RFC3339)

	return data, nil
}

// aggregateProjectDataProto returns overview data as proto for type-safe report formatting.
func aggregateProjectDataProto(ctx context.Context, projectRoot string, includePlanning bool) (*proto.ProjectOverviewData, error) {
	pb := &proto.ProjectOverviewData{}

	if projectInfo, err := getProjectInfo(projectRoot); err == nil {
		pb.Project = ProjectInfoToProto(projectInfo)
	}

	if IsGoProject() {
		opts := &ScorecardOptions{FastMode: true}

		scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
		if err == nil {
			scores := map[string]float64{
				"security":      calculateSecurityScore(scorecard),
				"testing":       calculateTestingScore(scorecard),
				"documentation": calculateDocumentationScore(scorecard),
				"completion":    calculateCompletionScore(scorecard),
			}
			scores["alignment"] = 50.0
			pb.Health = &proto.HealthData{
				OverallScore:    scorecard.Score,
				ProductionReady: scorecard.Score >= float64(config.MinCoverage()),
				Scores:          scores,
			}
		}
	}

	if codebase, err := getCodebaseMetrics(projectRoot); err == nil {
		pb.Codebase = CodebaseMetricsToProto(codebase)
	}

	if tasks, err := getTaskMetrics(projectRoot); err == nil {
		pb.Tasks = TaskMetricsToProto(tasks)
	}

	phasesMap := getProjectPhases()
	for _, phaseRaw := range phasesMap {
		if phase, ok := phaseRaw.(map[string]interface{}); ok {
			pbPhase := &proto.ProjectPhase{}
			if name, ok := phase["name"].(string); ok {
				pbPhase.Name = name
			}

			if status, ok := phase["status"].(string); ok {
				pbPhase.Status = status
			}

			if progress, ok := phase["progress"].(int); ok {
				pbPhase.Progress = int32(progress)
			} else if progress, ok := phase["progress"].(float64); ok {
				pbPhase.Progress = int32(progress)
			}

			pb.Phases = append(pb.Phases, pbPhase)
		}
	}

	if risks, err := getRisksAndBlockers(projectRoot); err == nil {
		for _, r := range risks {
			pb.Risks = append(pb.Risks, &proto.RiskOrBlocker{
				Type:        getStr(r, "type"),
				Description: getStr(r, "description"),
				TaskId:      getStr(r, "task_id"),
				Priority:    getStr(r, "priority"),
			})
		}
	}

	if actions, err := getNextActions(projectRoot); err == nil {
		for _, a := range actions {
			hours := 0.0
			if h, ok := a["estimated_hours"].(float64); ok {
				hours = h
			}

			pb.NextActions = append(pb.NextActions, &proto.NextAction{
				TaskId:         getStr(a, "task_id"),
				Name:           getStr(a, "name"),
				Priority:       getStr(a, "priority"),
				EstimatedHours: hours,
			})
		}
	}

	pb.GeneratedAt = time.Now().Format(time.RFC3339)

	if includePlanning {
		planning := getPlanningSnippet(projectRoot)
		if planning != nil {
			pb.Planning = &proto.PlanningSnippet{}
			if s, ok := planning["critical_path_summary"].(string); ok {
				pb.Planning.CriticalPathSummary = s
			}

			if s, ok := planning["suggested_backlog_summary"].(string); ok {
				pb.Planning.SuggestedBacklogSummary = s
			}
		}
	}

	return pb, nil
}

func getStr(m map[string]interface{}, key string) string {
	return cast.ToString(m[key])
}

// planFilenameFromTitle returns a safe filename stem from a plan title (e.g. "github.com/davidl71/exarp-go" -> "exarp-go").
func planFilenameFromTitle(title string) string {
	if title == "" {
		return "plan"
	}

	stem := title
	if strings.Contains(title, "/") {
		stem = filepath.Base(title)
	}

	stem = strings.ReplaceAll(stem, " ", "-")
	stem = strings.ReplaceAll(stem, ":", "-")

	return strings.TrimSpace(stem)
}

// ensurePlanMdSuffix ensures path ends with .plan.md. If path ends with .md but not .plan.md, replaces with .plan.md; otherwise appends .plan.md.
func ensurePlanMdSuffix(path string) string {
	if path == "" {
		return path
	}

	path = filepath.Clean(path)
	base := filepath.Base(path)

	if strings.HasSuffix(strings.ToLower(base), ".plan.md") {
		return path
	}

	if strings.HasSuffix(strings.ToLower(base), ".md") {
		return filepath.Join(filepath.Dir(path), base[:len(base)-3]+".plan.md")
	}

	return filepath.Join(filepath.Dir(path), base+".plan.md")
}

// getPlanningSnippet returns critical path summary and suggested backlog order (first 10).
func getPlanningSnippet(projectRoot string) map[string]interface{} {
	out := make(map[string]interface{})

	// Critical path
	if cp, err := AnalyzeCriticalPath(projectRoot); err == nil {
		if hasCP, _ := cp["has_critical_path"].(bool); hasCP {
			if path, ok := cp["critical_path"].([]string); ok && len(path) > 0 {
				out["critical_path"] = path
				out["critical_path_summary"] = fmt.Sprintf("%d tasks: %s", len(path), strings.Join(path, ", "))
			}
		}
	}

	// First 10 in backlog execution order
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return out
	}

	tasks := tasksFromPtrs(list)

	orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil || len(orderedIDs) == 0 {
		return out
	}

	const n = 10
	if len(orderedIDs) > n {
		orderedIDs = orderedIDs[:n]
	}

	out["suggested_backlog_order"] = orderedIDs
	out["suggested_backlog_summary"] = fmt.Sprintf("First %d: %s", len(orderedIDs), strings.Join(orderedIDs, ", "))

	return out
}

// getProjectInfo extracts project metadata.
func getProjectInfo(projectRoot string) (map[string]interface{}, error) {
	info := map[string]interface{}{
		"name":        filepath.Base(projectRoot),
		"version":     "0.1.0",
		"description": "MCP Server",
		"type":        "MCP Server",
		"status":      "Active Development",
	}

	// Try to get from go.mod (using file cache)
	goModPath := filepath.Join(projectRoot, "go.mod")
	cache := cache.GetGlobalFileCache()

	if data, _, err := cache.ReadFile(goModPath); err == nil {
		content := string(data)
		if strings.Contains(content, "module ") {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
					info["name"] = moduleName

					break
				}
			}
		}
	}

	return info, nil
}

// getCodebaseMetrics collects codebase statistics.
func getCodebaseMetrics(projectRoot string) (map[string]interface{}, error) {
	var goFiles, pythonFiles, totalFiles int

	// Count files (simplified - could use more sophisticated counting)
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			// Skip hidden and vendor directories
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}

			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			goFiles++
		case ".py":
			pythonFiles++
		}

		totalFiles++

		return nil
	})

	metrics := map[string]interface{}{
		"go_files":     goFiles,
		"go_lines":     0, // Could count lines if needed
		"python_files": pythonFiles,
		"python_lines": 0, // Could count lines if needed
		"total_files":  totalFiles,
		"total_lines":  0,  // Could count lines if needed
		"tools":        28, // From registry (28 base + 1 conditional Apple FM)
		"prompts":      35, // From templates.go (19 original + 16 migrated from Python)
		"resources":    21, // From resources/handlers.go
	}

	// Count tools, prompts, resources from registry
	// These would need to be exported from registry.go or counted differently
	// For now, use values matching MIGRATION_STATUS_CURRENT.md
	metrics["tools"] = 28
	metrics["prompts"] = 34
	metrics["resources"] = 21

	return metrics, err
}

// getTaskMetrics collects task statistics.
func getTaskMetrics(projectRoot string) (map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	tasks := tasksFromPtrs(list)

	pending := 0
	completed := 0
	totalHours := 0.0

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pending++
			// Check for estimated hours in metadata
			if task.Metadata != nil {
				if hours, ok := task.Metadata["estimatedHours"].(float64); ok {
					totalHours += hours
				}
			}
		} else if IsCompletedStatus(task.Status) {
			completed++
		}
	}

	completionRate := 0.0
	if len(tasks) > 0 {
		completionRate = float64(completed) / float64(len(tasks)) * 100
	}

	return map[string]interface{}{
		"total":           len(tasks),
		"pending":         pending,
		"completed":       completed,
		"completion_rate": completionRate,
		"remaining_hours": totalHours,
	}, nil
}

// getProjectPhases returns current phase status.
func getProjectPhases() map[string]interface{} {
	return map[string]interface{}{
		"phase_1": map[string]interface{}{
			"name":     "Foundation Tools",
			"status":   "complete",
			"progress": 100,
		},
		"phase_2": map[string]interface{}{
			"name":     "Medium Complexity Tools",
			"status":   "complete",
			"progress": 100,
		},
		"phase_3": map[string]interface{}{
			"name":     "Complex Tools",
			"status":   "complete",
			"progress": 100,
		},
		"phase_4": map[string]interface{}{
			"name":     "Resources",
			"status":   "complete",
			"progress": 100,
		},
		"phase_5": map[string]interface{}{
			"name":     "Prompts",
			"status":   "complete",
			"progress": 100,
		},
	}
}

// getRisksAndBlockers identifies project risks.
func getRisksAndBlockers(projectRoot string) ([]map[string]interface{}, error) {
	risks := []map[string]interface{}{}

	// Check for incomplete tasks with high priority
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err == nil {
		tasks := tasksFromPtrs(list)
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && task.Priority == models.PriorityCritical {
				risks = append(risks, map[string]interface{}{
					"type":        "blocker",
					"description": task.Content,
					"task_id":     task.ID,
					"priority":    task.Priority,
				})
			}
		}
	}

	return risks, nil
}

// isTaskReady returns true if task is in backlog and all its dependencies are Done.
func isTaskReady(task Todo2Task, taskMap map[string]Todo2Task) bool {
	for _, depID := range task.Dependencies {
		if dep, ok := taskMap[depID]; ok && !IsCompletedStatus(dep.Status) {
			return false
		}
	}

	return true
}

// getNextActions identifies next priority actions (dependency-aware: only suggests ready tasks).
func getNextActions(projectRoot string) ([]map[string]interface{}, error) {
	actions := []map[string]interface{}{}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return actions, nil
	}

	tasks := tasksFromPtrs(list)

	orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		// Fallback: high/critical pending without order
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && (task.Priority == models.PriorityHigh || task.Priority == models.PriorityCritical) {
				estimatedHours := 0.0

				if task.Metadata != nil {
					if hours, ok := task.Metadata["estimatedHours"].(float64); ok {
						estimatedHours = hours
					}
				}

				actions = append(actions, map[string]interface{}{
					"task_id":         task.ID,
					"name":            task.Content,
					"priority":        task.Priority,
					"estimated_hours": estimatedHours,
				})
				if len(actions) >= 10 {
					break
				}
			}
		}

		return actions, nil
	}

	taskMap := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	const maxNext = 10
	for _, taskID := range orderedIDs {
		if len(actions) >= maxNext {
			break
		}

		task, ok := taskMap[taskID]
		if !ok || !IsBacklogStatus(task.Status) {
			continue
		}

		if !isTaskReady(task, taskMap) {
			continue
		}

		estimatedHours := 0.0

		if task.Metadata != nil {
			if hours, ok := task.Metadata["estimatedHours"].(float64); ok {
				estimatedHours = hours
			}
		}

		actions = append(actions, map[string]interface{}{
			"task_id":         task.ID,
			"name":            task.Content,
			"priority":        task.Priority,
			"estimated_hours": estimatedHours,
		})
	}

	return actions, nil
}

// generatePRD generates a Product Requirements Document.
func generatePRD(ctx context.Context, projectRoot, projectName string, includeArchitecture, includeMetrics, includeTasks bool) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Product Requirements Document: %s\n\n", projectName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString("---\n\n")

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("%s is an MCP (Model Context Protocol) server providing project management automation tools.\n\n", projectName))

	// Requirements from Tasks
	if includeTasks {
		sb.WriteString("## Requirements\n\n")

		store := NewDefaultTaskStore(projectRoot)

		list, err := store.ListTasks(ctx, nil)
		if err == nil {
			tasks := tasksFromPtrs(list)
			// Group tasks by priority
			criticalTasks := []Todo2Task{}
			highTasks := []Todo2Task{}
			mediumTasks := []Todo2Task{}

			for _, task := range tasks {
				if IsPendingStatus(task.Status) {
					switch task.Priority {
					case models.PriorityCritical:
						criticalTasks = append(criticalTasks, task)
					case models.PriorityHigh:
						highTasks = append(highTasks, task)
					case models.PriorityMedium:
						mediumTasks = append(mediumTasks, task)
					}
				}
			}

			if len(criticalTasks) > 0 {
				sb.WriteString("### Critical Priority\n\n")

				for _, task := range criticalTasks {
					sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", task.Content, task.ID, task.LongDescription))
				}

				sb.WriteString("\n")
			}

			if len(highTasks) > 0 {
				sb.WriteString("### High Priority\n\n")

				for _, task := range highTasks {
					sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", task.Content, task.ID, task.LongDescription))
				}

				sb.WriteString("\n")
			}
		}
	}

	// Architecture
	if includeArchitecture {
		sb.WriteString("## Architecture\n\n")
		sb.WriteString("### System Components\n\n")
		sb.WriteString("- **MCP Server**: Go-based server implementing MCP protocol\n")
		sb.WriteString("- **Tool Handlers**: Native Go implementations for project management tools\n")
		sb.WriteString("- **Python Bridge**: Subprocess bridge for remaining Python tools\n")
		sb.WriteString("- **Database**: SQLite for Todo2 task storage\n")
		sb.WriteString("- **Resources**: Memory and scorecard resources\n\n")
	}

	// Metrics
	if includeMetrics {
		sb.WriteString("## Metrics\n\n")

		metrics, err := getCodebaseMetrics(projectRoot)
		if err == nil {
			sb.WriteString(fmt.Sprintf("- **Total Files**: %d\n", metrics["total_files"]))
			sb.WriteString(fmt.Sprintf("- **Go Files**: %d\n", metrics["go_files"]))
			sb.WriteString(fmt.Sprintf("- **Tools**: %d\n", metrics["tools"]))
			sb.WriteString(fmt.Sprintf("- **Prompts**: %d\n", metrics["prompts"]))
			sb.WriteString(fmt.Sprintf("- **Resources**: %d\n", metrics["resources"]))
		}

		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// formatOverviewText formats overview as plain text.
func formatOverviewText(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", project["status"]))
		sb.WriteString("\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("Health Scorecard:\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", score))
		}

		if ready, ok := health["production_ready"].(bool); ok {
			if ready {
				sb.WriteString("  Production Ready: YES ✅\n")
			} else {
				sb.WriteString("  Production Ready: NO ❌\n")
			}
		}

		sb.WriteString("\n")
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", tasks["completed"]))

		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", rate))
		}

		if hours, ok := tasks["remaining_hours"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", hours))
		}

		sb.WriteString("\n")
	}

	// Next Actions
	if actions, ok := data["next_actions"].([]map[string]interface{}); ok && len(actions) > 0 {
		sb.WriteString("Next Actions:\n")

		for i, action := range actions {
			if i >= 5 {
				break
			}

			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action["name"], action["priority"]))
		}

		sb.WriteString("\n")
	}

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("Critical Path: " + summary + "\n\n")
		}

		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("Suggested Backlog Order: " + summary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewMarkdown formats overview as markdown.
func formatOverviewMarkdown(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", project["status"]))
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("## Health Scorecard\n\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", score))
		}
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", tasks["completed"]))

		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", rate))
		}
	}

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + summary + "\n\n")
		}

		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + summary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewHTML formats overview as HTML.
func formatOverviewHTML(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", project["name"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", project["version"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", project["type"]))
		sb.WriteString("</ul>\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Health Scorecard</h2>\n")

		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", score))
		}
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

// formatOverviewTextProto formats overview from proto (type-safe, no map assertions).
func formatOverviewTextProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	if pb.Project != nil {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", pb.Project.Status))
		sb.WriteString("\n")
	}

	if pb.Health != nil {
		sb.WriteString("Health Scorecard:\n")
		sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", pb.Health.OverallScore))

		if pb.Health.ProductionReady {
			sb.WriteString("  Production Ready: YES ✅\n")
		} else {
			sb.WriteString("  Production Ready: NO ❌\n")
		}

		sb.WriteString("\n")
	}

	if pb.Tasks != nil {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", pb.Tasks.CompletionRate))
		sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", pb.Tasks.RemainingHours))
		sb.WriteString("\n")
	}

	if len(pb.NextActions) > 0 {
		sb.WriteString("Next Actions:\n")

		for i, action := range pb.NextActions {
			if i >= 5 {
				break
			}

			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action.Name, action.Priority))
		}

		sb.WriteString("\n")
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("Critical Path: " + pb.Planning.CriticalPathSummary + "\n\n")
		}

		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("Suggested Backlog Order: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// GetOverviewText returns project overview as plain text for TUI/CLI display.
// It aggregates project data (health when Go project, tasks, codebase, etc.) and formats as text.
func GetOverviewText(ctx context.Context, projectRoot string) (string, error) {
	pb, err := aggregateProjectDataProto(ctx, projectRoot, false)
	if err != nil {
		return "", err
	}

	return formatOverviewTextProto(pb), nil
}

// formatOverviewMarkdownProto formats overview as markdown from proto.
func formatOverviewMarkdownProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	if pb.Project != nil {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", pb.Project.Status))
	}

	if pb.Health != nil {
		sb.WriteString("## Health Scorecard\n\n")
		sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", pb.Health.OverallScore))
	}

	if pb.Tasks != nil {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", pb.Tasks.CompletionRate))
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + pb.Planning.CriticalPathSummary + "\n\n")
		}

		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewHTMLProto formats overview as HTML from proto.
func formatOverviewHTMLProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	if pb.Project != nil {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", pb.Project.Type))
		sb.WriteString("</ul>\n")
	}

	if pb.Health != nil {
		sb.WriteString("<h2>Health Scorecard</h2>\n")
		sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", pb.Health.OverallScore))
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

// getFloatParam safely extracts float64 from params.
func getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := params[key].(float64); ok {
		return val
	}

	return defaultValue
}
