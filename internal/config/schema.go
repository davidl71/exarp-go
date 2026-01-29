package config

import (
	"time"
)

// FullConfig represents the complete configuration structure (for future use)
type FullConfig struct {
	Version     string            `yaml:"version"`
	Timeouts    TimeoutsConfig    `yaml:"timeouts"`
	Thresholds  ThresholdsConfig  `yaml:"thresholds"`
	Tasks       TasksConfig       `yaml:"tasks"`
	Database    DatabaseConfig    `yaml:"database"`
	Security    SecurityConfig    `yaml:"security"`
	Logging     LoggingConfig     `yaml:"logging"`
	Tools       ToolsConfig       `yaml:"tools"`
	Workflow    WorkflowConfig    `yaml:"workflow"`
	Memory      MemoryConfig      `yaml:"memory"`
	Project     ProjectConfig     `yaml:"project"`
	Automations AutomationsConfig `yaml:"automations"`
}

// TimeoutsConfig contains all timeout and duration settings
type TimeoutsConfig struct {
	// Task management
	TaskLockLease      time.Duration `yaml:"task_lock_lease"`
	TaskLockRenewal    time.Duration `yaml:"task_lock_renewal"`
	StaleLockThreshold time.Duration `yaml:"stale_lock_threshold"`

	// Tool execution
	ToolDefault   time.Duration `yaml:"tool_default"`
	ToolScorecard time.Duration `yaml:"tool_scorecard"`
	ToolLinting   time.Duration `yaml:"tool_linting"`
	ToolTesting   time.Duration `yaml:"tool_testing"`
	ToolReport    time.Duration `yaml:"tool_report"`

	// External services
	OllamaDownload time.Duration `yaml:"ollama_download"`
	OllamaGenerate time.Duration `yaml:"ollama_generate"`
	HTTPClient     time.Duration `yaml:"http_client"`
	DatabaseRetry  time.Duration `yaml:"database_retry"`

	// Context management
	ContextSummarize time.Duration `yaml:"context_summarize"`
	ContextBudget    time.Duration `yaml:"context_budget"`
}

// ThresholdsConfig contains all threshold and limit settings
type ThresholdsConfig struct {
	// Task analysis
	SimilarityThreshold  float64 `yaml:"similarity_threshold"`
	MinDescriptionLength int     `yaml:"min_description_length"`
	MinTaskConfidence    float64 `yaml:"min_task_confidence"`

	// Testing
	MinCoverage       int     `yaml:"min_coverage"`
	MinTestConfidence float64 `yaml:"min_test_confidence"`

	// Estimation
	MinEstimationConfidence float64 `yaml:"min_estimation_confidence"`
	MLXWeight               float64 `yaml:"mlx_weight"`

	// Automation
	MaxParallelTasks        int `yaml:"max_parallel_tasks"`
	MaxTasksPerHost         int `yaml:"max_tasks_per_host"`
	MaxAutomationIterations int `yaml:"max_automation_iterations"`

	// Context management
	TokensPerChar             float64 `yaml:"tokens_per_char"`
	DefaultContextBudget      int     `yaml:"default_context_budget"`
	ContextReductionThreshold float64 `yaml:"context_reduction_threshold"`

	// Security
	RateLimitRequests int           `yaml:"rate_limit_requests"`
	RateLimitWindow   time.Duration `yaml:"rate_limit_window"`
	MaxFileSize       int64         `yaml:"max_file_size"`
	MaxPathDepth      int           `yaml:"max_path_depth"`
}

// TasksConfig contains task management defaults and settings
type TasksConfig struct {
	// Defaults
	DefaultStatus   string   `yaml:"default_status"`
	DefaultPriority string   `yaml:"default_priority"`
	DefaultTags     []string `yaml:"default_tags"`

	// Status transitions (map of status -> allowed next statuses)
	StatusWorkflow map[string][]string `yaml:"status_workflow"`

	// Cleanup
	StaleThresholdHours int  `yaml:"stale_threshold_hours"`
	AutoCleanupEnabled  bool `yaml:"auto_cleanup_enabled"`
	CleanupDryRun       bool `yaml:"cleanup_dry_run"`

	// Task ID
	IDFormat string `yaml:"id_format"`
	IDPrefix string `yaml:"id_prefix"`

	// Clarity checks
	MinDescriptionLength int  `yaml:"min_description_length"`
	RequireDescription   bool `yaml:"require_description"`
	AutoClarify          bool `yaml:"auto_clarify"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	SQLitePath          string        `yaml:"sqlite_path"`
	JSONFallbackPath    string        `yaml:"json_fallback_path"`
	BackupPath          string        `yaml:"backup_path"`
	MaxConnections      int           `yaml:"max_connections"`
	ConnectionTimeout   time.Duration `yaml:"connection_timeout"`
	QueryTimeout        time.Duration `yaml:"query_timeout"`
	RetryAttempts       int           `yaml:"retry_attempts"`
	RetryInitialDelay   time.Duration `yaml:"retry_initial_delay"`
	RetryMaxDelay       time.Duration `yaml:"retry_max_delay"`
	RetryMultiplier     float64       `yaml:"retry_multiplier"`
	AutoVacuum          bool          `yaml:"auto_vacuum"`
	WALMode             bool          `yaml:"wal_mode"`
	CheckpointInterval  int           `yaml:"checkpoint_interval"`
	BackupRetentionDays int           `yaml:"backup_retention_days"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	PathValidation PathValidationConfig `yaml:"path_validation"`
	FileLimits     FileLimitsConfig     `yaml:"file_limits"`
	AccessControl  AccessControlConfig  `yaml:"access_control"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled           bool          `yaml:"enabled"`
	RequestsPerWindow int           `yaml:"requests_per_window"`
	WindowDuration    time.Duration `yaml:"window_duration"`
	BurstSize         int           `yaml:"burst_size"`
}

// PathValidationConfig contains path validation settings
type PathValidationConfig struct {
	Enabled            bool     `yaml:"enabled"`
	AllowAbsolutePaths bool     `yaml:"allow_absolute_paths"`
	MaxDepth           int      `yaml:"max_depth"`
	BlockedPatterns    []string `yaml:"blocked_patterns"`
}

// FileLimitsConfig contains file operation limits
type FileLimitsConfig struct {
	MaxFileSize          int64    `yaml:"max_file_size"`
	MaxFilesPerOperation int      `yaml:"max_files_per_operation"`
	AllowedExtensions    []string `yaml:"allowed_extensions"`
}

// AccessControlConfig contains access control settings
type AccessControlConfig struct {
	Enabled         bool     `yaml:"enabled"`
	DefaultPolicy   string   `yaml:"default_policy"` // "allow" or "deny"
	RestrictedTools []string `yaml:"restricted_tools"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level             string            `yaml:"level"`
	ToolLevel         string            `yaml:"tool_level"`
	FrameworkLevel    string            `yaml:"framework_level"`
	Format            string            `yaml:"format"`
	IncludeTimestamps bool              `yaml:"include_timestamps"`
	IncludeCaller     bool              `yaml:"include_caller"`
	ColorOutput       bool              `yaml:"color_output"`
	LogDir            string            `yaml:"log_dir"`
	LogFile           string            `yaml:"log_file"`
	SessionLogDir     string            `yaml:"session_log_dir"`
	LogRotation       LogRotationConfig `yaml:"log_rotation"`
	RetentionDays     int               `yaml:"retention_days"`
	AutoCleanup       bool              `yaml:"auto_cleanup"`
}

// LogRotationConfig contains log rotation settings
type LogRotationConfig struct {
	Enabled  bool  `yaml:"enabled"`
	MaxSize  int64 `yaml:"max_size"`
	MaxFiles int   `yaml:"max_files"`
	Compress bool  `yaml:"compress"`
}

// ToolsConfig contains tool-specific settings
type ToolsConfig struct {
	Scorecard ScorecardConfig `yaml:"scorecard"`
	Report    ReportConfig    `yaml:"report"`
	Linting   LintingConfig   `yaml:"linting"`
	Testing   TestingConfig   `yaml:"testing"`
	MLX       MLXConfig       `yaml:"mlx"`
	Ollama    OllamaConfig    `yaml:"ollama"`
	Context   ContextConfig   `yaml:"context"`
}

// ScorecardConfig contains scorecard tool settings
type ScorecardConfig struct {
	DefaultScores map[string]float64 `yaml:"default_scores"`
	IncludeWisdom bool               `yaml:"include_wisdom"`
	OutputFormat  string             `yaml:"output_format"`
}

// ReportConfig contains report tool settings
type ReportConfig struct {
	DefaultFormat          string `yaml:"default_format"`
	DefaultOutputPath      string `yaml:"default_output_path"`
	IncludeMetrics         bool   `yaml:"include_metrics"`
	IncludeRecommendations bool   `yaml:"include_recommendations"`
}

// LintingConfig contains linting tool settings
type LintingConfig struct {
	DefaultLinter string        `yaml:"default_linter"`
	AutoFix       bool          `yaml:"auto_fix"`
	IncludeHints  bool          `yaml:"include_hints"`
	Timeout       time.Duration `yaml:"timeout"`
}

// TestingConfig contains testing tool settings
type TestingConfig struct {
	DefaultFramework string `yaml:"default_framework"`
	MinCoverage      int    `yaml:"min_coverage"`
	CoverageFormat   string `yaml:"coverage_format"`
	Verbose          bool   `yaml:"verbose"`
}

// MLXConfig contains MLX tool settings
type MLXConfig struct {
	DefaultModel       string  `yaml:"default_model"`
	DefaultMaxTokens   int     `yaml:"default_max_tokens"`
	DefaultTemperature float64 `yaml:"default_temperature"`
	Verbose            bool    `yaml:"verbose"`
}

// OllamaConfig contains Ollama tool settings
type OllamaConfig struct {
	DefaultModel       string `yaml:"default_model"`
	DefaultHost        string `yaml:"default_host"`
	DefaultContextSize int    `yaml:"default_context_size"`
	DefaultNumThreads  int    `yaml:"default_num_threads"`
	DefaultNumGPU      int    `yaml:"default_num_gpu"`
}

// ContextConfig contains context tool settings
type ContextConfig struct {
	DefaultBudget             int     `yaml:"default_budget"`
	DefaultSummarizationLevel string  `yaml:"default_summarization_level"`
	TokensPerChar             float64 `yaml:"tokens_per_char"`
	IncludeRaw                bool    `yaml:"include_raw"`
}

// WorkflowConfig contains workflow mode settings
type WorkflowConfig struct {
	DefaultMode     string                `yaml:"default_mode"`
	AutoDetectMode  bool                  `yaml:"auto_detect_mode"`
	ModeSuggestions ModeSuggestionsConfig `yaml:"mode_suggestions"`
	Modes           map[string]ModeConfig `yaml:"modes"`
	Focus           FocusConfig           `yaml:"focus"`
}

// ModeSuggestionsConfig contains time-based mode suggestions
type ModeSuggestionsConfig struct {
	Morning   string `yaml:"morning"`
	Afternoon string `yaml:"afternoon"`
	Evening   string `yaml:"evening"`
}

// ModeConfig contains settings for a specific workflow mode
type ModeConfig struct {
	EnabledTools  []string `yaml:"enabled_tools"`
	DisabledTools []string `yaml:"disabled_tools"`
	ToolLimit     int      `yaml:"tool_limit"`
}

// FocusConfig contains focus mode settings
type FocusConfig struct {
	Enabled           bool `yaml:"enabled"`
	ReductionTarget   int  `yaml:"reduction_target"`
	PreserveCoreTools bool `yaml:"preserve_core_tools"`
}

// MemoryConfig contains memory and session settings
type MemoryConfig struct {
	Categories     []string            `yaml:"categories"`
	StoragePath    string              `yaml:"storage_path"`
	SessionLogPath string              `yaml:"session_log_path"`
	RetentionDays  int                 `yaml:"retention_days"`
	AutoCleanup    bool                `yaml:"auto_cleanup"`
	MaxMemories    int                 `yaml:"max_memories"`
	Consolidation  ConsolidationConfig `yaml:"consolidation"`
}

// ConsolidationConfig contains memory consolidation settings
type ConsolidationConfig struct {
	Enabled             bool    `yaml:"enabled"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
	Frequency           string  `yaml:"frequency"`
}

// SessionConfig contains session settings
type SessionConfig struct {
	AutoPrime            bool   `yaml:"auto_prime"`
	IncludeHints         bool   `yaml:"include_hints"`
	IncludeTasks         bool   `yaml:"include_tasks"`
	TaskLimit            int    `yaml:"task_limit"`
	EnableHandoffs       bool   `yaml:"enable_handoffs"`
	HandoffRetentionDays int    `yaml:"handoff_retention_days"`
	TrackPrompts         bool   `yaml:"track_prompts"`
	PromptLogPath        string `yaml:"prompt_log_path"`
}

// ProjectConfig contains project-specific settings
type ProjectConfig struct {
	Name        string         `yaml:"name"`
	Type        string         `yaml:"type"`
	Language    string         `yaml:"language"`
	Root        string         `yaml:"root"`
	Todo2Path   string         `yaml:"todo2_path"`
	ExarpPath   string         `yaml:"exarp_path"`
	Features    FeaturesConfig `yaml:"features"`
	SkipChecks  []string       `yaml:"skip_checks"`
	CustomTools []string       `yaml:"custom_tools"`
}

// FeaturesConfig contains feature flags
type FeaturesConfig struct {
	SQLiteEnabled bool     `yaml:"sqlite_enabled"`
	JSONFallback  bool     `yaml:"json_fallback"`
	PythonBridge  bool     `yaml:"python_bridge"`
	MCPServers    []string `yaml:"mcp_servers"`
}

// AutomationsConfig contains automation workflow configurations
// This will be expanded in Phase 5
type AutomationsConfig struct {
	// Placeholder for now - will be implemented in Phase 5
	// See AUTOMATION_CONFIGURATION_ANALYSIS.md for full structure
}
