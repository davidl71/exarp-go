// validation.go — Config validation rules for timeouts, thresholds, and sections.
package config

import (
	"fmt"
	"time"
)

// ValidateConfig validates the configuration and returns an error if invalid.
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

	// Validate database
	if err := validateDatabase(cfg.Database); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	// Validate security
	if err := validateSecurity(cfg.Security); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// TODO: Validate other sections in future phases (logging, tools, workflow, memory, project, automations)

	return nil
}

// validateTimeouts validates timeout configuration.
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

// validateThresholds validates threshold configuration.
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

// validateTasks validates task configuration.
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

// validateDatabase validates database configuration.
func validateDatabase(db DatabaseConfig) error {
	// SQLite path should not be empty if using SQLite
	if db.SQLitePath == "" {
		return fmt.Errorf("sqlite_path cannot be empty")
	}

	// Max connections should be positive
	if db.MaxConnections < 1 {
		return fmt.Errorf("max_connections (%d) must be at least 1", db.MaxConnections)
	}

	// Timeouts should be positive
	if db.ConnectionTimeout > 0 && db.ConnectionTimeout < time.Second {
		return fmt.Errorf("connection_timeout (%v) must be at least 1 second", db.ConnectionTimeout)
	}

	if db.QueryTimeout > 0 && db.QueryTimeout < time.Second {
		return fmt.Errorf("query_timeout (%v) must be at least 1 second", db.QueryTimeout)
	}

	// Retry attempts should be non-negative
	if db.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts (%d) must be non-negative", db.RetryAttempts)
	}

	// Retry delays should be positive if retry attempts > 0
	if db.RetryAttempts > 0 {
		if db.RetryInitialDelay <= 0 {
			return fmt.Errorf("retry_initial_delay must be positive when retry_attempts > 0")
		}

		if db.RetryMaxDelay <= 0 {
			return fmt.Errorf("retry_max_delay must be positive when retry_attempts > 0")
		}

		if db.RetryMaxDelay < db.RetryInitialDelay {
			return fmt.Errorf("retry_max_delay (%v) must be >= retry_initial_delay (%v)",
				db.RetryMaxDelay, db.RetryInitialDelay)
		}

		if db.RetryMultiplier <= 0 {
			return fmt.Errorf("retry_multiplier (%f) must be positive", db.RetryMultiplier)
		}
	}

	// Checkpoint interval should be positive
	if db.CheckpointInterval < 0 {
		return fmt.Errorf("checkpoint_interval (%d) must be non-negative", db.CheckpointInterval)
	}

	// Backup retention should be non-negative
	if db.BackupRetentionDays < 0 {
		return fmt.Errorf("backup_retention_days (%d) must be non-negative", db.BackupRetentionDays)
	}

	return nil
}

// validateSecurity validates security configuration.
func validateSecurity(sec SecurityConfig) error {
	// Validate rate limit
	if sec.RateLimit.Enabled {
		if sec.RateLimit.RequestsPerWindow < 1 {
			return fmt.Errorf("rate_limit.requests_per_window (%d) must be at least 1", sec.RateLimit.RequestsPerWindow)
		}

		if sec.RateLimit.WindowDuration <= 0 {
			return fmt.Errorf("rate_limit.window_duration (%v) must be positive", sec.RateLimit.WindowDuration)
		}

		if sec.RateLimit.BurstSize < 1 {
			return fmt.Errorf("rate_limit.burst_size (%d) must be at least 1", sec.RateLimit.BurstSize)
		}
	}

	// Validate path validation
	if sec.PathValidation.Enabled {
		if sec.PathValidation.MaxDepth < 1 {
			return fmt.Errorf("path_validation.max_depth (%d) must be at least 1", sec.PathValidation.MaxDepth)
		}
	}

	// Validate file limits
	if sec.FileLimits.MaxFileSize < 0 {
		return fmt.Errorf("file_limits.max_file_size (%d) must be non-negative", sec.FileLimits.MaxFileSize)
	}

	if sec.FileLimits.MaxFilesPerOperation < 1 {
		return fmt.Errorf("file_limits.max_files_per_operation (%d) must be at least 1", sec.FileLimits.MaxFilesPerOperation)
	}

	// Validate access control
	if sec.AccessControl.Enabled {
		validPolicies := map[string]bool{
			"allow": true,
			"deny":  true,
		}
		if sec.AccessControl.DefaultPolicy != "" && !validPolicies[sec.AccessControl.DefaultPolicy] {
			return fmt.Errorf("access_control.default_policy (%s) must be 'allow' or 'deny'", sec.AccessControl.DefaultPolicy)
		}
	}

	return nil
}
