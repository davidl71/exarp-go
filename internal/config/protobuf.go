package config

import (
	"encoding/json"
	"fmt"
	"time"

	configpb "github.com/davidl71/exarp-go/proto"
)

// ToProtobuf converts Go FullConfig to protobuf FullConfig
func ToProtobuf(cfg *FullConfig) (*configpb.FullConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	pb := &configpb.FullConfig{
		Version: cfg.Version,
	}

	// Convert nested configs
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
		cfg.Thresholds.MaxTasksPerHost > 0 || cfg.Thresholds.MaxAutomationIterations > 0 ||
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
		cfg.Project.Features.PythonBridge || len(cfg.Project.Features.MCPServers) > 0 ||
		len(cfg.Project.SkipChecks) > 0 || len(cfg.Project.CustomTools) > 0 {
		pb.Project = projectToProtobuf(&cfg.Project)
	}

	// AutomationsConfig is not in protobuf schema (will be added in Phase 5)
	// Skip it for now

	return pb, nil
}

// FromProtobuf converts protobuf FullConfig to Go FullConfig
func FromProtobuf(pb *configpb.FullConfig) (*FullConfig, error) {
	if pb == nil {
		return nil, fmt.Errorf("protobuf config cannot be nil")
	}

	cfg := &FullConfig{
		Version: pb.GetVersion(),
	}

	// Convert nested configs
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

	// AutomationsConfig is not in protobuf schema (will be added in Phase 5)
	// Leave it as zero value

	return cfg, nil
}

// Helper functions for duration conversion

// durationToSeconds converts Go time.Duration to protobuf int64 (seconds)
func durationToSeconds(d time.Duration) int64 {
	return int64(d.Seconds())
}

// secondsToDuration converts protobuf int64 (seconds) to Go time.Duration
func secondsToDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

// Helper functions for JSON string conversion (for maps)

// mapToJSON converts a map to JSON string
func mapToJSON(m interface{}) (string, error) {
	if m == nil {
		return "", nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return string(data), nil
}

// jsonToMap converts JSON string to a map
func jsonToMap(jsonStr string, target interface{}) error {
	if jsonStr == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}
	return nil
}

// Conversion functions for each config type

func timeoutsToProtobuf(t *TimeoutsConfig) *configpb.TimeoutsConfig {
	if t == nil {
		return nil
	}
	return &configpb.TimeoutsConfig{
		TaskLockLease:      durationToSeconds(t.TaskLockLease),
		TaskLockRenewal:    durationToSeconds(t.TaskLockRenewal),
		StaleLockThreshold: durationToSeconds(t.StaleLockThreshold),
		ToolDefault:        durationToSeconds(t.ToolDefault),
		ToolScorecard:      durationToSeconds(t.ToolScorecard),
		ToolLinting:        durationToSeconds(t.ToolLinting),
		ToolTesting:        durationToSeconds(t.ToolTesting),
		ToolReport:         durationToSeconds(t.ToolReport),
		OllamaDownload:     durationToSeconds(t.OllamaDownload),
		OllamaGenerate:     durationToSeconds(t.OllamaGenerate),
		HttpClient:         durationToSeconds(t.HTTPClient),
		DatabaseRetry:      durationToSeconds(t.DatabaseRetry),
		ContextSummarize:   durationToSeconds(t.ContextSummarize),
		ContextBudget:      durationToSeconds(t.ContextBudget),
	}
}

func timeoutsFromProtobuf(pb *configpb.TimeoutsConfig) TimeoutsConfig {
	if pb == nil {
		return TimeoutsConfig{}
	}
	return TimeoutsConfig{
		TaskLockLease:      secondsToDuration(pb.GetTaskLockLease()),
		TaskLockRenewal:    secondsToDuration(pb.GetTaskLockRenewal()),
		StaleLockThreshold: secondsToDuration(pb.GetStaleLockThreshold()),
		ToolDefault:        secondsToDuration(pb.GetToolDefault()),
		ToolScorecard:      secondsToDuration(pb.GetToolScorecard()),
		ToolLinting:        secondsToDuration(pb.GetToolLinting()),
		ToolTesting:        secondsToDuration(pb.GetToolTesting()),
		ToolReport:         secondsToDuration(pb.GetToolReport()),
		OllamaDownload:     secondsToDuration(pb.GetOllamaDownload()),
		OllamaGenerate:     secondsToDuration(pb.GetOllamaGenerate()),
		HTTPClient:         secondsToDuration(pb.GetHttpClient()),
		DatabaseRetry:      secondsToDuration(pb.GetDatabaseRetry()),
		ContextSummarize:   secondsToDuration(pb.GetContextSummarize()),
		ContextBudget:      secondsToDuration(pb.GetContextBudget()),
	}
}

func thresholdsToProtobuf(t *ThresholdsConfig) *configpb.ThresholdsConfig {
	if t == nil {
		return nil
	}
	return &configpb.ThresholdsConfig{
		SimilarityThreshold:       t.SimilarityThreshold,
		MinDescriptionLength:       int32(t.MinDescriptionLength),
		MinTaskConfidence:          t.MinTaskConfidence,
		MinCoverage:                int32(t.MinCoverage),
		MinTestConfidence:          t.MinTestConfidence,
		MinEstimationConfidence:    t.MinEstimationConfidence,
		MlxWeight:                  t.MLXWeight,
		MaxParallelTasks:           int32(t.MaxParallelTasks),
		MaxTasksPerHost:            int32(t.MaxTasksPerHost),
		MaxAutomationIterations:    int32(t.MaxAutomationIterations),
		TokensPerChar:              t.TokensPerChar,
		DefaultContextBudget:       int32(t.DefaultContextBudget),
		ContextReductionThreshold: t.ContextReductionThreshold,
		RateLimitRequests:         int32(t.RateLimitRequests),
		RateLimitWindow:           durationToSeconds(t.RateLimitWindow),
		MaxFileSize:                t.MaxFileSize,
		MaxPathDepth:               int32(t.MaxPathDepth),
	}
}

func thresholdsFromProtobuf(pb *configpb.ThresholdsConfig) ThresholdsConfig {
	if pb == nil {
		return ThresholdsConfig{}
	}
	return ThresholdsConfig{
		SimilarityThreshold:       pb.GetSimilarityThreshold(),
		MinDescriptionLength:       int(pb.GetMinDescriptionLength()),
		MinTaskConfidence:          pb.GetMinTaskConfidence(),
		MinCoverage:                int(pb.GetMinCoverage()),
		MinTestConfidence:          pb.GetMinTestConfidence(),
		MinEstimationConfidence:    pb.GetMinEstimationConfidence(),
		MLXWeight:                  pb.GetMlxWeight(),
		MaxParallelTasks:           int(pb.GetMaxParallelTasks()),
		MaxTasksPerHost:            int(pb.GetMaxTasksPerHost()),
		MaxAutomationIterations:    int(pb.GetMaxAutomationIterations()),
		TokensPerChar:              pb.GetTokensPerChar(),
		DefaultContextBudget:       int(pb.GetDefaultContextBudget()),
		ContextReductionThreshold: pb.GetContextReductionThreshold(),
		RateLimitRequests:         int(pb.GetRateLimitRequests()),
		RateLimitWindow:           secondsToDuration(pb.GetRateLimitWindow()),
		MaxFileSize:                pb.GetMaxFileSize(),
		MaxPathDepth:               int(pb.GetMaxPathDepth()),
	}
}

func tasksToProtobuf(t *TasksConfig) *configpb.TasksConfig {
	if t == nil {
		return nil
	}
	
	// Convert StatusWorkflow map to JSON string
	statusWorkflowJSON, _ := mapToJSON(t.StatusWorkflow)
	
	return &configpb.TasksConfig{
		DefaultStatus:        t.DefaultStatus,
		DefaultPriority:      t.DefaultPriority,
		DefaultTags:          t.DefaultTags,
		StatusWorkflowJson:   statusWorkflowJSON,
		StaleThresholdHours: int32(t.StaleThresholdHours),
		AutoCleanupEnabled:  t.AutoCleanupEnabled,
		CleanupDryRun:        t.CleanupDryRun,
		IdFormat:             t.IDFormat,
		IdPrefix:             t.IDPrefix,
		MinDescriptionLength: int32(t.MinDescriptionLength),
		RequireDescription:   t.RequireDescription,
		AutoClarify:          t.AutoClarify,
	}
}

func tasksFromProtobuf(pb *configpb.TasksConfig) (TasksConfig, error) {
	if pb == nil {
		return TasksConfig{}, nil
	}
	
	// Convert JSON string to StatusWorkflow map
	var statusWorkflow map[string][]string
	if pb.GetStatusWorkflowJson() != "" {
		if err := jsonToMap(pb.GetStatusWorkflowJson(), &statusWorkflow); err != nil {
			return TasksConfig{}, fmt.Errorf("failed to parse status_workflow_json: %w", err)
		}
	}
	
	return TasksConfig{
		DefaultStatus:        pb.GetDefaultStatus(),
		DefaultPriority:       pb.GetDefaultPriority(),
		DefaultTags:           pb.GetDefaultTags(),
		StatusWorkflow:        statusWorkflow,
		StaleThresholdHours:  int(pb.GetStaleThresholdHours()),
		AutoCleanupEnabled:   pb.GetAutoCleanupEnabled(),
		CleanupDryRun:        pb.GetCleanupDryRun(),
		IDFormat:             pb.GetIdFormat(),
		IDPrefix:             pb.GetIdPrefix(),
		MinDescriptionLength: int(pb.GetMinDescriptionLength()),
		RequireDescription:   pb.GetRequireDescription(),
		AutoClarify:          pb.GetAutoClarify(),
	}, nil
}

func databaseToProtobuf(d *DatabaseConfig) *configpb.DatabaseConfig {
	if d == nil {
		return nil
	}
	return &configpb.DatabaseConfig{
		SqlitePath:         d.SQLitePath,
		JsonFallbackPath:   d.JSONFallbackPath,
		BackupPath:         d.BackupPath,
		MaxConnections:     int32(d.MaxConnections),
		ConnectionTimeout:  durationToSeconds(d.ConnectionTimeout),
		QueryTimeout:       durationToSeconds(d.QueryTimeout),
		RetryAttempts:      int32(d.RetryAttempts),
		RetryInitialDelay:  durationToSeconds(d.RetryInitialDelay),
		RetryMaxDelay:      durationToSeconds(d.RetryMaxDelay),
		RetryMultiplier:    d.RetryMultiplier,
		AutoVacuum:         d.AutoVacuum,
		WalMode:            d.WALMode,
		CheckpointInterval: int32(d.CheckpointInterval),
		BackupRetentionDays: int32(d.BackupRetentionDays),
	}
}

func databaseFromProtobuf(pb *configpb.DatabaseConfig) DatabaseConfig {
	if pb == nil {
		return DatabaseConfig{}
	}
	return DatabaseConfig{
		SQLitePath:         pb.GetSqlitePath(),
		JSONFallbackPath:   pb.GetJsonFallbackPath(),
		BackupPath:         pb.GetBackupPath(),
		MaxConnections:     int(pb.GetMaxConnections()),
		ConnectionTimeout: secondsToDuration(pb.GetConnectionTimeout()),
		QueryTimeout:       secondsToDuration(pb.GetQueryTimeout()),
		RetryAttempts:      int(pb.GetRetryAttempts()),
		RetryInitialDelay:  secondsToDuration(pb.GetRetryInitialDelay()),
		RetryMaxDelay:      secondsToDuration(pb.GetRetryMaxDelay()),
		RetryMultiplier:    pb.GetRetryMultiplier(),
		AutoVacuum:         pb.GetAutoVacuum(),
		WALMode:            pb.GetWalMode(),
		CheckpointInterval: int(pb.GetCheckpointInterval()),
		BackupRetentionDays: int(pb.GetBackupRetentionDays()),
	}
}

func securityToProtobuf(s *SecurityConfig) *configpb.SecurityConfig {
	if s == nil {
		return nil
	}
	return &configpb.SecurityConfig{
		RateLimit:      rateLimitToProtobuf(&s.RateLimit),
		PathValidation: pathValidationToProtobuf(&s.PathValidation),
		FileLimits:     fileLimitsToProtobuf(&s.FileLimits),
		AccessControl:  accessControlToProtobuf(&s.AccessControl),
	}
}

func securityFromProtobuf(pb *configpb.SecurityConfig) SecurityConfig {
	if pb == nil {
		return SecurityConfig{}
	}
	return SecurityConfig{
		RateLimit:      rateLimitFromProtobuf(pb.GetRateLimit()),
		PathValidation: pathValidationFromProtobuf(pb.GetPathValidation()),
		FileLimits:     fileLimitsFromProtobuf(pb.GetFileLimits()),
		AccessControl:  accessControlFromProtobuf(pb.GetAccessControl()),
	}
}

func rateLimitToProtobuf(r *RateLimitConfig) *configpb.RateLimitConfig {
	if r == nil {
		return nil
	}
	return &configpb.RateLimitConfig{
		Enabled:           r.Enabled,
		RequestsPerWindow: int32(r.RequestsPerWindow),
		WindowDuration:    durationToSeconds(r.WindowDuration),
		BurstSize:         int32(r.BurstSize),
	}
}

func rateLimitFromProtobuf(pb *configpb.RateLimitConfig) RateLimitConfig {
	if pb == nil {
		return RateLimitConfig{}
	}
	return RateLimitConfig{
		Enabled:           pb.GetEnabled(),
		RequestsPerWindow: int(pb.GetRequestsPerWindow()),
		WindowDuration:    secondsToDuration(pb.GetWindowDuration()),
		BurstSize:         int(pb.GetBurstSize()),
	}
}

func pathValidationToProtobuf(p *PathValidationConfig) *configpb.PathValidationConfig {
	if p == nil {
		return nil
	}
	return &configpb.PathValidationConfig{
		Enabled:          p.Enabled,
		AllowAbsolutePaths: p.AllowAbsolutePaths,
		MaxDepth:         int32(p.MaxDepth),
		BlockedPatterns:  p.BlockedPatterns,
	}
}

func pathValidationFromProtobuf(pb *configpb.PathValidationConfig) PathValidationConfig {
	if pb == nil {
		return PathValidationConfig{}
	}
	return PathValidationConfig{
		Enabled:          pb.GetEnabled(),
		AllowAbsolutePaths: pb.GetAllowAbsolutePaths(),
		MaxDepth:         int(pb.GetMaxDepth()),
		BlockedPatterns:  pb.GetBlockedPatterns(),
	}
}

func fileLimitsToProtobuf(f *FileLimitsConfig) *configpb.FileLimitsConfig {
	if f == nil {
		return nil
	}
	return &configpb.FileLimitsConfig{
		MaxFileSize:          f.MaxFileSize,
		MaxFilesPerOperation: int32(f.MaxFilesPerOperation),
		AllowedExtensions:    f.AllowedExtensions,
	}
}

func fileLimitsFromProtobuf(pb *configpb.FileLimitsConfig) FileLimitsConfig {
	if pb == nil {
		return FileLimitsConfig{}
	}
	return FileLimitsConfig{
		MaxFileSize:          pb.GetMaxFileSize(),
		MaxFilesPerOperation: int(pb.GetMaxFilesPerOperation()),
		AllowedExtensions:    pb.GetAllowedExtensions(),
	}
}

func accessControlToProtobuf(a *AccessControlConfig) *configpb.AccessControlConfig {
	if a == nil {
		return nil
	}
	return &configpb.AccessControlConfig{
		Enabled:         a.Enabled,
		DefaultPolicy:   a.DefaultPolicy,
		RestrictedTools: a.RestrictedTools,
	}
}

func accessControlFromProtobuf(pb *configpb.AccessControlConfig) AccessControlConfig {
	if pb == nil {
		return AccessControlConfig{}
	}
	return AccessControlConfig{
		Enabled:         pb.GetEnabled(),
		DefaultPolicy:   pb.GetDefaultPolicy(),
		RestrictedTools: pb.GetRestrictedTools(),
	}
}

func loggingToProtobuf(l *LoggingConfig) *configpb.LoggingConfig {
	if l == nil {
		return nil
	}
	return &configpb.LoggingConfig{
		Level:            l.Level,
		ToolLevel:        l.ToolLevel,
		FrameworkLevel:   l.FrameworkLevel,
		Format:           l.Format,
		IncludeTimestamps: l.IncludeTimestamps,
		IncludeCaller:    l.IncludeCaller,
		ColorOutput:      l.ColorOutput,
		LogDir:           l.LogDir,
		LogFile:          l.LogFile,
		SessionLogDir:    l.SessionLogDir,
		LogRotation:      logRotationToProtobuf(&l.LogRotation),
		RetentionDays:    int32(l.RetentionDays),
		AutoCleanup:      l.AutoCleanup,
	}
}

func loggingFromProtobuf(pb *configpb.LoggingConfig) LoggingConfig {
	if pb == nil {
		return LoggingConfig{}
	}
	return LoggingConfig{
		Level:            pb.GetLevel(),
		ToolLevel:        pb.GetToolLevel(),
		FrameworkLevel:   pb.GetFrameworkLevel(),
		Format:           pb.GetFormat(),
		IncludeTimestamps: pb.GetIncludeTimestamps(),
		IncludeCaller:    pb.GetIncludeCaller(),
		ColorOutput:      pb.GetColorOutput(),
		LogDir:           pb.GetLogDir(),
		LogFile:          pb.GetLogFile(),
		SessionLogDir:    pb.GetSessionLogDir(),
		LogRotation:      logRotationFromProtobuf(pb.GetLogRotation()),
		RetentionDays:    int(pb.GetRetentionDays()),
		AutoCleanup:      pb.GetAutoCleanup(),
	}
}

func logRotationToProtobuf(l *LogRotationConfig) *configpb.LogRotationConfig {
	if l == nil {
		return nil
	}
	return &configpb.LogRotationConfig{
		Enabled:  l.Enabled,
		MaxSize:  l.MaxSize,
		MaxFiles: int32(l.MaxFiles),
		Compress: l.Compress,
	}
}

func logRotationFromProtobuf(pb *configpb.LogRotationConfig) LogRotationConfig {
	if pb == nil {
		return LogRotationConfig{}
	}
	return LogRotationConfig{
		Enabled:  pb.GetEnabled(),
		MaxSize:  pb.GetMaxSize(),
		MaxFiles: int(pb.GetMaxFiles()),
		Compress: pb.GetCompress(),
	}
}

func toolsToProtobuf(t *ToolsConfig) *configpb.ToolsConfig {
	if t == nil {
		return nil
	}
	return &configpb.ToolsConfig{
		Scorecard: scorecardToProtobuf(&t.Scorecard),
		Report:    reportToProtobuf(&t.Report),
		Linting:   lintingToProtobuf(&t.Linting),
		Testing:   testingToProtobuf(&t.Testing),
		Mlx:       mlxToProtobuf(&t.MLX),
		Ollama:    ollamaToProtobuf(&t.Ollama),
		Context:   contextToProtobuf(&t.Context),
	}
}

func toolsFromProtobuf(pb *configpb.ToolsConfig) (ToolsConfig, error) {
	if pb == nil {
		return ToolsConfig{}, nil
	}
	
	scorecard, err := scorecardFromProtobuf(pb.GetScorecard())
	if err != nil {
		return ToolsConfig{}, fmt.Errorf("failed to convert scorecard config: %w", err)
	}
	
	return ToolsConfig{
		Scorecard: scorecard,
		Report:    reportFromProtobuf(pb.GetReport()),
		Linting:   lintingFromProtobuf(pb.GetLinting()),
		Testing:   testingFromProtobuf(pb.GetTesting()),
		MLX:       mlxFromProtobuf(pb.GetMlx()),
		Ollama:    ollamaFromProtobuf(pb.GetOllama()),
		Context:   contextFromProtobuf(pb.GetContext()),
	}, nil
}

func scorecardToProtobuf(s *ScorecardConfig) *configpb.ScorecardConfig {
	if s == nil {
		return nil
	}
	
	// Convert DefaultScores map to JSON string
	defaultScoresJSON, _ := mapToJSON(s.DefaultScores)
	
	return &configpb.ScorecardConfig{
		DefaultScoresJson: defaultScoresJSON,
		IncludeWisdom:     s.IncludeWisdom,
		OutputFormat:      s.OutputFormat,
	}
}

func scorecardFromProtobuf(pb *configpb.ScorecardConfig) (ScorecardConfig, error) {
	if pb == nil {
		return ScorecardConfig{}, nil
	}
	
	// Convert JSON string to DefaultScores map
	var defaultScores map[string]float64
	if pb.GetDefaultScoresJson() != "" {
		if err := jsonToMap(pb.GetDefaultScoresJson(), &defaultScores); err != nil {
			return ScorecardConfig{}, fmt.Errorf("failed to parse default_scores_json: %w", err)
		}
	}
	
	return ScorecardConfig{
		DefaultScores: defaultScores,
		IncludeWisdom: pb.GetIncludeWisdom(),
		OutputFormat:  pb.GetOutputFormat(),
	}, nil
}

func reportToProtobuf(r *ReportConfig) *configpb.ReportConfig {
	if r == nil {
		return nil
	}
	return &configpb.ReportConfig{
		DefaultFormat:         r.DefaultFormat,
		DefaultOutputPath:     r.DefaultOutputPath,
		IncludeMetrics:        r.IncludeMetrics,
		IncludeRecommendations: r.IncludeRecommendations,
	}
}

func reportFromProtobuf(pb *configpb.ReportConfig) ReportConfig {
	if pb == nil {
		return ReportConfig{}
	}
	return ReportConfig{
		DefaultFormat:         pb.GetDefaultFormat(),
		DefaultOutputPath:     pb.GetDefaultOutputPath(),
		IncludeMetrics:        pb.GetIncludeMetrics(),
		IncludeRecommendations: pb.GetIncludeRecommendations(),
	}
}

func lintingToProtobuf(l *LintingConfig) *configpb.LintingConfig {
	if l == nil {
		return nil
	}
	return &configpb.LintingConfig{
		DefaultLinter: l.DefaultLinter,
		AutoFix:       l.AutoFix,
		IncludeHints:  l.IncludeHints,
		Timeout:       durationToSeconds(l.Timeout),
	}
}

func lintingFromProtobuf(pb *configpb.LintingConfig) LintingConfig {
	if pb == nil {
		return LintingConfig{}
	}
	return LintingConfig{
		DefaultLinter: pb.GetDefaultLinter(),
		AutoFix:       pb.GetAutoFix(),
		IncludeHints:  pb.GetIncludeHints(),
		Timeout:       secondsToDuration(pb.GetTimeout()),
	}
}

func testingToProtobuf(t *TestingConfig) *configpb.TestingConfig {
	if t == nil {
		return nil
	}
	return &configpb.TestingConfig{
		DefaultFramework: t.DefaultFramework,
		MinCoverage:       int32(t.MinCoverage),
		CoverageFormat:    t.CoverageFormat,
		Verbose:           t.Verbose,
	}
}

func testingFromProtobuf(pb *configpb.TestingConfig) TestingConfig {
	if pb == nil {
		return TestingConfig{}
	}
	return TestingConfig{
		DefaultFramework: pb.GetDefaultFramework(),
		MinCoverage:       int(pb.GetMinCoverage()),
		CoverageFormat:    pb.GetCoverageFormat(),
		Verbose:           pb.GetVerbose(),
	}
}

func mlxToProtobuf(m *MLXConfig) *configpb.MLXConfig {
	if m == nil {
		return nil
	}
	return &configpb.MLXConfig{
		DefaultModel:      m.DefaultModel,
		DefaultMaxTokens:   int32(m.DefaultMaxTokens),
		DefaultTemperature: m.DefaultTemperature,
		Verbose:            m.Verbose,
	}
}

func mlxFromProtobuf(pb *configpb.MLXConfig) MLXConfig {
	if pb == nil {
		return MLXConfig{}
	}
	return MLXConfig{
		DefaultModel:      pb.GetDefaultModel(),
		DefaultMaxTokens:   int(pb.GetDefaultMaxTokens()),
		DefaultTemperature: pb.GetDefaultTemperature(),
		Verbose:            pb.GetVerbose(),
	}
}

func ollamaToProtobuf(o *OllamaConfig) *configpb.OllamaConfig {
	if o == nil {
		return nil
	}
	return &configpb.OllamaConfig{
		DefaultModel:       o.DefaultModel,
		DefaultHost:        o.DefaultHost,
		DefaultContextSize: int32(o.DefaultContextSize),
		DefaultNumThreads:  int32(o.DefaultNumThreads),
		DefaultNumGpu:      int32(o.DefaultNumGPU),
	}
}

func ollamaFromProtobuf(pb *configpb.OllamaConfig) OllamaConfig {
	if pb == nil {
		return OllamaConfig{}
	}
	return OllamaConfig{
		DefaultModel:       pb.GetDefaultModel(),
		DefaultHost:        pb.GetDefaultHost(),
		DefaultContextSize: int(pb.GetDefaultContextSize()),
		DefaultNumThreads:  int(pb.GetDefaultNumThreads()),
		DefaultNumGPU:      int(pb.GetDefaultNumGpu()),
	}
}

func contextToProtobuf(c *ContextConfig) *configpb.ContextConfig {
	if c == nil {
		return nil
	}
	return &configpb.ContextConfig{
		DefaultBudget:            int32(c.DefaultBudget),
		DefaultSummarizationLevel: c.DefaultSummarizationLevel,
		TokensPerChar:            c.TokensPerChar,
		IncludeRaw:               c.IncludeRaw,
	}
}

func contextFromProtobuf(pb *configpb.ContextConfig) ContextConfig {
	if pb == nil {
		return ContextConfig{}
	}
	return ContextConfig{
		DefaultBudget:            int(pb.GetDefaultBudget()),
		DefaultSummarizationLevel: pb.GetDefaultSummarizationLevel(),
		TokensPerChar:            pb.GetTokensPerChar(),
		IncludeRaw:               pb.GetIncludeRaw(),
	}
}

func workflowToProtobuf(w *WorkflowConfig) *configpb.WorkflowConfig {
	if w == nil {
		return nil
	}
	
	// Convert Modes map to JSON string
	modesJSON, _ := mapToJSON(w.Modes)
	
	return &configpb.WorkflowConfig{
		DefaultMode:    w.DefaultMode,
		AutoDetectMode: w.AutoDetectMode,
		ModeSuggestions: modeSuggestionsToProtobuf(&w.ModeSuggestions),
		ModesJson:      modesJSON,
		Focus:          focusToProtobuf(&w.Focus),
	}
}

func workflowFromProtobuf(pb *configpb.WorkflowConfig) (WorkflowConfig, error) {
	if pb == nil {
		return WorkflowConfig{}, nil
	}
	
	// Convert JSON string to Modes map
	var modes map[string]ModeConfig
	if pb.GetModesJson() != "" {
		if err := jsonToMap(pb.GetModesJson(), &modes); err != nil {
			return WorkflowConfig{}, fmt.Errorf("failed to parse modes_json: %w", err)
		}
	}
	
	return WorkflowConfig{
		DefaultMode:    pb.GetDefaultMode(),
		AutoDetectMode: pb.GetAutoDetectMode(),
		ModeSuggestions: modeSuggestionsFromProtobuf(pb.GetModeSuggestions()),
		Modes:          modes,
		Focus:          focusFromProtobuf(pb.GetFocus()),
	}, nil
}

func modeSuggestionsToProtobuf(m *ModeSuggestionsConfig) *configpb.ModeSuggestionsConfig {
	if m == nil {
		return nil
	}
	return &configpb.ModeSuggestionsConfig{
		Morning:   m.Morning,
		Afternoon: m.Afternoon,
		Evening:   m.Evening,
	}
}

func modeSuggestionsFromProtobuf(pb *configpb.ModeSuggestionsConfig) ModeSuggestionsConfig {
	if pb == nil {
		return ModeSuggestionsConfig{}
	}
	return ModeSuggestionsConfig{
		Morning:   pb.GetMorning(),
		Afternoon: pb.GetAfternoon(),
		Evening:   pb.GetEvening(),
	}
}

func focusToProtobuf(f *FocusConfig) *configpb.FocusConfig {
	if f == nil {
		return nil
	}
	return &configpb.FocusConfig{
		Enabled:          f.Enabled,
		ReductionTarget:  int32(f.ReductionTarget),
		PreserveCoreTools: f.PreserveCoreTools,
	}
}

func focusFromProtobuf(pb *configpb.FocusConfig) FocusConfig {
	if pb == nil {
		return FocusConfig{}
	}
	return FocusConfig{
		Enabled:          pb.GetEnabled(),
		ReductionTarget:  int(pb.GetReductionTarget()),
		PreserveCoreTools: pb.GetPreserveCoreTools(),
	}
}

func memoryToProtobuf(m *MemoryConfig) *configpb.MemoryConfig {
	if m == nil {
		return nil
	}
	return &configpb.MemoryConfig{
		Categories:     m.Categories,
		StoragePath:    m.StoragePath,
		SessionLogPath: m.SessionLogPath,
		RetentionDays:  int32(m.RetentionDays),
		AutoCleanup:    m.AutoCleanup,
		MaxMemories:    int32(m.MaxMemories),
		Consolidation:  consolidationToProtobuf(&m.Consolidation),
	}
}

func memoryFromProtobuf(pb *configpb.MemoryConfig) MemoryConfig {
	if pb == nil {
		return MemoryConfig{}
	}
	return MemoryConfig{
		Categories:     pb.GetCategories(),
		StoragePath:    pb.GetStoragePath(),
		SessionLogPath: pb.GetSessionLogPath(),
		RetentionDays:  int(pb.GetRetentionDays()),
		AutoCleanup:    pb.GetAutoCleanup(),
		MaxMemories:    int(pb.GetMaxMemories()),
		Consolidation:  consolidationFromProtobuf(pb.GetConsolidation()),
	}
}

func consolidationToProtobuf(c *ConsolidationConfig) *configpb.ConsolidationConfig {
	if c == nil {
		return nil
	}
	return &configpb.ConsolidationConfig{
		Enabled:            c.Enabled,
		SimilarityThreshold: c.SimilarityThreshold,
		Frequency:          c.Frequency,
	}
}

func consolidationFromProtobuf(pb *configpb.ConsolidationConfig) ConsolidationConfig {
	if pb == nil {
		return ConsolidationConfig{}
	}
	return ConsolidationConfig{
		Enabled:            pb.GetEnabled(),
		SimilarityThreshold: pb.GetSimilarityThreshold(),
		Frequency:          pb.GetFrequency(),
	}
}

func projectToProtobuf(p *ProjectConfig) *configpb.ProjectConfig {
	if p == nil {
		return nil
	}
	return &configpb.ProjectConfig{
		Name:        p.Name,
		Type:        p.Type,
		Language:    p.Language,
		Root:        p.Root,
		Todo2Path:   p.Todo2Path,
		ExarpPath:   p.ExarpPath,
		Features:    featuresToProtobuf(&p.Features),
		SkipChecks:  p.SkipChecks,
		CustomTools: p.CustomTools,
	}
}

func projectFromProtobuf(pb *configpb.ProjectConfig) ProjectConfig {
	if pb == nil {
		return ProjectConfig{}
	}
	return ProjectConfig{
		Name:        pb.GetName(),
		Type:        pb.GetType(),
		Language:    pb.GetLanguage(),
		Root:        pb.GetRoot(),
		Todo2Path:   pb.GetTodo2Path(),
		ExarpPath:   pb.GetExarpPath(),
		Features:    featuresFromProtobuf(pb.GetFeatures()),
		SkipChecks:  pb.GetSkipChecks(),
		CustomTools: pb.GetCustomTools(),
	}
}

func featuresToProtobuf(f *FeaturesConfig) *configpb.FeaturesConfig {
	if f == nil {
		return nil
	}
	return &configpb.FeaturesConfig{
		SqliteEnabled: f.SQLiteEnabled,
		JsonFallback:  f.JSONFallback,
		PythonBridge:  f.PythonBridge,
		McpServers:    f.MCPServers,
	}
}

func featuresFromProtobuf(pb *configpb.FeaturesConfig) FeaturesConfig {
	if pb == nil {
		return FeaturesConfig{}
	}
	return FeaturesConfig{
		SQLiteEnabled: pb.GetSqliteEnabled(),
		JSONFallback:  pb.GetJsonFallback(),
		PythonBridge:  pb.GetPythonBridge(),
		MCPServers:    pb.GetMcpServers(),
	}
}
