// loader.go — Config loading from .exarp/config.pb (protobuf) with YAML fallback detection.
//
// Package config provides protobuf-based project configuration loading, validation, and defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/projectroot"
	configpb "github.com/davidl71/exarp-go/proto"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from .exarp/config.pb only (protobuf mandatory).
// If no config file exists, returns defaults. If only .exarp/config.yaml exists,
// returns an error instructing the user to run "exarp-go config init" or "exarp-go config convert".
func LoadConfig(projectRoot string) (*FullConfig, error) {
	pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
	yamlPath := filepath.Join(projectRoot, ".exarp", "config.yaml")

	if _, err := os.Stat(pbPath); err == nil {
		return LoadConfigProtobuf(projectRoot)
	}

	// Protobuf mandatory: if YAML exists but no .pb, error with instructions
	if _, err := os.Stat(yamlPath); err == nil {
		return nil, fmt.Errorf("configuration must be in protobuf format: run 'exarp-go config convert yaml protobuf' to create .exarp/config.pb, or 'exarp-go config init' for defaults")
	}

	// No config file: use defaults
	cfg := GetDefaults()
	applyEnvOverrides(cfg)

	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadConfigYAML loads configuration from .exarp/config.yaml only (for convert/import).
// Use LoadConfig for normal runtime loading (protobuf mandatory).
func LoadConfigYAML(projectRoot string) (*FullConfig, error) {
	cfg := GetDefaults()
	yamlPath := filepath.Join(projectRoot, ".exarp", "config.yaml")

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", yamlPath, err)
	}

	var fileConfig FullConfig
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", yamlPath, err)
	}

	cfg = mergeConfig(cfg, &fileConfig)
	applyEnvOverrides(cfg)

	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// WriteConfigToProtobufFile writes cfg to .exarp/config.pb (creates .exarp if needed).
func WriteConfigToProtobufFile(projectRoot string, cfg *FullConfig) error {
	pbConfig, err := ToProtobuf(cfg)
	if err != nil {
		return fmt.Errorf("convert to protobuf: %w", err)
	}

	data, err := proto.Marshal(pbConfig)
	if err != nil {
		return fmt.Errorf("marshal protobuf: %w", err)
	}

	pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
	if err := os.MkdirAll(filepath.Dir(pbPath), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(pbPath, data, 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", pbPath, err)
	}

	return nil
}

// LoadConfigProtobuf loads configuration from .exarp/config.pb (protobuf binary format).
func LoadConfigProtobuf(projectRoot string) (*FullConfig, error) {
	// Start with defaults
	cfg := GetDefaults()

	// Load protobuf config file
	configPath := filepath.Join(projectRoot, ".exarp", "config.pb")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read protobuf config file %s: %w", configPath, err)
	}

	// Unmarshal protobuf
	var pbConfig configpb.FullConfig
	if err := proto.Unmarshal(data, &pbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf config file %s: %w", configPath, err)
	}

	// Convert protobuf to Go structs
	fileConfig, err := FromProtobuf(&pbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf config: %w", err)
	}

	// Merge file config with defaults (file config takes precedence)
	cfg = mergeConfig(cfg, fileConfig)

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// GetConfigFormat returns the format of the config file (yaml, protobuf, or none).
func GetConfigFormat(projectRoot string) (string, error) {
	pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
	yamlPath := filepath.Join(projectRoot, ".exarp", "config.yaml")

	if _, err := os.Stat(pbPath); err == nil {
		return "protobuf", nil
	}

	if _, err := os.Stat(yamlPath); err == nil {
		return "yaml", nil
	}

	return "none", nil
}

// detectConfigFormat detects the format of a config file based on its extension.
func detectConfigFormat(configPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		return "yaml", nil
	case ".pb":
		return "protobuf", nil
	default:
		return "", fmt.Errorf("unknown config file format: %s (expected .yaml, .yml, or .pb)", ext)
	}
}

// mergeConfig merges fileConfig into defaults, with fileConfig taking precedence.
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

	// Merge database
	if fileConfig.Database.SQLitePath != "" {
		merged.Database.SQLitePath = fileConfig.Database.SQLitePath
	}

	if fileConfig.Database.JSONFallbackPath != "" {
		merged.Database.JSONFallbackPath = fileConfig.Database.JSONFallbackPath
	}

	if fileConfig.Database.BackupPath != "" {
		merged.Database.BackupPath = fileConfig.Database.BackupPath
	}

	if fileConfig.Database.MaxConnections > 0 {
		merged.Database.MaxConnections = fileConfig.Database.MaxConnections
	}

	if fileConfig.Database.ConnectionTimeout > 0 {
		merged.Database.ConnectionTimeout = fileConfig.Database.ConnectionTimeout
	}

	if fileConfig.Database.QueryTimeout > 0 {
		merged.Database.QueryTimeout = fileConfig.Database.QueryTimeout
	}

	if fileConfig.Database.RetryAttempts > 0 {
		merged.Database.RetryAttempts = fileConfig.Database.RetryAttempts
	}

	if fileConfig.Database.RetryInitialDelay > 0 {
		merged.Database.RetryInitialDelay = fileConfig.Database.RetryInitialDelay
	}

	if fileConfig.Database.RetryMaxDelay > 0 {
		merged.Database.RetryMaxDelay = fileConfig.Database.RetryMaxDelay
	}

	if fileConfig.Database.RetryMultiplier > 0 {
		merged.Database.RetryMultiplier = fileConfig.Database.RetryMultiplier
	}

	merged.Database.AutoVacuum = fileConfig.Database.AutoVacuum
	merged.Database.WALMode = fileConfig.Database.WALMode

	if fileConfig.Database.CheckpointInterval > 0 {
		merged.Database.CheckpointInterval = fileConfig.Database.CheckpointInterval
	}

	if fileConfig.Database.BackupRetentionDays > 0 {
		merged.Database.BackupRetentionDays = fileConfig.Database.BackupRetentionDays
	}

	// Merge security
	merged.Security.RateLimit.Enabled = fileConfig.Security.RateLimit.Enabled
	if fileConfig.Security.RateLimit.RequestsPerWindow > 0 {
		merged.Security.RateLimit.RequestsPerWindow = fileConfig.Security.RateLimit.RequestsPerWindow
	}

	if fileConfig.Security.RateLimit.WindowDuration > 0 {
		merged.Security.RateLimit.WindowDuration = fileConfig.Security.RateLimit.WindowDuration
	}

	if fileConfig.Security.RateLimit.BurstSize > 0 {
		merged.Security.RateLimit.BurstSize = fileConfig.Security.RateLimit.BurstSize
	}

	merged.Security.PathValidation.Enabled = fileConfig.Security.PathValidation.Enabled
	merged.Security.PathValidation.AllowAbsolutePaths = fileConfig.Security.PathValidation.AllowAbsolutePaths

	if fileConfig.Security.PathValidation.MaxDepth > 0 {
		merged.Security.PathValidation.MaxDepth = fileConfig.Security.PathValidation.MaxDepth
	}

	if len(fileConfig.Security.PathValidation.BlockedPatterns) > 0 {
		merged.Security.PathValidation.BlockedPatterns = fileConfig.Security.PathValidation.BlockedPatterns
	}

	if fileConfig.Security.FileLimits.MaxFileSize > 0 {
		merged.Security.FileLimits.MaxFileSize = fileConfig.Security.FileLimits.MaxFileSize
	}

	if fileConfig.Security.FileLimits.MaxFilesPerOperation > 0 {
		merged.Security.FileLimits.MaxFilesPerOperation = fileConfig.Security.FileLimits.MaxFilesPerOperation
	}

	if len(fileConfig.Security.FileLimits.AllowedExtensions) > 0 {
		merged.Security.FileLimits.AllowedExtensions = fileConfig.Security.FileLimits.AllowedExtensions
	}

	merged.Security.AccessControl.Enabled = fileConfig.Security.AccessControl.Enabled
	if fileConfig.Security.AccessControl.DefaultPolicy != "" {
		merged.Security.AccessControl.DefaultPolicy = fileConfig.Security.AccessControl.DefaultPolicy
	}

	if len(fileConfig.Security.AccessControl.RestrictedTools) > 0 {
		merged.Security.AccessControl.RestrictedTools = fileConfig.Security.AccessControl.RestrictedTools
	}

	// Logging
	if fileConfig.Logging.Level != "" {
		merged.Logging.Level = fileConfig.Logging.Level
	}
	if fileConfig.Logging.Format != "" {
		merged.Logging.Format = fileConfig.Logging.Format
	}
	if fileConfig.Logging.LogDir != "" {
		merged.Logging.LogDir = fileConfig.Logging.LogDir
	}
	if fileConfig.Logging.LogFile != "" {
		merged.Logging.LogFile = fileConfig.Logging.LogFile
	}
	merged.Logging.IncludeTimestamps = fileConfig.Logging.IncludeTimestamps
	merged.Logging.IncludeCaller = fileConfig.Logging.IncludeCaller
	merged.Logging.ColorOutput = fileConfig.Logging.ColorOutput
	if fileConfig.Logging.RetentionDays > 0 {
		merged.Logging.RetentionDays = fileConfig.Logging.RetentionDays
	}

	// Tools
	if fileConfig.Tools.Scorecard.OutputFormat != "" {
		merged.Tools.Scorecard.OutputFormat = fileConfig.Tools.Scorecard.OutputFormat
	}
	merged.Tools.Scorecard.IncludeWisdom = fileConfig.Tools.Scorecard.IncludeWisdom
	if fileConfig.Tools.Linting.DefaultLinter != "" {
		merged.Tools.Linting.DefaultLinter = fileConfig.Tools.Linting.DefaultLinter
	}
	if fileConfig.Tools.Linting.Timeout > 0 {
		merged.Tools.Linting.Timeout = fileConfig.Tools.Linting.Timeout
	}
	merged.Tools.Linting.AutoFix = fileConfig.Tools.Linting.AutoFix
	if fileConfig.Tools.Ollama.DefaultHost != "" {
		merged.Tools.Ollama.DefaultHost = fileConfig.Tools.Ollama.DefaultHost
	}
	if fileConfig.Tools.Ollama.DefaultModel != "" {
		merged.Tools.Ollama.DefaultModel = fileConfig.Tools.Ollama.DefaultModel
	}
	if fileConfig.Tools.MLX.DefaultModel != "" {
		merged.Tools.MLX.DefaultModel = fileConfig.Tools.MLX.DefaultModel
	}
	if fileConfig.Tools.MLX.DefaultMaxTokens > 0 {
		merged.Tools.MLX.DefaultMaxTokens = fileConfig.Tools.MLX.DefaultMaxTokens
	}

	// Workflow
	if fileConfig.Workflow.DefaultMode != "" {
		merged.Workflow.DefaultMode = fileConfig.Workflow.DefaultMode
	}
	merged.Workflow.AutoDetectMode = fileConfig.Workflow.AutoDetectMode
	if len(fileConfig.Workflow.Modes) > 0 {
		merged.Workflow.Modes = fileConfig.Workflow.Modes
	}

	// Memory
	if fileConfig.Memory.StoragePath != "" {
		merged.Memory.StoragePath = fileConfig.Memory.StoragePath
	}
	if fileConfig.Memory.RetentionDays > 0 {
		merged.Memory.RetentionDays = fileConfig.Memory.RetentionDays
	}
	if fileConfig.Memory.MaxMemories > 0 {
		merged.Memory.MaxMemories = fileConfig.Memory.MaxMemories
	}
	merged.Memory.AutoCleanup = fileConfig.Memory.AutoCleanup

	// Project
	if fileConfig.Project.Name != "" {
		merged.Project.Name = fileConfig.Project.Name
	}
	if fileConfig.Project.Type != "" {
		merged.Project.Type = fileConfig.Project.Type
	}
	if fileConfig.Project.Language != "" {
		merged.Project.Language = fileConfig.Project.Language
	}
	if fileConfig.Project.Root != "" {
		merged.Project.Root = fileConfig.Project.Root
	}

	return &merged
}

// applyEnvOverrides applies environment variable overrides to config.
func applyEnvOverrides(cfg *FullConfig) {
	// Example: EXARP_TIMEOUT_TOOL_DEFAULT=120s
	// This will be expanded in future phases
	// For now, we keep it simple and focus on file-based config
}

// FindProjectRoot finds the exarp project root. Delegates to projectroot.Find().
// Returns current directory on error (config needs a path; tools.FindProjectRoot returns error).
func FindProjectRoot() (string, error) {
	return projectroot.Find()
}
