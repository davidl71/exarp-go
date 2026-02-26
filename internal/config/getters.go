// getters.go — Global config access and field getter methods.
package config

import (
	"os"
	"sync"
	"time"
)

var (
	globalConfig *FullConfig
	configOnce   sync.Once
	configMu     sync.RWMutex
)

// SetGlobalConfig sets the global configuration instance
// This should be called once at application startup.
func SetGlobalConfig(cfg *FullConfig) {
	configMu.Lock()
	defer configMu.Unlock()

	globalConfig = cfg
}

// GetGlobalConfig returns the global configuration instance
// Returns defaults if not set.
func GetGlobalConfig() *FullConfig {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return GetDefaults()
	}

	return globalConfig
}

// Timeout getters - convenient access to timeout values

// TaskLockLease returns the task lock lease duration.
func TaskLockLease() time.Duration {
	return GetGlobalConfig().Timeouts.TaskLockLease
}

// ToolTimeout returns the default tool timeout, or a specific tool timeout if available.
func ToolTimeout(toolName string) time.Duration {
	cfg := GetGlobalConfig()

	switch toolName {
	case "scorecard":
		if cfg.Timeouts.ToolScorecard > 0 {
			return cfg.Timeouts.ToolScorecard
		}
	case "linting", "lint":
		if cfg.Timeouts.ToolLinting > 0 {
			return cfg.Timeouts.ToolLinting
		}
	case "testing", "test":
		if cfg.Timeouts.ToolTesting > 0 {
			return cfg.Timeouts.ToolTesting
		}
	case "report":
		if cfg.Timeouts.ToolReport > 0 {
			return cfg.Timeouts.ToolReport
		}
	}
	// Default timeout
	if cfg.Timeouts.ToolDefault > 0 {
		return cfg.Timeouts.ToolDefault
	}

	return 60 * time.Second // Fallback
}

// OllamaDownloadTimeout returns the Ollama download timeout.
func OllamaDownloadTimeout() time.Duration {
	return GetGlobalConfig().Timeouts.OllamaDownload
}

// OllamaGenerateTimeout returns the Ollama generation timeout.
func OllamaGenerateTimeout() time.Duration {
	return GetGlobalConfig().Timeouts.OllamaGenerate
}

// Threshold getters - convenient access to threshold values

// SimilarityThreshold returns the similarity threshold for duplicate detection.
func SimilarityThreshold() float64 {
	return GetGlobalConfig().Thresholds.SimilarityThreshold
}

// MinCoverage returns the minimum test coverage percentage.
func MinCoverage() int {
	return GetGlobalConfig().Thresholds.MinCoverage
}

// MaxTasksPerWave returns the maximum number of tasks per wave (0 = no limit).
func MaxTasksPerWave() int {
	return GetGlobalConfig().Thresholds.MaxTasksPerWave
}

// MaxAutomationIterations returns the maximum automation iterations.
func MaxAutomationIterations() int {
	return GetGlobalConfig().Thresholds.MaxAutomationIterations
}

// TokensPerChar returns the token estimation ratio.
func TokensPerChar() float64 {
	return GetGlobalConfig().Thresholds.TokensPerChar
}

// DefaultContextBudget returns the default context token budget.
func DefaultContextBudget() int {
	return GetGlobalConfig().Thresholds.DefaultContextBudget
}

// Task getters - convenient access to task configuration

// DefaultTaskStatus returns the default task status.
func DefaultTaskStatus() string {
	return GetGlobalConfig().Tasks.DefaultStatus
}

// DefaultTaskPriority returns the default task priority.
func DefaultTaskPriority() string {
	return GetGlobalConfig().Tasks.DefaultPriority
}

// DefaultTaskTags returns the default task tags.
func DefaultTaskTags() []string {
	return GetGlobalConfig().Tasks.DefaultTags
}

// StaleThresholdHours returns the stale task threshold in hours.
func StaleThresholdHours() int {
	return GetGlobalConfig().Tasks.StaleThresholdHours
}

// TaskMinDescriptionLength returns the minimum description length for tasks.
func TaskMinDescriptionLength() int {
	return GetGlobalConfig().Tasks.MinDescriptionLength
}

// Database getters - convenient access to database configuration

// DatabaseConnectionTimeout returns the database connection timeout.
func DatabaseConnectionTimeout() time.Duration {
	return GetGlobalConfig().Database.ConnectionTimeout
}

// DatabaseQueryTimeout returns the database query timeout.
func DatabaseQueryTimeout() time.Duration {
	return GetGlobalConfig().Database.QueryTimeout
}

// Project and Workflow getters - convenient access to project and workflow configuration

// ProjectExarpPath returns the exarp data directory path (e.g. ".exarp").
func ProjectExarpPath() string {
	return GetGlobalConfig().Project.ExarpPath
}

// WorkflowDefaultMode returns the default workflow mode.
func WorkflowDefaultMode() string {
	return GetGlobalConfig().Workflow.DefaultMode
}

// MemoryCategories returns the valid memory categories from config.
func MemoryCategories() []string {
	cfg := GetGlobalConfig()
	if len(cfg.Memory.Categories) > 0 {
		return cfg.Memory.Categories
	}

	return []string{"debug", "research", "architecture", "preference", "insight"}
}

// MemoryStoragePath returns the memory storage path (e.g. ".exarp/memories").
func MemoryStoragePath() string {
	path := GetGlobalConfig().Memory.StoragePath
	if path == "" {
		return ".exarp/memories"
	}

	return path
}

// GetOllamaDefaultModel returns the Ollama default model (env OLLAMA_DEFAULT_MODEL overrides config).
func GetOllamaDefaultModel() string {
	if s := os.Getenv("OLLAMA_DEFAULT_MODEL"); s != "" {
		return s
	}
	return GetGlobalConfig().Tools.Ollama.DefaultModel
}

// GetOllamaCodeModel returns the Ollama code model (env OLLAMA_CODE_MODEL overrides default).
func GetOllamaCodeModel() string {
	if s := os.Getenv("OLLAMA_CODE_MODEL"); s != "" {
		return s
	}
	return "codellama"
}
