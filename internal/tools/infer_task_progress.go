// infer_task_progress.go — MCP "infer_task_progress" tool: heuristic task completion detection.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
)

// Default file extensions for evidence gathering (match Python default).
var defaultInferTaskProgressExtensions = []string{".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".java", ".cs", ".rs"}

// InferredResult is one task's completion inference (confidence and evidence).
type InferredResult struct {
	TaskID     string   `json:"task_id"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// InferTaskProgressResponseToMap converts InferTaskProgressResponse proto to map for response.FormatResult.
func InferTaskProgressResponseToMap(resp *proto.InferTaskProgressResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{}
	if resp.GetOutputPath() != "" {
		out["output_path"] = resp.GetOutputPath()
	}

	if resp.GetResultJson() != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(resp.GetResultJson()), &payload) == nil {
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	return out
}

// handleInferTaskProgressNative runs task completion inference (heuristics only in Iteration 1).
// Loads tasks by status filter (default: In Progress), gathers codebase evidence, scores each task, returns JSON.
// Does not call database.UpdateTask in this iteration.
func handleInferTaskProgressNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	if root, ok := params["project_root"].(string); ok && root != "" {
		projectRoot = root
	}

	scanDepth := 3
	if d, ok := params["scan_depth"].(float64); ok && d >= 1 && d <= 5 {
		scanDepth = int(d)
	}

	confidenceThreshold := 0.7
	if t, ok := params["confidence_threshold"].(float64); ok && t >= 0 && t <= 1 {
		confidenceThreshold = t
	}

	extensions := defaultInferTaskProgressExtensions
	if ex, ok := params["file_extensions"].([]interface{}); ok && len(ex) > 0 {
		extensions = make([]string, 0, len(ex))

		for _, e := range ex {
			if s, ok := e.(string); ok && s != "" {
				extensions = append(extensions, s)
			}
		}
	}

	if len(extensions) == 0 {
		extensions = defaultInferTaskProgressExtensions
	}

	// Default to In Progress, but allow filtering by other statuses (e.g., Todo)
	statusFilter := database.StatusInProgress
	if sf, ok := params["status_filter"].(string); ok && sf != "" {
		statusFilter = sf
	}

	candidates, err := loadTasksByStatus(ctx, projectRoot, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	evidence, err := GatherEvidence(projectRoot, scanDepth, extensions)
	if err != nil {
		return nil, fmt.Errorf("failed to gather evidence: %w", err)
	}

	allScored := scoreAllTasksHeuristic(candidates, evidence)

	useFM := true
	if u, ok := params["use_fm"].(bool); ok {
		useFM = u
	}

	if useFM && FMAvailable() {
		allScored = enhanceWithFM(ctx, candidates, allScored, evidence)
	}

	inferred := filterByThreshold(allScored, confidenceThreshold)

	dryRun := true
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	autoUpdate := false
	if au, ok := params["auto_update_tasks"].(bool); ok {
		autoUpdate = au
	}

	tasksUpdated := 0
	if autoUpdate && !dryRun && len(inferred) > 0 {
		tasksUpdated, err = applyInferredCompletions(ctx, projectRoot, inferred, candidates)
		if err != nil {
			return nil, fmt.Errorf("apply inferred completions: %w", err)
		}
	}

	method := "native_go_heuristics"
	if useFM && FMAvailable() {
		method = "native_go_heuristics_and_fm"
	}

	result := map[string]interface{}{
		"success":              true,
		"total_tasks_analyzed": len(candidates),
		"inferences_made":      len(inferred),
		"tasks_updated":        tasksUpdated,
		"inferred_results":     inferred,
		"method":               method,
		"dry_run":              dryRun,
		"auto_update":          autoUpdate,
	}

	outputPath := ""
	if p, ok := params["output_path"].(string); ok && p != "" {
		outputPath = p
	}

	if outputPath != "" {
		if writeErr := writeInferReport(outputPath, result, dryRun, method); writeErr != nil {
			return nil, fmt.Errorf("write report: %w", writeErr)
		}
	}

	resultJSON, _ := json.Marshal(result)
	resp := &proto.InferTaskProgressResponse{OutputPath: outputPath, ResultJson: string(resultJSON)}

	return framework.FormatResult(InferTaskProgressResponseToMap(resp), resp.GetOutputPath())
}

// loadTasksByStatus returns tasks filtered by status via TaskStore.
// Supports any valid status: Todo, In Progress, Review, Done, Cancelled.
func loadTasksByStatus(ctx context.Context, projectRoot string, status string) ([]Todo2Task, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, &database.TaskFilters{Status: &status})
	if err != nil {
		return nil, err
	}

	return tasksFromPtrs(list), nil
}

// loadInProgressTasks returns only In Progress tasks via TaskStore (legacy wrapper).
func loadInProgressTasks(ctx context.Context, projectRoot string) ([]Todo2Task, error) {
	return loadTasksByStatus(ctx, projectRoot, database.StatusInProgress)
}

var wordTokenRe = regexp.MustCompile(`[a-zA-Z0-9_]{2,}`)

// scoreAllTasksHeuristic scores every task (confidence 0-1 and evidence). No threshold filter.
func scoreAllTasksHeuristic(tasks []Todo2Task, evidence *CodebaseEvidence) []InferredResult {
	if evidence == nil {
		return nil
	}

	pathLower := make([]string, len(evidence.Paths))
	for i, p := range evidence.Paths {
		pathLower[i] = strings.ToLower(p)
	}

	results := make([]InferredResult, 0, len(tasks))

	for _, task := range tasks {
		text := strings.ToLower(task.Content + " " + task.LongDescription)
		tokens := wordTokenRe.FindAllString(text, -1)

		var evidenceList []string

		pathMatches := 0
		snippetMatches := 0
		seenPath := make(map[string]bool)
		seenSnippet := make(map[string]bool)

		if len(tokens) > 0 {
			tokenSet := make(map[string]bool)
			for _, t := range tokens {
				tokenSet[t] = true
			}

			for i, p := range pathLower {
				relPath := evidence.Paths[i]

				for tok := range tokenSet {
					if strings.Contains(p, tok) {
						pathMatches++

						if !seenPath[relPath] {
							seenPath[relPath] = true

							evidenceList = append(evidenceList, "path:"+relPath)
						}

						break
					}
				}
			}

			for path, snippet := range evidence.Snippets {
				snipLower := strings.ToLower(snippet)
				for tok := range tokenSet {
					if strings.Contains(snipLower, tok) {
						snippetMatches++

						if !seenSnippet[path] {
							seenSnippet[path] = true

							evidenceList = append(evidenceList, "snippet:"+path)
						}

						break
					}
				}
			}
		}

		confidence := 0.0

		if pathMatches > 0 || snippetMatches > 0 {
			pmCap := pathMatches
			if pmCap > 3 {
				pmCap = 3
			}

			smCap := snippetMatches
			if smCap > 3 {
				smCap = 3
			}

			confidence = 0.3 + 0.4*float64(pmCap)/3 + 0.3*float64(smCap)/3
			if confidence > 1.0 {
				confidence = 1.0
			}
		}

		results = append(results, InferredResult{
			TaskID:     task.ID,
			Confidence: confidence,
			Evidence:   evidenceList,
		})
	}

	return results
}

// scoreTasksHeuristic scores each task and returns only those with confidence >= threshold (for tests).
func scoreTasksHeuristic(tasks []Todo2Task, evidence *CodebaseEvidence, threshold float64) []InferredResult {
	all := scoreAllTasksHeuristic(tasks, evidence)
	return filterByThreshold(all, threshold)
}

func filterByThreshold(results []InferredResult, threshold float64) []InferredResult {
	out := make([]InferredResult, 0)

	for _, r := range results {
		if r.Confidence >= threshold {
			out = append(out, r)
		}
	}

	return out
}

// enhanceWithFM updates confidence (and optionally evidence) using DefaultFMProvider when available.
// On FM error for a task, heuristic confidence is left unchanged.
func enhanceWithFM(ctx context.Context, tasks []Todo2Task, scored []InferredResult, evidence *CodebaseEvidence) []InferredResult {
	taskByID := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskByID[t.ID] = t
	}

	out := make([]InferredResult, len(scored))
	copy(out, scored)

	for i := range out {
		task, ok := taskByID[out[i].TaskID]
		if !ok {
			continue
		}

		prompt := buildFMCompletionPrompt(task, evidence)
		model := DefaultModelRouter.SelectModel("general", ModelRequirements{})

		genCtx, genCancel := context.WithTimeout(ctx, config.OllamaGenerateTimeout())
		resp, err := DefaultModelRouter.Generate(genCtx, model, prompt, 300, 0.2)
		genCancel()
		if err != nil {
			continue
		}

		fmConf, fmEvidence := parseFMCompletionResponse(resp, out[i].TaskID)
		if fmConf >= 0 {
			if fmConf > out[i].Confidence {
				out[i].Confidence = fmConf
			}

			if fmConf > 1.0 {
				out[i].Confidence = 1.0
			}

			for _, e := range fmEvidence {
				out[i].Evidence = append(out[i].Evidence, "fm:"+e)
			}
		}
	}

	return out
}

func buildFMCompletionPrompt(task Todo2Task, evidence *CodebaseEvidence) string {
	snippet := task.Content
	if task.LongDescription != "" {
		snippet += " " + task.LongDescription
	}

	pathsLine := ""

	if evidence != nil && len(evidence.Paths) > 0 {
		maxPaths := 20
		if len(evidence.Paths) < maxPaths {
			maxPaths = len(evidence.Paths)
		}

		pathsLine = strings.Join(evidence.Paths[:maxPaths], ", ")
	}

	return fmt.Sprintf(`Given this task and codebase paths, is the task complete? Task ID: %s. Task: %s. Codebase paths: %s. Return JSON only: {"task_id": "%s", "complete": true or false, "confidence": 0.0-1.0, "evidence": ["..."]}`,
		task.ID, snippet, pathsLine, task.ID)
}

func parseFMCompletionResponse(resp string, taskID string) (confidence float64, evidence []string) {
	confidence = -1

	candidate := resp
	if idx := strings.Index(candidate, "{"); idx >= 0 {
		candidate = candidate[idx:]
	}

	if idx := strings.LastIndex(candidate, "}"); idx >= 0 {
		candidate = candidate[:idx+1]
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(candidate), &m); err != nil {
		return -1, nil
	}

	if id, _ := m["task_id"].(string); id != taskID {
		return -1, nil
	}

	if c, ok := m["confidence"].(float64); ok {
		confidence = c
	}

	if ev, ok := m["evidence"].([]interface{}); ok {
		for _, e := range ev {
			if s, ok := e.(string); ok {
				evidence = append(evidence, s)
			}
		}
	}

	return confidence, evidence
}

// applyInferredCompletions marks inferred tasks as Done. Uses DB when available, else JSON.
// Returns the number of tasks updated and any error.
func applyInferredCompletions(ctx context.Context, projectRoot string, inferred []InferredResult, candidates []Todo2Task) (int, error) {
	taskByID := make(map[string]Todo2Task)
	for _, t := range candidates {
		taskByID[t.ID] = t
	}

	store := NewDefaultTaskStore(projectRoot)
	updated := 0

	for _, r := range inferred {
		task, err := store.GetTask(ctx, r.TaskID)
		if err != nil || task == nil {
			continue
		}

		task.Status = database.StatusDone
		task.Completed = true

		if err := store.UpdateTask(ctx, task); err != nil {
			continue
		}

		updated++
	}

	return updated, nil
}

// writeInferReport writes a markdown report to outputPath (same structure as Python report).
func writeInferReport(outputPath string, result map[string]interface{}, dryRun bool, method string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	total, _ := result["total_tasks_analyzed"].(int)
	inferences, _ := result["inferences_made"].(int)
	updated, _ := result["tasks_updated"].(int)
	inferredList, _ := result["inferred_results"].([]InferredResult)

	sb := strings.Builder{}
	sb.WriteString("# Task Completion Check Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tasks Analyzed:** %d\n", total))
	sb.WriteString(fmt.Sprintf("- **Inferences Made:** %d\n", inferences))
	sb.WriteString(fmt.Sprintf("- **Tasks Updated:** %d\n", updated))
	sb.WriteString(fmt.Sprintf("- **Method:** %s\n", method))
	sb.WriteString(fmt.Sprintf("- **Dry Run:** %v\n\n", dryRun))
	sb.WriteString("## Inferred Completions\n\n")

	for _, r := range inferredList {
		sb.WriteString(fmt.Sprintf("### Task %s\n", r.TaskID))
		sb.WriteString(fmt.Sprintf("- **Confidence:** %.1f%%\n", r.Confidence*100))
		sb.WriteString(fmt.Sprintf("- **Evidence:** %d items\n", len(r.Evidence)))

		maxEv := 3
		if len(r.Evidence) < maxEv {
			maxEv = len(r.Evidence)
		}

		for _, ev := range r.Evidence[:maxEv] {
			sb.WriteString(fmt.Sprintf("  - %s\n", ev))
		}

		sb.WriteString("\n")
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}
