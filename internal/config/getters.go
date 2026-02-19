// getters.go — Global config access and field getter methods.
package config

import (
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

// HTTPClientTimeout returns the HTTP client timeout.
func HTTPClientTimeout() time.Duration {
	return GetGlobalConfig().Timeouts.HTTPClient
}

// DatabaseRetryTimeout returns the database retry timeout.
func DatabaseRetryTimeout() time.Duration {
	return GetGlobalConfig().Timeouts.DatabaseRetry
}

// Threshold getters - convenient access to threshold values

// SimilarityThreshold returns the similarity threshold for duplicate detection.
func SimilarityThreshold() float64 {
	return GetGlobalConfig().Thresholds.SimilarityThreshold
}

// MinDescriptionLength returns the minimum description length for tasks.
func MinDescriptionLength() int {
	return GetGlobalConfig().Thresholds.MinDescriptionLength
}

// MinTaskConfidence returns the minimum task confidence threshold.
func MinTaskConfidence() float64 {
	return GetGlobalConfig().Thresholds.MinTaskConfidence
}

// MinCoverage returns the minimum test coverage percentage.
func MinCoverage() int {
	return GetGlobalConfig().Thresholds.MinCoverage
}

// MinTestConfidence returns the minimum test confidence threshold.
func MinTestConfidence() float64 {
	return GetGlobalConfig().Thresholds.MinTestConfidence
}

// MinEstimationConfidence returns the minimum estimation confidence threshold.
func MinEstimationConfidence() float64 {
	return GetGlobalConfig().Thresholds.MinEstimationConfidence
}

// MLXWeight returns the MLX model weight for estimation.
func MLXWeight() float64 {
	return GetGlobalConfig().Thresholds.MLXWeight
}

// MaxParallelTasks returns the maximum number of parallel tasks.
func MaxParallelTasks() int {
	return GetGlobalConfig().Thresholds.MaxParallelTasks
}

// MaxTasksPerHost returns the maximum number of tasks per host.
func MaxTasksPerHost() int {
	return GetGlobalConfig().Thresholds.MaxTasksPerHost
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

// ContextReductionThreshold returns the context reduction threshold.
func ContextReductionThreshold() float64 {
	return GetGlobalConfig().Thresholds.ContextReductionThreshold
}

// RateLimitRequests returns the rate limit requests per window.
func RateLimitRequests() int {
	return GetGlobalConfig().Thresholds.RateLimitRequests
}

// RateLimitWindow returns the rate limit window duration.
func RateLimitWindow() time.Duration {
	return GetGlobalConfig().Thresholds.RateLimitWindow
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

// RequireTaskDescription returns whether task description is required.
func RequireTaskDescription() bool {
	return GetGlobalConfig().Tasks.RequireDescription
}

// AutoClarifyTasks returns whether to auto-request task clarification.
func AutoClarifyTasks() bool {
	return GetGlobalConfig().Tasks.AutoClarify
}

// Database getters - convenient access to database configuration

// GetDatabaseConfig returns the full database configuration.
func GetDatabaseConfig() DatabaseConfig {
	return GetGlobalConfig().Database
}

// SQLitePath returns the SQLite database file path.
func SQLitePath() string {
	return GetGlobalConfig().Database.SQLitePath
}

// JSONFallbackPath returns the JSON fallback file path.
func JSONFallbackPath() string {
	return GetGlobalConfig().Database.JSONFallbackPath
}

// DatabaseMaxConnections returns the maximum database connections.
func DatabaseMaxConnections() int {
	return GetGlobalConfig().Database.MaxConnections
}

// DatabaseConnectionTimeout returns the database connection timeout.
func DatabaseConnectionTimeout() time.Duration {
	return GetGlobalConfig().Database.ConnectionTimeout
}

// DatabaseQueryTimeout returns the database query timeout.
func DatabaseQueryTimeout() time.Duration {
	return GetGlobalConfig().Database.QueryTimeout
}

// DatabaseRetryAttempts returns the number of database retry attempts.
func DatabaseRetryAttempts() int {
	return GetGlobalConfig().Database.RetryAttempts
}

// DatabaseRetryInitialDelay returns the initial retry delay.
func DatabaseRetryInitialDelay() time.Duration {
	return GetGlobalConfig().Database.RetryInitialDelay
}

// DatabaseRetryMaxDelay returns the maximum retry delay.
func DatabaseRetryMaxDelay() time.Duration {
	return GetGlobalConfig().Database.RetryMaxDelay
}

// DatabaseRetryMultiplier returns the retry delay multiplier.
func DatabaseRetryMultiplier() float64 {
	return GetGlobalConfig().Database.RetryMultiplier
}

// Security getters - convenient access to security configuration

// GetSecurityConfig returns the full security configuration.
func GetSecurityConfig() SecurityConfig {
	return GetGlobalConfig().Security
}

// RateLimitEnabled returns whether rate limiting is enabled.
func RateLimitEnabled() bool {
	return GetGlobalConfig().Security.RateLimit.Enabled
}

// RateLimitRequestsPerWindow returns the number of requests allowed per window.
func RateLimitRequestsPerWindow() int {
	return GetGlobalConfig().Security.RateLimit.RequestsPerWindow
}

// RateLimitWindowDuration returns the rate limit window duration.
func RateLimitWindowDuration() time.Duration {
	return GetGlobalConfig().Security.RateLimit.WindowDuration
}

// RateLimitBurstSize returns the rate limit burst size.
func RateLimitBurstSize() int {
	return GetGlobalConfig().Security.RateLimit.BurstSize
}

// PathValidationEnabled returns whether path validation is enabled.
func PathValidationEnabled() bool {
	return GetGlobalConfig().Security.PathValidation.Enabled
}

// PathValidationAllowAbsolutePaths returns whether absolute paths are allowed.
func PathValidationAllowAbsolutePaths() bool {
	return GetGlobalConfig().Security.PathValidation.AllowAbsolutePaths
}

// PathValidationMaxDepth returns the maximum path depth.
func PathValidationMaxDepth() int {
	return GetGlobalConfig().Security.PathValidation.MaxDepth
}

// PathValidationBlockedPatterns returns the blocked path patterns.
func PathValidationBlockedPatterns() []string {
	return GetGlobalConfig().Security.PathValidation.BlockedPatterns
}

// FileLimitsMaxFileSize returns the maximum file size.
func FileLimitsMaxFileSize() int64 {
	return GetGlobalConfig().Security.FileLimits.MaxFileSize
}

// FileLimitsMaxFilesPerOperation returns the maximum files per operation.
func FileLimitsMaxFilesPerOperation() int {
	return GetGlobalConfig().Security.FileLimits.MaxFilesPerOperation
}

// FileLimitsAllowedExtensions returns the allowed file extensions.
func FileLimitsAllowedExtensions() []string {
	return GetGlobalConfig().Security.FileLimits.AllowedExtensions
}

// AccessControlEnabled returns whether access control is enabled.
func AccessControlEnabled() bool {
	return GetGlobalConfig().Security.AccessControl.Enabled
}

// AccessControlDefaultPolicy returns the default access control policy.
func AccessControlDefaultPolicy() string {
	return GetGlobalConfig().Security.AccessControl.DefaultPolicy
}

// AccessControlRestrictedTools returns the list of restricted tools.
func AccessControlRestrictedTools() []string {
	return GetGlobalConfig().Security.AccessControl.RestrictedTools
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
