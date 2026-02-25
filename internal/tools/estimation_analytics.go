// estimation_analytics.go — Estimation analytics: analyze, stats, batch actions.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

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
			priority = models.PriorityMedium
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
	} else if statusFilter := strings.TrimSpace(cast.ToString(params["status_filter"])); statusFilter != "" {
		for _, t := range tasks {
			if t.Status == statusFilter {
				target = append(target, t)
			}
		}
	} else {
		// Default: all Todo
		for _, t := range tasks {
			if t.Status == models.StatusTodo {
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
	if _, ok := params["use_historical"]; ok {
		useHistorical = cast.ToBool(params["use_historical"])
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
			priority = models.PriorityMedium
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

