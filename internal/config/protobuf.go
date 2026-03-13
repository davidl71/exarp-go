// protobuf.go — Protobuf serialization/deserialization for FullConfig.
// Section conversions live in protobuf_helpers.go and protobuf_<section>.go.
package config

import (
	"fmt"

	configpb "github.com/davidl71/exarp-go/proto"
)

// ToProtobuf converts Go FullConfig to protobuf FullConfig.
func ToProtobuf(cfg *FullConfig) (*configpb.FullConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	pb := &configpb.FullConfig{
		Version: cfg.Version,
	}

	if cfg.Timeouts.TaskLockLease > 0 || cfg.Timeouts.TaskLockRenewal > 0 || cfg.Timeouts.StaleLockThreshold > 0 ||
		cfg.Timeouts.ToolDefault > 0 || cfg.Timeouts.ToolScorecard > 0 || cfg.Timeouts.ToolLinting > 0 ||
		cfg.Timeouts.ToolTesting > 0 || cfg.Timeouts.ToolReport > 0 || cfg.Timeouts.OllamaDownload > 0 ||
		cfg.Timeouts.OllamaGenerate > 0 || cfg.Timeouts.HTTPClient > 0 || cfg.Timeouts.DatabaseRetry > 0 ||
		cfg.Timeouts.ContextSummarize > 0 || cfg.Timeouts.ContextBudget > 0 {
		pb.Timeouts = timeoutsToProtobuf(&cfg.Timeouts)
	}

	if cfg.Thresholds.SimilarityThreshold > 0 || cfg.Thresholds.MinDescriptionLength > 0 ||
		cfg.Thresholds.MinTaskConfidence > 0 || cfg.Thresholds.MinCoverage > 0 ||
		cfg.Thresholds.MinTestConfidence > 0 || cfg.Thresholds.MinEstimationConfidence > 0 ||
		cfg.Thresholds.MLXWeight > 0 || cfg.Thresholds.MaxParallelTasks > 0 ||
		cfg.Thresholds.MaxTasksPerHost > 0 || cfg.Thresholds.MaxTasksPerWave > 0 || cfg.Thresholds.MaxAutomationIterations > 0 ||
		cfg.Thresholds.TokensPerChar > 0 || cfg.Thresholds.DefaultContextBudget > 0 ||
		cfg.Thresholds.ContextReductionThreshold > 0 || cfg.Thresholds.RateLimitRequests > 0 ||
		cfg.Thresholds.RateLimitWindow > 0 || cfg.Thresholds.MaxFileSize > 0 ||
		cfg.Thresholds.MaxPathDepth > 0 {
		pb.Thresholds = thresholdsToProtobuf(&cfg.Thresholds)
	}

	if cfg.Tasks.DefaultStatus != "" || cfg.Tasks.DefaultPriority != "" ||
		len(cfg.Tasks.DefaultTags) > 0 || len(cfg.Tasks.StatusWorkflow) > 0 ||
		cfg.Tasks.StaleThresholdHours > 0 || cfg.Tasks.AutoCleanupEnabled ||
		cfg.Tasks.CleanupDryRun || cfg.Tasks.IDFormat != "" || cfg.Tasks.IDPrefix != "" ||
		cfg.Tasks.MinDescriptionLength > 0 || cfg.Tasks.RequireDescription ||
		cfg.Tasks.AutoClarify {
		pb.Tasks = tasksToProtobuf(&cfg.Tasks)
	}

	if cfg.Database.SQLitePath != "" || cfg.Database.JSONFallbackPath != "" ||
		cfg.Database.BackupPath != "" || cfg.Database.MaxConnections > 0 ||
		cfg.Database.ConnectionTimeout > 0 || cfg.Database.QueryTimeout > 0 ||
		cfg.Database.RetryAttempts > 0 || cfg.Database.RetryInitialDelay > 0 ||
		cfg.Database.RetryMaxDelay > 0 || cfg.Database.RetryMultiplier > 0 ||
		cfg.Database.AutoVacuum || cfg.Database.WALMode ||
		cfg.Database.CheckpointInterval > 0 || cfg.Database.BackupRetentionDays > 0 {
		pb.Database = databaseToProtobuf(&cfg.Database)
	}

	if cfg.Security.RateLimit.Enabled || cfg.Security.RateLimit.RequestsPerWindow > 0 ||
		cfg.Security.RateLimit.WindowDuration > 0 || cfg.Security.RateLimit.BurstSize > 0 ||
		cfg.Security.PathValidation.Enabled || cfg.Security.PathValidation.AllowAbsolutePaths ||
		cfg.Security.PathValidation.MaxDepth > 0 || len(cfg.Security.PathValidation.BlockedPatterns) > 0 ||
		cfg.Security.FileLimits.MaxFileSize > 0 || cfg.Security.FileLimits.MaxFilesPerOperation > 0 ||
		len(cfg.Security.FileLimits.AllowedExtensions) > 0 || cfg.Security.AccessControl.Enabled ||
		cfg.Security.AccessControl.DefaultPolicy != "" || len(cfg.Security.AccessControl.RestrictedTools) > 0 {
		pb.Security = securityToProtobuf(&cfg.Security)
	}

	if cfg.Logging.Level != "" || cfg.Logging.ToolLevel != "" || cfg.Logging.FrameworkLevel != "" ||
		cfg.Logging.Format != "" || cfg.Logging.IncludeTimestamps || cfg.Logging.IncludeCaller ||
		cfg.Logging.ColorOutput || cfg.Logging.LogDir != "" || cfg.Logging.LogFile != "" ||
		cfg.Logging.SessionLogDir != "" || cfg.Logging.RetentionDays > 0 || cfg.Logging.AutoCleanup ||
		cfg.Logging.LogRotation.Enabled || cfg.Logging.LogRotation.MaxSize > 0 ||
		cfg.Logging.LogRotation.MaxFiles > 0 || cfg.Logging.LogRotation.Compress {
		pb.Logging = loggingToProtobuf(&cfg.Logging)
	}

	if cfg.Tools.Scorecard.IncludeWisdom || cfg.Tools.Scorecard.OutputFormat != "" ||
		len(cfg.Tools.Scorecard.DefaultScores) > 0 || cfg.Tools.Report.DefaultFormat != "" ||
		cfg.Tools.Report.DefaultOutputPath != "" || cfg.Tools.Report.IncludeMetrics ||
		cfg.Tools.Report.IncludeRecommendations || cfg.Tools.Linting.DefaultLinter != "" ||
		cfg.Tools.Linting.AutoFix || cfg.Tools.Linting.IncludeHints || cfg.Tools.Linting.Timeout > 0 ||
		cfg.Tools.Testing.DefaultFramework != "" || cfg.Tools.Testing.MinCoverage > 0 ||
		cfg.Tools.Testing.CoverageFormat != "" || cfg.Tools.Testing.Verbose ||
		cfg.Tools.MLX.DefaultModel != "" || cfg.Tools.MLX.DefaultMaxTokens > 0 ||
		cfg.Tools.MLX.DefaultTemperature > 0 || cfg.Tools.MLX.Verbose ||
		cfg.Tools.Ollama.DefaultModel != "" || cfg.Tools.Ollama.DefaultHost != "" ||
		cfg.Tools.Ollama.DefaultContextSize > 0 || cfg.Tools.Ollama.DefaultNumThreads > 0 ||
		cfg.Tools.Ollama.DefaultNumGPU > 0 || cfg.Tools.Context.DefaultBudget > 0 ||
		cfg.Tools.Context.DefaultSummarizationLevel != "" || cfg.Tools.Context.TokensPerChar > 0 ||
		cfg.Tools.Context.IncludeRaw {
		pb.Tools = toolsToProtobuf(&cfg.Tools)
	}

	if cfg.Workflow.DefaultMode != "" || cfg.Workflow.AutoDetectMode ||
		cfg.Workflow.ModeSuggestions.Morning != "" || cfg.Workflow.ModeSuggestions.Afternoon != "" ||
		cfg.Workflow.ModeSuggestions.Evening != "" || len(cfg.Workflow.Modes) > 0 ||
		cfg.Workflow.Focus.Enabled || cfg.Workflow.Focus.ReductionTarget > 0 ||
		cfg.Workflow.Focus.PreserveCoreTools {
		pb.Workflow = workflowToProtobuf(&cfg.Workflow)
	}

	if len(cfg.Memory.Categories) > 0 || cfg.Memory.StoragePath != "" ||
		cfg.Memory.SessionLogPath != "" || cfg.Memory.RetentionDays > 0 ||
		cfg.Memory.AutoCleanup || cfg.Memory.MaxMemories > 0 ||
		cfg.Memory.Consolidation.Enabled || cfg.Memory.Consolidation.SimilarityThreshold > 0 ||
		cfg.Memory.Consolidation.Frequency != "" {
		pb.Memory = memoryToProtobuf(&cfg.Memory)
	}

	if cfg.Project.Name != "" || cfg.Project.Type != "" || cfg.Project.Language != "" ||
		cfg.Project.Root != "" || cfg.Project.Todo2Path != "" || cfg.Project.ExarpPath != "" ||
		cfg.Project.Features.SQLiteEnabled || cfg.Project.Features.JSONFallback ||
		len(cfg.Project.Features.MCPServers) > 0 ||
		len(cfg.Project.SkipChecks) > 0 || len(cfg.Project.CustomTools) > 0 ||
		len(cfg.Project.TaskDiscoveryIgnorePaths) > 0 {
		pb.Project = projectToProtobuf(&cfg.Project)
	}

	return pb, nil
}

// FromProtobuf converts protobuf FullConfig to Go FullConfig.
func FromProtobuf(pb *configpb.FullConfig) (*FullConfig, error) {
	if pb == nil {
		return nil, fmt.Errorf("protobuf config cannot be nil")
	}

	cfg := &FullConfig{
		Version: pb.GetVersion(),
	}

	if pb.GetTimeouts() != nil {
		cfg.Timeouts = timeoutsFromProtobuf(pb.GetTimeouts())
	}
	if pb.GetThresholds() != nil {
		cfg.Thresholds = thresholdsFromProtobuf(pb.GetThresholds())
	}
	if pb.GetTasks() != nil {
		var err error
		cfg.Tasks, err = tasksFromProtobuf(pb.GetTasks())
		if err != nil {
			return nil, fmt.Errorf("failed to convert tasks config: %w", err)
		}
	}
	if pb.GetDatabase() != nil {
		cfg.Database = databaseFromProtobuf(pb.GetDatabase())
	}
	if pb.GetSecurity() != nil {
		cfg.Security = securityFromProtobuf(pb.GetSecurity())
	}
	if pb.GetLogging() != nil {
		cfg.Logging = loggingFromProtobuf(pb.GetLogging())
	}
	if pb.GetTools() != nil {
		var err error
		cfg.Tools, err = toolsFromProtobuf(pb.GetTools())
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools config: %w", err)
		}
	}
	if pb.GetWorkflow() != nil {
		var err error
		cfg.Workflow, err = workflowFromProtobuf(pb.GetWorkflow())
		if err != nil {
			return nil, fmt.Errorf("failed to convert workflow config: %w", err)
		}
	}
	if pb.GetMemory() != nil {
		cfg.Memory = memoryFromProtobuf(pb.GetMemory())
	}
	if pb.GetProject() != nil {
		cfg.Project = projectFromProtobuf(pb.GetProject())
	}

	return cfg, nil
}
