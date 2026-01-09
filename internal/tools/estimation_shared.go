package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HistoricalTask represents a completed task for historical analysis
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

// EstimationResult represents the result of task duration estimation
type EstimationResult struct {
	EstimateHours float64                `json:"estimate_hours"`
	Confidence    float64                `json:"confidence"`
	Method        string                 `json:"method"`
	LowerBound    float64                `json:"lower_bound"`
	UpperBound    float64                `json:"upper_bound"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// estimateStatistically estimates task duration using statistical methods
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
func handleEstimationAnalyze(projectRoot string, params map[string]interface{}) (string, error) {
	// TODO: Implement analysis of estimation accuracy
	return `{"message": "Analysis action not yet implemented in native Go"}`, nil
}

// handleEstimationStats handles the stats action
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

// loadHistoricalTasks loads completed tasks from Todo2
func loadHistoricalTasks(projectRoot string) ([]HistoricalTask, error) {
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
		if status != "Done" && status != "Completed" {
			continue
		}

		// Get actual hours
		actualHours := task.ActualHours

		// If no actual hours, try to estimate from timestamps
		// NOTE: We can't accurately calculate work time from timestamps alone,
		// so we use a conservative estimate based on the time difference
		// but cap it at a reasonable maximum to avoid inflated estimates
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

						// Conservative estimate: assume at most 8 hours per day of calendar time
						// This prevents tasks that sat in backlog for weeks from inflating estimates
						days := calendarHours / 24.0
						estimatedWorkHours := math.Min(calendarHours, days*8.0)

						// Cap at reasonable maximum (16 hours) unless we have explicit actual hours
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

// estimateFromHistory estimates using historical data matching
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
		if strings.ToLower(priority) == strings.ToLower(record.Priority) {
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
// These are conservative estimates based on typical task durations
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

// getPriorityMultiplier returns time multiplier based on priority
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

// HistoricalTaskData represents task data from Todo2 for historical analysis
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
