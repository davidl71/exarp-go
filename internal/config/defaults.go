// defaults.go — Default configuration values for all config sections.
package config

import (
	"time"
)

// GetDefaults returns a FullConfig with all default values matching current hard-coded behavior
// This ensures backward compatibility when no config file is present.
func GetDefaults() *FullConfig {
	return &FullConfig{
		Version: "1.0",
		Timeouts: TimeoutsConfig{
			// Task management - matches current hard-coded values
			TaskLockLease:      30 * time.Minute,
			TaskLockRenewal:    20 * time.Minute,
			StaleLockThreshold: 5 * time.Minute,

			// Tool execution - matches current hard-coded values
			ToolDefault:   60 * time.Second,
			ToolScorecard: 60 * time.Second,
			ToolLinting:   60 * time.Second,
			ToolTesting:   300 * time.Second,
			ToolReport:    60 * time.Second,

			// External services
			OllamaDownload: 600 * time.Second, // 10 minutes for model downloads
			OllamaGenerate: 300 * time.Second,
			HTTPClient:     30 * time.Second,
			DatabaseRetry:  60 * time.Second,

			// Context management
			ContextSummarize: 30 * time.Second,
			ContextBudget:    10 * time.Second,
		},
		Thresholds: ThresholdsConfig{
			// Task analysis - matches current hard-coded values
			SimilarityThreshold:  0.85,
			MinDescriptionLength: 50,
			MinTaskConfidence:    0.7,

			// Testing - matches current hard-coded values
			MinCoverage:       80,
			MinTestConfidence: 0.7,

			// Estimation
			MinEstimationConfidence: 0.7,
			MLXWeight:               0.3,

			// Automation - matches current hard-coded values
			MaxParallelTasks:        10,
			MaxTasksPerHost:         5,
			MaxTasksPerWave:         0, // 0 = no limit; cap tasks per wave when > 0
			MaxAutomationIterations: 10,

			// Context management - matches current hard-coded values
			TokensPerChar:             0.25,
			DefaultContextBudget:      4000,
			ContextReductionThreshold: 0.5,

			// Security - matches current hard-coded values
			RateLimitRequests: 100,
			RateLimitWindow:   1 * time.Minute,
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			MaxPathDepth:      20,
		},
		Tasks: TasksConfig{
			// Defaults - matches current hard-coded values
			DefaultStatus:   "Todo",
			DefaultPriority: "medium",
			DefaultTags:     []string{},

			// Status transitions
			StatusWorkflow: map[string][]string{
				"Todo":        {"In Progress", "Review"},
				"In Progress": {"Done", "Review", "Todo"},
				"Review":      {"Todo", "In Progress", "Done"},
				"Done":        {}, // Terminal state
			},

			// Cleanup - matches current hard-coded values
			StaleThresholdHours: 2,
			AutoCleanupEnabled:  false,
			CleanupDryRun:       true,

			// Task ID - matches current format
			IDFormat: "T-{epoch_milliseconds}",
			IDPrefix: "T-",

			// Clarity checks - matches current hard-coded values
			MinDescriptionLength: 50,
			RequireDescription:   false,
			AutoClarify:          false,
		},
		Database: DatabaseConfig{
			SQLitePath:          ".todo2/todo2.db",
			JSONFallbackPath:    ".todo2/state.todo2.json",
			BackupPath:          ".todo2/backups",
			MaxConnections:      10,
			ConnectionTimeout:   30 * time.Second,
			QueryTimeout:        60 * time.Second,
			RetryAttempts:       3,
			RetryInitialDelay:   100 * time.Millisecond,
			RetryMaxDelay:       5 * time.Second,
			RetryMultiplier:     2.0,
			AutoVacuum:          true,
			WALMode:             true,
			CheckpointInterval:  1000,
			BackupRetentionDays: 30,
		},
		Security: SecurityConfig{
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerWindow: 100,
				WindowDuration:    1 * time.Minute,
				BurstSize:         10,
			},
			PathValidation: PathValidationConfig{
				Enabled:            true,
				AllowAbsolutePaths: false,
				MaxDepth:           20,
				BlockedPatterns: []string{
					"**/.git/**",
					"**/node_modules/**",
					"**/vendor/**",
				},
			},
			FileLimits: FileLimitsConfig{
				MaxFileSize:          10 * 1024 * 1024, // 10MB
				MaxFilesPerOperation: 1000,
				AllowedExtensions:    []string{}, // Empty = all allowed
			},
			AccessControl: AccessControlConfig{
				Enabled:         false,
				DefaultPolicy:   "allow",
				RestrictedTools: []string{},
			},
		},
		Logging: LoggingConfig{
			Level:             "info",
			ToolLevel:         "info",
			FrameworkLevel:    "warn",
			Format:            "json",
			IncludeTimestamps: true,
			IncludeCaller:     false,
			ColorOutput:       true,
			LogDir:            ".exarp/logs",
			LogFile:           "exarp.log",
			SessionLogDir:     ".exarp/logs/sessions",
			LogRotation: LogRotationConfig{
				Enabled:  true,
				MaxSize:  10 * 1024 * 1024, // 10MB
				MaxFiles: 10,
				Compress: true,
			},
			RetentionDays: 30,
			AutoCleanup:   true,
		},
		Tools: ToolsConfig{
			Scorecard: ScorecardConfig{
				DefaultScores: map[string]float64{
					"security":      50.0,
					"testing":       50.0,
					"documentation": 50.0,
					"completion":    50.0,
					"alignment":     50.0,
					"clarity":       50.0,
					"cicd":          50.0,
					"dogfooding":    50.0,
				},
				IncludeWisdom: true,
				OutputFormat:  "text",
			},
			Report: ReportConfig{
				DefaultFormat:          "text",
				DefaultOutputPath:      "",
				IncludeMetrics:         true,
				IncludeRecommendations: true,
			},
			Linting: LintingConfig{
				DefaultLinter: "auto",
				AutoFix:       false,
				IncludeHints:  true,
				Timeout:       60 * time.Second,
			},
			Testing: TestingConfig{
				DefaultFramework: "auto",
				MinCoverage:      80,
				CoverageFormat:   "html",
				Verbose:          false,
			},
			MLX: MLXConfig{
				DefaultModel:       "mlx-community/Phi-3.5-mini-instruct-4bit",
				DefaultMaxTokens:   512,
				DefaultTemperature: 0.7,
				Verbose:            false,
			},
			Ollama: OllamaConfig{
				DefaultModel:       "llama3.2",
				DefaultHost:        "http://localhost:11434",
				DefaultContextSize: 4096,
				DefaultNumThreads:  4,
				DefaultNumGPU:      1,
			},
			Context: ContextConfig{
				DefaultBudget:             4000,
				DefaultSummarizationLevel: "brief",
				TokensPerChar:             0.25,
				IncludeRaw:                false,
			},
		},
		Workflow: WorkflowConfig{
			DefaultMode:    "development",
			AutoDetectMode: true,
			ModeSuggestions: ModeSuggestionsConfig{
				Morning:   "daily_checkin",
				Afternoon: "development",
				Evening:   "review",
			},
			Modes: map[string]ModeConfig{
				"development": {
					EnabledTools:  []string{"*"}, // All tools
					DisabledTools: []string{},
					ToolLimit:     0, // No limit
				},
				"security_review": {
					EnabledTools:  []string{"security", "health", "linting"},
					DisabledTools: []string{"*"}, // Disable all others
					ToolLimit:     10,
				},
				"task_management": {
					EnabledTools:  []string{"task_*", "automation", "session"},
					DisabledTools: []string{},
					ToolLimit:     15,
				},
			},
			Focus: FocusConfig{
				Enabled:           true,
				ReductionTarget:   70,
				PreserveCoreTools: true,
			},
		},
		Memory: MemoryConfig{
			Categories: []string{
				"debug",
				"research",
				"architecture",
				"preference",
				"insight",
			},
			StoragePath:    ".exarp/memories",
			SessionLogPath: ".exarp/logs/sessions",
			RetentionDays:  90,
			AutoCleanup:    true,
			MaxMemories:    1000,
			Consolidation: ConsolidationConfig{
				Enabled:             true,
				SimilarityThreshold: 0.85,
				Frequency:           "weekly",
			},
		},
		Project: ProjectConfig{
			Name:      "",        // Auto-detected
			Type:      "library", // library | application | service | tool
			Language:  "go",
			Root:      ".", // Auto-detected
			Todo2Path: ".todo2",
			ExarpPath: ".exarp",
			Features: FeaturesConfig{
				SQLiteEnabled: true,
				JSONFallback:  true,
				PythonBridge:  true,
				MCPServers:    []string{},
			},
			SkipChecks:  []string{},
			CustomTools: []string{},
		},
		Automations: AutomationsConfig{
			// Will be implemented in Phase 5
		},
	}
}
