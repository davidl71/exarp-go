package config

import (
	"fmt"
	"time"
)

// ValidateConfig validates the configuration and returns an error if invalid
func ValidateConfig(cfg *FullConfig) error {
	// Validate timeouts
	if err := validateTimeouts(cfg.Timeouts); err != nil {
		return fmt.Errorf("timeouts validation failed: %w", err)
	}

	// Validate thresholds
	if err := validateThresholds(cfg.Thresholds); err != nil {
		return fmt.Errorf("thresholds validation failed: %w", err)
	}

	// Validate tasks
	if err := validateTasks(cfg.Tasks); err != nil {
		return fmt.Errorf("tasks validation failed: %w", err)
	}

	// TODO: Validate other sections in future phases

	return nil
}

// validateTimeouts validates timeout configuration
func validateTimeouts(timeouts TimeoutsConfig) error {
	// Check for reasonable timeout values (not too small, not too large)
	minTimeout := 1 * time.Second
	maxTimeout := 24 * time.Hour

	timeoutChecks := []struct {
		name     string
		duration time.Duration
	}{
		{"task_lock_lease", timeouts.TaskLockLease},
		{"tool_default", timeouts.ToolDefault},
		{"tool_scorecard", timeouts.ToolScorecard},
		{"tool_linting", timeouts.ToolLinting},
		{"tool_testing", timeouts.ToolTesting},
		{"tool_report", timeouts.ToolReport},
		{"http_client", timeouts.HTTPClient},
		{"database_retry", timeouts.DatabaseRetry},
	}

	for _, check := range timeoutChecks {
		if check.duration > 0 && (check.duration < minTimeout || check.duration > maxTimeout) {
			return fmt.Errorf("timeout %s (%v) is out of valid range (%v - %v)",
				check.name, check.duration, minTimeout, maxTimeout)
		}
	}

	// Special checks for longer timeouts
	if timeouts.OllamaDownload > 0 && (timeouts.OllamaDownload < minTimeout || timeouts.OllamaDownload > 2*time.Hour) {
		return fmt.Errorf("ollama_download timeout (%v) is out of valid range (%v - %v)",
			timeouts.OllamaDownload, minTimeout, 2*time.Hour)
	}

	if timeouts.OllamaGenerate > 0 && (timeouts.OllamaGenerate < minTimeout || timeouts.OllamaGenerate > 1*time.Hour) {
		return fmt.Errorf("ollama_generate timeout (%v) is out of valid range (%v - %v)",
			timeouts.OllamaGenerate, minTimeout, 1*time.Hour)
	}

	return nil
}

// validateThresholds validates threshold configuration
func validateThresholds(thresholds ThresholdsConfig) error {
	// Similarity threshold should be between 0 and 1
	if thresholds.SimilarityThreshold < 0 || thresholds.SimilarityThreshold > 1 {
		return fmt.Errorf("similarity_threshold (%f) must be between 0 and 1", thresholds.SimilarityThreshold)
	}

	// Confidence thresholds should be between 0 and 1
	confidenceChecks := []struct {
		name  string
		value float64
	}{
		{"min_task_confidence", thresholds.MinTaskConfidence},
		{"min_test_confidence", thresholds.MinTestConfidence},
		{"min_estimation_confidence", thresholds.MinEstimationConfidence},
		{"mlx_weight", thresholds.MLXWeight},
		{"context_reduction_threshold", thresholds.ContextReductionThreshold},
	}

	for _, check := range confidenceChecks {
		if check.value < 0 || check.value > 1 {
			return fmt.Errorf("%s (%f) must be between 0 and 1", check.name, check.value)
		}
	}

	// Min description length should be positive
	if thresholds.MinDescriptionLength < 0 {
		return fmt.Errorf("min_description_length (%d) must be positive", thresholds.MinDescriptionLength)
	}

	// Coverage should be between 0 and 100
	if thresholds.MinCoverage < 0 || thresholds.MinCoverage > 100 {
		return fmt.Errorf("min_coverage (%d) must be between 0 and 100", thresholds.MinCoverage)
	}

	// Parallel tasks should be positive
	if thresholds.MaxParallelTasks < 1 {
		return fmt.Errorf("max_parallel_tasks (%d) must be at least 1", thresholds.MaxParallelTasks)
	}

	if thresholds.MaxTasksPerHost < 1 {
		return fmt.Errorf("max_tasks_per_host (%d) must be at least 1", thresholds.MaxTasksPerHost)
	}

	if thresholds.MaxAutomationIterations < 1 {
		return fmt.Errorf("max_automation_iterations (%d) must be at least 1", thresholds.MaxAutomationIterations)
	}

	// Token estimation should be positive
	if thresholds.TokensPerChar <= 0 {
		return fmt.Errorf("tokens_per_char (%f) must be positive", thresholds.TokensPerChar)
	}

	// Context budget should be positive
	if thresholds.DefaultContextBudget < 0 {
		return fmt.Errorf("default_context_budget (%d) must be positive", thresholds.DefaultContextBudget)
	}

	// Rate limit should be positive
	if thresholds.RateLimitRequests < 1 {
		return fmt.Errorf("rate_limit_requests (%d) must be at least 1", thresholds.RateLimitRequests)
	}

	if thresholds.RateLimitWindow <= 0 {
		return fmt.Errorf("rate_limit_window (%v) must be positive", thresholds.RateLimitWindow)
	}

	// File size should be positive
	if thresholds.MaxFileSize < 0 {
		return fmt.Errorf("max_file_size (%d) must be positive", thresholds.MaxFileSize)
	}

	// Path depth should be positive
	if thresholds.MaxPathDepth < 1 {
		return fmt.Errorf("max_path_depth (%d) must be at least 1", thresholds.MaxPathDepth)
	}

	return nil
}

// validateTasks validates task configuration
func validateTasks(tasks TasksConfig) error {
	// Valid statuses
	validStatuses := map[string]bool{
		"Todo":        true,
		"In Progress": true,
		"Review":      true,
		"Done":        true,
	}

	// Validate default status
	if tasks.DefaultStatus != "" && !validStatuses[tasks.DefaultStatus] {
		return fmt.Errorf("default_status (%s) is not a valid status", tasks.DefaultStatus)
	}

	// Validate default priority
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if tasks.DefaultPriority != "" && !validPriorities[tasks.DefaultPriority] {
		return fmt.Errorf("default_priority (%s) is not a valid priority (low, medium, high)", tasks.DefaultPriority)
	}

	// Validate status workflow
	for status, nextStatuses := range tasks.StatusWorkflow {
		if !validStatuses[status] {
			return fmt.Errorf("status_workflow contains invalid status: %s", status)
		}
		for _, nextStatus := range nextStatuses {
			if !validStatuses[nextStatus] {
				return fmt.Errorf("status_workflow for %s contains invalid next status: %s", status, nextStatus)
			}
		}
	}

	// Validate stale threshold
	if tasks.StaleThresholdHours < 0 {
		return fmt.Errorf("stale_threshold_hours (%d) must be non-negative", tasks.StaleThresholdHours)
	}

	// Validate min description length
	if tasks.MinDescriptionLength < 0 {
		return fmt.Errorf("min_description_length (%d) must be non-negative", tasks.MinDescriptionLength)
	}

	return nil
}
