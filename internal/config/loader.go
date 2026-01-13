package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from .exarp/config.yaml with fallback to defaults
func LoadConfig(projectRoot string) (*FullConfig, error) {
	// Start with defaults
	cfg := GetDefaults()

	// Try to load config file
	configPath := filepath.Join(projectRoot, ".exarp", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		// Config file exists, load it
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}

		// Parse YAML
		var fileConfig FullConfig
		if err := yaml.Unmarshal(data, &fileConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
		}

		// Merge file config with defaults (file config takes precedence)
		cfg = mergeConfig(cfg, &fileConfig)
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// mergeConfig merges fileConfig into defaults, with fileConfig taking precedence
func mergeConfig(defaults, fileConfig *FullConfig) *FullConfig {
	merged := *defaults

	// Merge each section if present in file config
	if fileConfig.Version != "" {
		merged.Version = fileConfig.Version
	}

	// Merge timeouts
	if fileConfig.Timeouts.TaskLockLease > 0 {
		merged.Timeouts.TaskLockLease = fileConfig.Timeouts.TaskLockLease
	}
	if fileConfig.Timeouts.TaskLockRenewal > 0 {
		merged.Timeouts.TaskLockRenewal = fileConfig.Timeouts.TaskLockRenewal
	}
	if fileConfig.Timeouts.StaleLockThreshold > 0 {
		merged.Timeouts.StaleLockThreshold = fileConfig.Timeouts.StaleLockThreshold
	}
	if fileConfig.Timeouts.ToolDefault > 0 {
		merged.Timeouts.ToolDefault = fileConfig.Timeouts.ToolDefault
	}
	if fileConfig.Timeouts.ToolScorecard > 0 {
		merged.Timeouts.ToolScorecard = fileConfig.Timeouts.ToolScorecard
	}
	if fileConfig.Timeouts.ToolLinting > 0 {
		merged.Timeouts.ToolLinting = fileConfig.Timeouts.ToolLinting
	}
	if fileConfig.Timeouts.ToolTesting > 0 {
		merged.Timeouts.ToolTesting = fileConfig.Timeouts.ToolTesting
	}
	if fileConfig.Timeouts.ToolReport > 0 {
		merged.Timeouts.ToolReport = fileConfig.Timeouts.ToolReport
	}
	if fileConfig.Timeouts.OllamaDownload > 0 {
		merged.Timeouts.OllamaDownload = fileConfig.Timeouts.OllamaDownload
	}
	if fileConfig.Timeouts.OllamaGenerate > 0 {
		merged.Timeouts.OllamaGenerate = fileConfig.Timeouts.OllamaGenerate
	}
	if fileConfig.Timeouts.HTTPClient > 0 {
		merged.Timeouts.HTTPClient = fileConfig.Timeouts.HTTPClient
	}
	if fileConfig.Timeouts.DatabaseRetry > 0 {
		merged.Timeouts.DatabaseRetry = fileConfig.Timeouts.DatabaseRetry
	}
	if fileConfig.Timeouts.ContextSummarize > 0 {
		merged.Timeouts.ContextSummarize = fileConfig.Timeouts.ContextSummarize
	}
	if fileConfig.Timeouts.ContextBudget > 0 {
		merged.Timeouts.ContextBudget = fileConfig.Timeouts.ContextBudget
	}

	// Merge thresholds
	if fileConfig.Thresholds.SimilarityThreshold > 0 {
		merged.Thresholds.SimilarityThreshold = fileConfig.Thresholds.SimilarityThreshold
	}
	if fileConfig.Thresholds.MinDescriptionLength > 0 {
		merged.Thresholds.MinDescriptionLength = fileConfig.Thresholds.MinDescriptionLength
	}
	if fileConfig.Thresholds.MinTaskConfidence > 0 {
		merged.Thresholds.MinTaskConfidence = fileConfig.Thresholds.MinTaskConfidence
	}
	if fileConfig.Thresholds.MinCoverage > 0 {
		merged.Thresholds.MinCoverage = fileConfig.Thresholds.MinCoverage
	}
	if fileConfig.Thresholds.MinTestConfidence > 0 {
		merged.Thresholds.MinTestConfidence = fileConfig.Thresholds.MinTestConfidence
	}
	if fileConfig.Thresholds.MinEstimationConfidence > 0 {
		merged.Thresholds.MinEstimationConfidence = fileConfig.Thresholds.MinEstimationConfidence
	}
	if fileConfig.Thresholds.MLXWeight > 0 {
		merged.Thresholds.MLXWeight = fileConfig.Thresholds.MLXWeight
	}
	if fileConfig.Thresholds.MaxParallelTasks > 0 {
		merged.Thresholds.MaxParallelTasks = fileConfig.Thresholds.MaxParallelTasks
	}
	if fileConfig.Thresholds.MaxTasksPerHost > 0 {
		merged.Thresholds.MaxTasksPerHost = fileConfig.Thresholds.MaxTasksPerHost
	}
	if fileConfig.Thresholds.MaxAutomationIterations > 0 {
		merged.Thresholds.MaxAutomationIterations = fileConfig.Thresholds.MaxAutomationIterations
	}
	if fileConfig.Thresholds.TokensPerChar > 0 {
		merged.Thresholds.TokensPerChar = fileConfig.Thresholds.TokensPerChar
	}
	if fileConfig.Thresholds.DefaultContextBudget > 0 {
		merged.Thresholds.DefaultContextBudget = fileConfig.Thresholds.DefaultContextBudget
	}
	if fileConfig.Thresholds.ContextReductionThreshold > 0 {
		merged.Thresholds.ContextReductionThreshold = fileConfig.Thresholds.ContextReductionThreshold
	}
	if fileConfig.Thresholds.RateLimitRequests > 0 {
		merged.Thresholds.RateLimitRequests = fileConfig.Thresholds.RateLimitRequests
	}
	if fileConfig.Thresholds.RateLimitWindow > 0 {
		merged.Thresholds.RateLimitWindow = fileConfig.Thresholds.RateLimitWindow
	}
	if fileConfig.Thresholds.MaxFileSize > 0 {
		merged.Thresholds.MaxFileSize = fileConfig.Thresholds.MaxFileSize
	}
	if fileConfig.Thresholds.MaxPathDepth > 0 {
		merged.Thresholds.MaxPathDepth = fileConfig.Thresholds.MaxPathDepth
	}

	// Merge tasks
	if fileConfig.Tasks.DefaultStatus != "" {
		merged.Tasks.DefaultStatus = fileConfig.Tasks.DefaultStatus
	}
	if fileConfig.Tasks.DefaultPriority != "" {
		merged.Tasks.DefaultPriority = fileConfig.Tasks.DefaultPriority
	}
	if len(fileConfig.Tasks.DefaultTags) > 0 {
		merged.Tasks.DefaultTags = fileConfig.Tasks.DefaultTags
	}
	if len(fileConfig.Tasks.StatusWorkflow) > 0 {
		merged.Tasks.StatusWorkflow = fileConfig.Tasks.StatusWorkflow
	}
	if fileConfig.Tasks.StaleThresholdHours > 0 {
		merged.Tasks.StaleThresholdHours = fileConfig.Tasks.StaleThresholdHours
	}
	// Boolean fields
	merged.Tasks.AutoCleanupEnabled = fileConfig.Tasks.AutoCleanupEnabled
	merged.Tasks.CleanupDryRun = fileConfig.Tasks.CleanupDryRun
	if fileConfig.Tasks.IDFormat != "" {
		merged.Tasks.IDFormat = fileConfig.Tasks.IDFormat
	}
	if fileConfig.Tasks.IDPrefix != "" {
		merged.Tasks.IDPrefix = fileConfig.Tasks.IDPrefix
	}
	if fileConfig.Tasks.MinDescriptionLength > 0 {
		merged.Tasks.MinDescriptionLength = fileConfig.Tasks.MinDescriptionLength
	}
	merged.Tasks.RequireDescription = fileConfig.Tasks.RequireDescription
	merged.Tasks.AutoClarify = fileConfig.Tasks.AutoClarify

	// TODO: Merge remaining sections (database, security, logging, tools, workflow, memory, project, automations)
	// For Phase 1, we focus on timeouts, thresholds, and tasks

	return &merged
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(cfg *FullConfig) {
	// Example: EXARP_TIMEOUT_TOOL_DEFAULT=120s
	// This will be expanded in future phases
	// For now, we keep it simple and focus on file-based config
}

// FindProjectRoot finds the project root directory by looking for .exarp or .todo2
func FindProjectRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up the directory tree looking for .exarp or .todo2
	for {
		// Check for .exarp directory
		if _, err := os.Stat(filepath.Join(dir, ".exarp")); err == nil {
			return dir, nil
		}
		// Check for .todo2 directory (fallback)
		if _, err := os.Stat(filepath.Join(dir, ".todo2")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, stop
			break
		}
		dir = parent
	}

	// If not found, return current directory
	return os.Getwd()
}
