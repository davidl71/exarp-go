package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
)

// MetadataKeyPreferredBackend is the task metadata key for local AI backend preference (fm, mlx, ollama).
const MetadataKeyPreferredBackend = "preferred_backend"

// MetadataKeyRecommendedTools is the task metadata key for recommended MCP tools (e.g. tractatus_thinking, context7).
const MetadataKeyRecommendedTools = "recommended_tools"

// GetRecommendedTools returns the recommended tool IDs from task metadata, or nil if unset.
func GetRecommendedTools(metadata map[string]interface{}) []string {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata[MetadataKeyRecommendedTools]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// GetPreferredBackend returns the preferred local AI backend from task metadata, or "" if unset.
// Valid values: "fm", "mlx", "ollama".
func GetPreferredBackend(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}

	s, _ := metadata[MetadataKeyPreferredBackend].(string)
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "fm", "mlx", "ollama":
		return s
	default:
		return ""
	}
}

// HistoricalTask represents a completed task for historical analysis.
type HistoricalTask struct {
	Name           string
	Details        string
	Tags           []string
	Priority       string
	EstimatedHours float64
	ActualHours    float64
	Created        string
	CompletedAt    string
}

// EstimationResult represents the result of task duration estimation.
type EstimationResult struct {
	EstimateHours float64                `json:"estimate_hours"`
	Confidence    float64                `json:"confidence"`
	Method        string                 `json:"method"`
	LowerBound    float64                `json:"lower_bound"`
	UpperBound    float64                `json:"upper_bound"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// estimateStatistically estimates task duration using statistical methods.
func estimateStatistically(projectRoot, name, details string, tags []string, priority string, useHistorical bool) (*EstimationResult, error) {
	text := strings.ToLower(name + " " + details)

	var historicalEstimate *float64

	var historicalConfidence float64

	// Strategy 1: Historical data matching
	if useHistorical {
		historical, err := loadHistoricalTasks(projectRoot)
		if err == nil && len(historical) > 0 {
			est, conf := estimateFromHistory(text, tags, priority, historical)
			if est != nil {
				historicalEstimate = est
				historicalConfidence = conf
			}
		}
	}

	// Strategy 2: Keyword-based heuristic
	heuristicEstimate := estimateFromKeywords(text)
	heuristicConfidence := 0.3

	// Strategy 3: Priority-based adjustment
	priorityMultiplier := getPriorityMultiplier(priority)

	// Combine estimates
	var baseEstimate float64

	var confidence float64

	var method string

	if historicalEstimate != nil && historicalConfidence > 0.2 {
		// Use historical estimate (already includes actual work time)
		baseEstimate = *historicalEstimate
		confidence = historicalConfidence
		method = "historical_match"
	} else {
		// Use keyword heuristic
		baseEstimate = heuristicEstimate
		confidence = heuristicConfidence
		method = "keyword_heuristic"
	}

	// Apply priority multiplier once (not twice!)
	finalEstimate := baseEstimate * priorityMultiplier

	// Cap estimates at reasonable maximum (20 hours for this tool)
	// Most tasks should be much shorter
	finalEstimate = math.Min(20.0, finalEstimate)

	// Round to 1 decimal place
	finalEstimate = math.Round(finalEstimate*10) / 10

	// Calculate confidence interval
	stdDev := finalEstimate * 0.3
	lowerBound := math.Max(0.5, finalEstimate-1.96*stdDev)
	upperBound := finalEstimate + 1.96*stdDev

	return &EstimationResult{
		EstimateHours: finalEstimate,
		Confidence:    math.Min(0.95, confidence),
		Method:        method,
		LowerBound:    math.Round(lowerBound*10) / 10,
		UpperBound:    math.Round(upperBound*10) / 10,
		Metadata: map[string]interface{}{
			"historical_match":      historicalEstimate != nil,
			"historical_confidence": historicalConfidence,
			"heuristic_estimate":    heuristicEstimate,
			"priority_multiplier":   priorityMultiplier,
		},
	}, nil
}

// handleEstimationAnalyze handles the analyze action
// Analyzes estimation accuracy by comparing estimated vs actual hours from historical tasks.
func handleEstimationAnalyze(projectRoot string, params map[string]interface{}) (string, error) {
	historical, err := loadHistoricalTasks(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load historical data: %w", err)
	}

	// Filter tasks that have both estimated and actual hours
	completedTasks := make([]struct {
		name           string
		tags           []string
		priority       string
		estimatedHours float64
		actualHours    float64
		error          float64
		errorPct       float64
		absErrorPct    float64
	}, 0)

	for _, task := range historical {
		if task.EstimatedHours > 0 && task.ActualHours > 0 {
			error := task.ActualHours - task.EstimatedHours
			errorPct := (error / task.EstimatedHours) * 100.0
			absErrorPct := math.Abs(errorPct)

			completedTasks = append(completedTasks, struct {
				name           string
				tags           []string
				priority       string
				estimatedHours float64
				actualHours    float64
				error          float64
				errorPct       float64
				absErrorPct    float64
			}{
				name:           task.Name,
				tags:           task.Tags,
				priority:       task.Priority,
				estimatedHours: task.EstimatedHours,
				actualHours:    task.ActualHours,
				error:          error,
				errorPct:       errorPct,
				absErrorPct:    absErrorPct,
			})
		}
	}

	if len(completedTasks) == 0 {
		return `{"success": false, "message": "No completed tasks with both estimates and actuals", "completed_tasks_count": 0}`, nil
	}

	// Calculate overall accuracy metrics
	errors := make([]float64, len(completedTasks))
	errorPcts := make([]float64, len(completedTasks))
	absErrorPcts := make([]float64, len(completedTasks))

	for i, task := range completedTasks {
		errors[i] = task.error
		errorPcts[i] = task.errorPct
		absErrorPcts[i] = task.absErrorPct
	}

	// Mean error
	meanError := 0.0
	for _, e := range errors {
		meanError += e
	}

	meanError /= float64(len(errors))

	// Median error
	sortedErrors := make([]float64, len(errors))
	copy(sortedErrors, errors)
	sort.Float64s(sortedErrors)

	medianError := sortedErrors[len(sortedErrors)/2]
	if len(sortedErrors)%2 == 0 {
		medianError = (sortedErrors[len(sortedErrors)/2-1] + sortedErrors[len(sortedErrors)/2]) / 2
	}

	// Mean absolute error
	meanAbsoluteError := 0.0
	for _, e := range errors {
		meanAbsoluteError += math.Abs(e)
	}

	meanAbsoluteError /= float64(len(errors))

	// Mean error percentage
	meanErrorPct := 0.0
	for _, e := range errorPcts {
		meanErrorPct += e
	}

	meanErrorPct /= float64(len(errorPcts))

	// Mean absolute error percentage
	meanAbsErrorPct := 0.0
	for _, e := range absErrorPcts {
		meanAbsErrorPct += e
	}

	meanAbsErrorPct /= float64(len(absErrorPcts))

	// Count over/under estimations
	overEstimatedCount := 0
	underEstimatedCount := 0
	accurateCount := 0

	for _, task := range completedTasks {
		if task.error < 0 {
			overEstimatedCount++
		} else if task.error > 0 {
			underEstimatedCount++
		}

		if task.absErrorPct < 20 {
			accurateCount++
		}
	}

	accuracyMetrics := map[string]interface{}{
		"total_tasks":                    len(completedTasks),
		"mean_error":                     math.Round(meanError*100) / 100,
		"median_error":                   math.Round(medianError*100) / 100,
		"mean_absolute_error":            math.Round(meanAbsoluteError*100) / 100,
		"mean_error_percentage":          math.Round(meanErrorPct*100) / 100,
		"mean_absolute_error_percentage": math.Round(meanAbsErrorPct*100) / 100,
		"over_estimated_count":           overEstimatedCount,
		"under_estimated_count":          underEstimatedCount,
		"accurate_count":                 accurateCount,
		"accuracy_rate_percentage":       math.Round(float64(accurateCount)/float64(len(completedTasks))*100*100) / 100,
	}

	// Analyze by tag
	tagAccuracy := analyzeByTag(completedTasks)

	// Analyze by priority
	priorityAccuracy := analyzeByPriority(completedTasks)

	// Build result
	result := map[string]interface{}{
		"success":               true,
		"completed_tasks_count": len(completedTasks),
		"accuracy_metrics":      accuracyMetrics,
		"tag_accuracy":          tagAccuracy,
		"priority_accuracy":     priorityAccuracy,
		"generated":             time.Now().Format(time.RFC3339),
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// analyzeByTag analyzes estimation accuracy by tag.
func analyzeByTag(completedTasks []struct {
	name           string
	tags           []string
	priority       string
	estimatedHours float64
	actualHours    float64
	error          float64
	errorPct       float64
	absErrorPct    float64
}) map[string]interface{} {
	tagTasks := make(map[string][]struct {
		name           string
		tags           []string
		priority       string
		estimatedHours float64
		actualHours    float64
		error          float64
		errorPct       float64
		absErrorPct    float64
	})

	// Group tasks by tag
	for _, task := range completedTasks {
		if len(task.tags) == 0 {
			tagTasks["untagged"] = append(tagTasks["untagged"], task)
			continue
		}

		for _, tag := range task.tags {
			tagTasks[tag] = append(tagTasks[tag], task)
		}
	}

	tagStats := make(map[string]interface{})

	for tag, tasks := range tagTasks {
		if len(tasks) == 0 {
			continue
		}

		meanError := 0.0
		meanAbsErrorPct := 0.0

		for _, task := range tasks {
			meanError += task.error
			meanAbsErrorPct += task.absErrorPct
		}

		meanError /= float64(len(tasks))
		meanAbsErrorPct /= float64(len(tasks))

		tagStats[tag] = map[string]interface{}{
			"count":                          len(tasks),
			"mean_error":                     math.Round(meanError*100) / 100,
			"mean_absolute_error_percentage": math.Round(meanAbsErrorPct*100) / 100,
		}
	}

	return tagStats
}

// analyzeByPriority analyzes estimation accuracy by priority.
func analyzeByPriority(completedTasks []struct {
	name           string
	tags           []string
	priority       string
	estimatedHours float64
	actualHours    float64
	error          float64
	errorPct       float64
	absErrorPct    float64
}) map[string]interface{} {
	priorityTasks := make(map[string][]struct {
		name           string
		tags           []string
		priority       string
		estimatedHours float64
		actualHours    float64
		error          float64
		errorPct       float64
		absErrorPct    float64
	})

	// Group tasks by priority
	for _, task := range completedTasks {
		priority := task.priority
		if priority == "" {
			priority = "medium"
		}

		priorityTasks[priority] = append(priorityTasks[priority], task)
	}

	priorityStats := make(map[string]interface{})

	for priority, tasks := range priorityTasks {
		if len(tasks) == 0 {
			continue
		}

		meanError := 0.0
		meanAbsErrorPct := 0.0

		for _, task := range tasks {
			meanError += task.error
			meanAbsErrorPct += task.absErrorPct
		}

		meanError /= float64(len(tasks))
		meanAbsErrorPct /= float64(len(tasks))

		priorityStats[priority] = map[string]interface{}{
			"count":                          len(tasks),
			"mean_error":                     math.Round(meanError*100) / 100,
			"mean_absolute_error_percentage": math.Round(meanAbsErrorPct*100) / 100,
		}
	}

	return priorityStats
}

// handleEstimationStats handles the stats action.
func handleEstimationStats(projectRoot string, params map[string]interface{}) (string, error) {
	historical, err := loadHistoricalTasks(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load historical data: %w", err)
	}

	if len(historical) == 0 {
		return `{"count": 0, "message": "No historical data available"}`, nil
	}

	actualHours := make([]float64, len(historical))
	for i, task := range historical {
		actualHours[i] = task.ActualHours
	}

	// Calculate statistics
	sum := 0.0
	for _, h := range actualHours {
		sum += h
	}

	mean := sum / float64(len(actualHours))

	// Sort for median
	sorted := make([]float64, len(actualHours))
	copy(sorted, actualHours)
	sort.Float64s(sorted)

	median := sorted[len(sorted)/2]
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}

	min := sorted[0]
	max := sorted[len(sorted)-1]

	// Standard deviation
	variance := 0.0
	for _, h := range actualHours {
		variance += (h - mean) * (h - mean)
	}

	stdDev := math.Sqrt(variance / float64(len(actualHours)))

	stats := map[string]interface{}{
		"count":  len(historical),
		"mean":   math.Round(mean*100) / 100,
		"median": math.Round(median*100) / 100,
		"stdev":  math.Round(stdDev*100) / 100,
		"min":    math.Round(min*100) / 100,
		"max":    math.Round(max*100) / 100,
	}

	resultJSON, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal stats: %w", err)
	}

	return string(resultJSON), nil
}

const estimateBatchMaxTasks = 50

// handleEstimationBatch estimates duration for multiple tasks (statistical only; cap at estimateBatchMaxTasks).
// Params: task_ids (array or comma-separated string), or status_filter (e.g. "Todo") to estimate all matching tasks.
func handleEstimationBatch(projectRoot string, params map[string]interface{}) (string, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Resolve task set: by IDs or by status filter
	var target []Todo2Task

	if idsRaw, ok := params["task_ids"]; ok && idsRaw != nil {
		var ids []string

		switch v := idsRaw.(type) {
		case []interface{}:
			for _, x := range v {
				if s, ok := x.(string); ok && s != "" {
					ids = append(ids, s)
				}
			}
		case string:
			if v != "" {
				for _, s := range strings.Split(v, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						ids = append(ids, s)
					}
				}
			}
		}

		idSet := make(map[string]bool)
		for _, id := range ids {
			idSet[id] = true
		}

		for _, t := range tasks {
			if idSet[t.ID] {
				target = append(target, t)
			}
		}
	} else if statusFilter, ok := params["status_filter"].(string); ok && statusFilter != "" {
		for _, t := range tasks {
			if t.Status == statusFilter {
				target = append(target, t)
			}
		}
	} else {
		// Default: all Todo
		for _, t := range tasks {
			if t.Status == "Todo" {
				target = append(target, t)
			}
		}
	}

	if len(target) == 0 {
		return `{"total_tasks": 0, "total_hours": 0, "by_priority": {}, "estimates": []}`, nil
	}

	if len(target) > estimateBatchMaxTasks {
		target = target[:estimateBatchMaxTasks]
	}

	useHistorical := true
	if useHist, ok := params["use_historical"].(bool); ok {
		useHistorical = useHist
	}

	totalHours := 0.0
	byPriority := make(map[string]float64)
	estimates := make([]map[string]interface{}, 0, len(target))

	for _, task := range target {
		details := task.LongDescription
		if details == "" {
			details = task.Content
		}

		priority := task.Priority
		if priority == "" {
			priority = "medium"
		}

		res, err := estimateStatistically(projectRoot, task.Content, details, task.Tags, priority, useHistorical)
		if err != nil {
			estimates = append(estimates, map[string]interface{}{
				"task_id":        task.ID,
				"content":        task.Content,
				"estimate_hours": 0.0,
				"method":         "error",
				"error":          err.Error(),
			})

			continue
		}

		totalHours += res.EstimateHours
		byPriority[priority] = byPriority[priority] + res.EstimateHours
		estimates = append(estimates, map[string]interface{}{
			"task_id":        task.ID,
			"content":        task.Content,
			"estimate_hours": res.EstimateHours,
			"method":         res.Method,
		})
	}

	result := map[string]interface{}{
		"total_tasks": len(estimates),
		"total_hours": math.Round(totalHours*10) / 10,
		"by_priority": byPriority,
		"estimates":   estimates,
		"capped":      len(target) == estimateBatchMaxTasks,
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal batch result: %w", err)
	}

	return string(resultJSON), nil
}

// loadHistoricalTasks loads completed tasks from Todo2 (DB-first, then JSON fallback).
// When the project uses SQLite, Done tasks are loaded from the database with
// estimation columns (created, last_modified, completed_at, estimated_hours, actual_hours).
// If the database is unavailable or returns no usable rows, falls back to .todo2/state.todo2.json.
func loadHistoricalTasks(projectRoot string) ([]HistoricalTask, error) {
	// Try database first (same pattern as LoadTodo2Tasks)
	if list, err := database.GetDoneTasksForEstimation(context.Background()); err == nil && len(list) > 0 {
		historical := make([]HistoricalTask, 0, len(list))

		for _, task := range list {
			actualHours := task.ActualHours
			if actualHours == 0 {
				completedAt := task.CompletedAt
				if completedAt == "" {
					completedAt = task.LastModified
				}

				if task.Created != "" && completedAt != "" {
					if createdTime, err := time.Parse(time.RFC3339, task.Created); err == nil {
						if completedTime, err := time.Parse(time.RFC3339, completedAt); err == nil {
							duration := completedTime.Sub(createdTime)
							calendarHours := duration.Hours()
							days := calendarHours / 24.0
							estimatedWorkHours := math.Min(calendarHours, days*8.0)
							estimatedWorkHours = math.Min(16.0, estimatedWorkHours)
							actualHours = math.Max(0.1, estimatedWorkHours)
						}
					}
				}
			}

			if actualHours > 0 {
				name := task.Content
				if name == "" {
					name = task.ID
				}

				historical = append(historical, HistoricalTask{
					Name:           name,
					Details:        task.LongDescription,
					Tags:           task.Tags,
					Priority:       task.Priority,
					EstimatedHours: task.EstimatedHours,
					ActualHours:    actualHours,
					Created:        task.Created,
					CompletedAt:    task.CompletedAt,
				})
			}
		}

		if len(historical) > 0 {
			return historical, nil
		}
	}

	// Fallback: load from JSON file (e.g. DB unavailable, or no estimation columns populated)
	return loadHistoricalTasksFromJSON(projectRoot)
}

// loadHistoricalTasksFromJSON loads completed tasks from .todo2/state.todo2.json.
// Used when the database is unavailable or returns no Done tasks with estimation data.
func loadHistoricalTasksFromJSON(projectRoot string) ([]HistoricalTask, error) {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")

	data, err := os.ReadFile(todo2Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoricalTask{}, nil
		}

		return nil, fmt.Errorf("failed to read Todo2 file: %w", err)
	}

	var state struct {
		Todos []HistoricalTaskData `json:"todos"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Todo2 JSON: %w", err)
	}

	historical := make([]HistoricalTask, 0)

	for _, task := range state.Todos {
		status := normalizeStatus(task.Status)
		if status != "Done" {
			continue
		}

		actualHours := task.ActualHours
		if actualHours == 0 {
			completedAt := task.CompletedAt
			if completedAt == "" {
				completedAt = task.LastModified
			}

			if task.Created != "" && completedAt != "" {
				if createdTime, err := time.Parse(time.RFC3339, task.Created); err == nil {
					if completedTime, err := time.Parse(time.RFC3339, completedAt); err == nil {
						duration := completedTime.Sub(createdTime)
						calendarHours := duration.Hours()
						days := calendarHours / 24.0
						estimatedWorkHours := math.Min(calendarHours, days*8.0)
						estimatedWorkHours = math.Min(16.0, estimatedWorkHours)
						actualHours = math.Max(0.1, estimatedWorkHours)
					}
				}
			}
		}

		if actualHours > 0 {
			name := task.Name
			if name == "" {
				name = task.Content
			}

			details := task.LongDescription
			if details == "" {
				details = task.Details
			}

			historical = append(historical, HistoricalTask{
				Name:           name,
				Details:        details,
				Tags:           task.Tags,
				Priority:       task.Priority,
				EstimatedHours: task.EstimatedHours,
				ActualHours:    actualHours,
				Created:        task.Created,
				CompletedAt:    task.CompletedAt,
			})
		}
	}

	return historical, nil
}

// estimateFromHistory estimates using historical data matching.
func estimateFromHistory(text string, tags []string, priority string, historical []HistoricalTask) (*float64, float64) {
	type match struct {
		actualHours float64
		score       float64
	}

	matches := make([]match, 0)

	textWords := make(map[string]bool)
	for _, word := range strings.Fields(text) {
		textWords[word] = true
	}

	for _, record := range historical {
		score := 0.0
		recordText := strings.ToLower(record.Name + " " + record.Details)
		recordWords := make(map[string]bool)

		for _, word := range strings.Fields(recordText) {
			recordWords[word] = true
		}

		// Text similarity (word overlap)
		if len(textWords) > 0 && len(recordWords) > 0 {
			intersection := 0
			union := len(textWords)

			for word := range recordWords {
				if textWords[word] {
					intersection++
				} else {
					union++
				}
			}

			if union > 0 {
				wordOverlap := float64(intersection) / float64(union)
				score += wordOverlap * 0.5
			}
		}

		// Tag matching
		if len(tags) > 0 && len(record.Tags) > 0 {
			tagOverlap := 0

			tagSet := make(map[string]bool)
			for _, tag := range tags {
				tagSet[strings.ToLower(tag)] = true
			}

			for _, tag := range record.Tags {
				if tagSet[strings.ToLower(tag)] {
					tagOverlap++
				}
			}

			if len(record.Tags) > 0 {
				score += float64(tagOverlap) / float64(len(record.Tags)) * 0.3
			}
		}

		// Priority matching
		if strings.EqualFold(priority, record.Priority) {
			score += 0.2
		}

		if score > 0.1 {
			matches = append(matches, match{
				actualHours: record.ActualHours,
				score:       score,
			})
		}
	}

	if len(matches) == 0 {
		return nil, 0.0
	}

	// Sort by score and take top matches
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	topN := 10
	if len(matches) < topN {
		topN = len(matches)
	}

	topMatches := matches[:topN]

	// Weighted average
	totalWeight := 0.0
	weightedSum := 0.0

	for _, m := range topMatches {
		totalWeight += m.score
		weightedSum += m.actualHours * m.score
	}

	if totalWeight == 0 {
		return nil, 0.0
	}

	estimate := weightedSum / totalWeight

	// Confidence based on match quality
	avgScore := totalWeight / float64(len(topMatches))
	confidence := math.Min(0.9, 0.3+avgScore*0.6)

	return &estimate, confidence
}

// estimateFromKeywords estimates using keyword heuristics
// These are conservative estimates based on typical task durations.
func estimateFromKeywords(text string) float64 {
	// Very quick tasks (typos, version bumps, simple configs)
	quickKeywords := []string{"quick", "simple", "minor", "small", "fix typo", "typo", "update version", "bump", "version", "config", "spelling"}
	for _, kw := range quickKeywords {
		if strings.Contains(text, kw) {
			return 0.25 // 15 minutes
		}
	}

	// Small tasks (single feature, simple addition)
	smallKeywords := []string{"add", "create", "implement", "setup", "install", "configure", "fix", "bug", "patch"}
	for _, kw := range smallKeywords {
		if strings.Contains(text, kw) {
			return 1.0 // 1 hour
		}
	}

	// Medium tasks (refactoring, integration, updates)
	mediumKeywords := []string{"refactor", "migrate", "integrate", "update", "improve", "enhance", "modify", "change"}
	for _, kw := range mediumKeywords {
		if strings.Contains(text, kw) {
			return 2.0 // 2 hours
		}
	}

	// Large tasks (complex features, major changes)
	largeKeywords := []string{"complex", "major", "rewrite", "redesign", "architecture", "system", "framework", "platform"}
	for _, kw := range largeKeywords {
		if strings.Contains(text, kw) {
			return 4.0 // 4 hours
		}
	}

	// Default: assume small task (1 hour)
	// Most tasks are small fixes or features
	return 1.0
}

// getPriorityMultiplier returns time multiplier based on priority.
func getPriorityMultiplier(priority string) float64 {
	priority = strings.ToLower(priority)
	multipliers := map[string]float64{
		"low":      0.8,
		"medium":   1.0,
		"high":     1.2,
		"critical": 1.5,
	}

	if mult, ok := multipliers[priority]; ok {
		return mult
	}

	return 1.0
}

// HistoricalTaskData represents task data from Todo2 for historical analysis.
type HistoricalTaskData struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name,omitempty"`
	Content         string   `json:"content,omitempty"`
	LongDescription string   `json:"long_description,omitempty"`
	Details         string   `json:"details,omitempty"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	EstimatedHours  float64  `json:"estimatedHours,omitempty"`
	ActualHours     float64  `json:"actualHours,omitempty"`
	Created         string   `json:"created,omitempty"`
	CompletedAt     string   `json:"completedAt,omitempty"`
	LastModified    string   `json:"lastModified,omitempty"`
}

// BuildEstimationPrompt returns the standard task estimation prompt for LLM backends (Ollama, MLX, etc.).
func BuildEstimationPrompt(name, details string, tags []string, priority string) string {
	tagsStr := "none"
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	}

	return fmt.Sprintf(`You are an expert software development task estimator. Respond ONLY with valid JSON.

Estimate the time required to complete this software development task.

TASK INFORMATION:
- Name: %s
- Details: %s
- Tags: %s
- Priority: %s

RESPONSE FORMAT (JSON only):
{
    "estimate_hours": <number between 0.5 and 20>,
    "confidence": <number between 0.0 and 1.0>,
    "complexity": <number between 1 and 10>,
    "reasoning": "<brief 1-2 sentence explanation>"
}`, name, details, tagsStr, priority)
}

// ParseLLMEstimationResponse extracts EstimationResult from LLM response text (Ollama/MLX/any JSON).
func ParseLLMEstimationResponse(text string) (*EstimationResult, error) {
	jsonStr := ExtractJSONObjectFromLLMResponse(text)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	var parsed struct {
		EstimateHours float64 `json:"estimate_hours"`
		Confidence    float64 `json:"confidence"`
		Complexity    float64 `json:"complexity"`
		Reasoning     string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if parsed.EstimateHours < 0.5 || parsed.EstimateHours > 20 {
		parsed.EstimateHours = math.Max(0.5, math.Min(20, parsed.EstimateHours))
	}

	if parsed.Confidence < 0 || parsed.Confidence > 1 {
		parsed.Confidence = math.Max(0, math.Min(1, parsed.Confidence))
	}

	stdDev := parsed.EstimateHours * 0.3
	lowerBound := math.Max(0.5, parsed.EstimateHours-1.96*stdDev)
	upperBound := parsed.EstimateHours + 1.96*stdDev

	return &EstimationResult{
		EstimateHours: math.Round(parsed.EstimateHours*10) / 10,
		Confidence:    math.Min(0.95, parsed.Confidence),
		Method:        "ollama",
		LowerBound:    math.Round(lowerBound*10) / 10,
		UpperBound:    math.Round(upperBound*10) / 10,
		Metadata: map[string]interface{}{
			"complexity": parsed.Complexity,
			"reasoning":  parsed.Reasoning,
		},
	}, nil
}

// EstimateWithOllama runs task estimation using the default Ollama provider.
// Model defaults to llama3.2; pass empty to use default.
func EstimateWithOllama(ctx context.Context, name, details string, tags []string, priority, model string) (*EstimationResult, error) {
	prompt := BuildEstimationPrompt(name, details, tags, priority)

	if model == "" {
		model = "llama3.2"
	}

	params := map[string]interface{}{
		"action": "generate",
		"prompt": prompt,
		"model":  model,
		"stream": false,
	}

	tc, err := DefaultOllama().Invoke(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("ollama invoke: %w", err)
	}

	var responseText string

	if len(tc) > 0 && tc[0].Text != "" {
		var genResp map[string]interface{}
		if err := json.Unmarshal([]byte(tc[0].Text), &genResp); err == nil {
			if resp, ok := genResp["response"].(string); ok {
				responseText = resp
			}
		}

		if responseText == "" {
			responseText = tc[0].Text
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("ollama returned empty response")
	}

	return ParseLLMEstimationResponse(responseText)
}
