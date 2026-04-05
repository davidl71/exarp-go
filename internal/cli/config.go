// config.go — CLI "config" subcommand: show, set, export, init, convert.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// handleConfigCommand handles the config subcommand using ParseArgs result.
func handleConfigCommand(parsed *mcpcli.Args) error {
	subcommand := parsed.Subcommand
	if subcommand == "" && len(parsed.Positional) > 0 {
		subcommand = parsed.Positional[0]
	}

	if subcommand == "" {
		return printConfigHelp()
	}

	switch subcommand {
	case "init":
		return handleConfigInit(parsed.Positional)
	case "validate":
		return handleConfigValidate(parsed.Positional)
	case "show":
		// Format from positional (e.g. "show yaml") or flag
		formatArgs := parsed.Positional
		if format := parsed.GetFlag("format", ""); format != "" {
			formatArgs = []string{format}
		}

		return handleConfigShow(formatArgs)
	case "set":
		return handleConfigSet(parsed.Positional)
	case "get":
		return handleConfigGet(parsed.Positional)
	case "reset":
		return handleConfigReset(parsed.Positional)
	case "diff":
		return handleConfigDiff(parsed.Positional)
	case "history":
		return handleConfigHistory(parsed.Positional)
	case "template":
		return handleConfigTemplate(parsed.Positional)
	case "reload":
		return handleConfigReload()
	case "export":
		formatArgs := parsed.Positional
		if format := parsed.GetFlag("format", ""); format != "" {
			formatArgs = []string{format}
		}

		return handleConfigExport(formatArgs)
	case "convert":
		return handleConfigConvert(parsed.Positional)
	case "help", "--help", "-h":
		return printConfigHelp()
	default:
		return fmt.Errorf("unknown config subcommand: %s (use: init, validate, show, get, set, reset, diff, history, template, export, convert, help)", subcommand)
	}
}

// handleConfigInit generates a default config file in protobuf format (.exarp/config.pb).
func handleConfigInit(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	configPath := filepath.Join(projectRoot, ".exarp", "config.pb")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Config file already exists: %s\n", configPath)
		fmt.Printf("   Use 'exarp-go config show' to view current config\n")
		fmt.Printf("   Or delete the file and run 'init' again\n")

		return nil
	}

	defaults := config.GetDefaults()
	if err := config.WriteConfigToProtobufFile(projectRoot, defaults); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Created default config file: %s\n", configPath)
	fmt.Printf("   Use 'exarp-go config export yaml' to edit as YAML, then 'exarp-go config convert yaml protobuf' to save\n")

	return nil
}

// handleConfigValidate validates the config file.
func handleConfigValidate(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		fmt.Printf("❌ Config validation failed:\n")
		fmt.Printf("   %v\n", err)

		return err
	}

	// Detect format
	format, err := config.GetConfigFormat(projectRoot)
	if err != nil {
		format = "unknown"
	}

	fmt.Printf("✅ Config file is valid\n")
	fmt.Printf("   Version: %s\n", cfg.Version)
	fmt.Printf("   Config format: %s\n", format)

	if format == "protobuf" {
		fmt.Printf("   Config loaded from: %s/.exarp/config.pb\n", projectRoot)
	} else {
		fmt.Printf("   Config: defaults (no .exarp/config.pb)\n")
	}

	return nil
}

// handleConfigShow displays the current configuration.
func handleConfigShow(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine output format
	format := "yaml"
	if len(args) > 0 {
		format = args[0]
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Print(string(data))
	default:
		return fmt.Errorf("unknown format: %s (use: yaml, json)", format)
	}

	return nil
}

// handleConfigSet sets a config value (simple key=value format).
func handleConfigSet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: exarp-go config set <key>=<value>")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Parse key=value
	parts := strings.SplitN(args[0], "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: use key=value (e.g., timeouts.task_lock_lease=45m)")
	}

	key := parts[0]
	value := parts[1]

	// Validate the value against schema
	if err := validateConfigValue(key, value); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Load current config
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set the value based on key path
	if err := setConfigValue(cfg, key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	configPath := filepath.Join(projectRoot, ".exarp", "config.pb")

	fmt.Printf("✅ Set %s = %s\n", key, value)
	fmt.Printf("   Config saved to: %s\n", configPath)
	fmt.Printf("   Run 'exarp-go config validate' to verify\n")

	return nil
}

// validateConfigValue validates a config value against schema rules.
func validateConfigValue(keyPath, value string) error {
	keys := strings.Split(keyPath, ".")
	if len(keys) == 0 {
		return fmt.Errorf("invalid key path: %s", keyPath)
	}

	switch keys[0] {
	case "version":
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("version cannot be empty")
		}
	case "timeouts":
		return validateTimeoutValue(keys[1:], value)
	case "thresholds":
		return validateThresholdValue(keys[1:], value)
	case "tasks":
		return validateTaskValue(keys[1:], value)
	case "project":
		return validateProjectValue(keys[1:], value)
	case "database":
		return validateDatabaseValue(keys[1:], value)
	case "logging":
		return validateLoggingValue(keys[1:], value)
	default:
		// Allow unknown sections for future expansion
	}

	return nil
}

func validateTimeoutValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("timeout keys must be one level deep")
	}

	_, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w (use format like 30m, 1h, 45s)", value, err)
	}

	return nil
}

func validateThresholdValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("threshold keys must be one level deep")
	}

	switch keys[0] {
	case "similarity_threshold", "min_task_confidence", "min_test_confidence":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", value, err)
		}
		if f < 0 || f > 1 {
			return fmt.Errorf("value must be between 0 and 1")
		}
	case "min_coverage", "min_description_length":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		if n < 0 {
			return fmt.Errorf("value must be non-negative")
		}
		if keys[0] == "min_coverage" && n > 100 {
			return fmt.Errorf("min_coverage must be between 0 and 100")
		}
	}

	return nil
}

func validateTaskValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("task keys must be one level deep")
	}

	switch keys[0] {
	case "default_status":
		validStatuses := map[string]bool{"Todo": true, "In Progress": true, "Review": true, "Done": true, "Cancelled": true}
		if !validStatuses[value] {
			return fmt.Errorf("invalid status %q (valid: Todo, In Progress, Review, Done, Cancelled)", value)
		}
	case "default_priority":
		validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
		if !validPriorities[value] {
			return fmt.Errorf("invalid priority %q (valid: high, medium, low)", value)
		}
	}

	return nil
}

func validateProjectValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("project keys must be one level deep")
	}

	// Currently project values are mostly strings or string arrays
	return nil
}

func validateDatabaseValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("database keys must be one level deep")
	}

	switch keys[0] {
	case "max_connections", "retry_attempts", "checkpoint_interval", "backup_retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		if n < 1 {
			return fmt.Errorf("value must be positive")
		}
	case "connection_timeout", "query_timeout", "retry_initial_delay", "retry_max_delay":
		_, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", value, err)
		}
	case "auto_vacuum", "wal_mode":
		if value != "true" && value != "false" {
			return fmt.Errorf("invalid boolean value %q (use: true, false)", value)
		}
	}

	return nil
}

func validateLoggingValue(keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("logging keys must be one level deep")
	}

	switch keys[0] {
	case "level":
		validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if !validLevels[value] {
			return fmt.Errorf("invalid log level %q (valid: debug, info, warn, error)", value)
		}
	case "format":
		validFormats := map[string]bool{"json": true, "text": true}
		if !validFormats[value] {
			return fmt.Errorf("invalid format %q (valid: json, text)", value)
		}
	case "color_output", "include_timestamps", "include_caller", "auto_cleanup":
		if value != "true" && value != "false" {
			return fmt.Errorf("invalid boolean value %q (use: true, false)", value)
		}
	case "retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		if n < 1 {
			return fmt.Errorf("retention_days must be positive")
		}
	}

	return nil
}

// handleConfigGet gets a config value by key path.
func handleConfigGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: exarp-go config get <key>")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	keyPath := args[0]

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	value, err := getConfigValue(cfg, keyPath)
	if err != nil {
		return fmt.Errorf("failed to get config value: %w", err)
	}

	fmt.Printf("%s\n", value)

	return nil
}

// handleConfigReset resets config values to defaults.
func handleConfigReset(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	defaults := config.GetDefaults()

	// Reset all if no key specified
	if len(args) == 0 || args[0] == "all" {
		cfg = defaults
		configPath := filepath.Join(projectRoot, ".exarp", "config.pb")
		if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Printf("✅ Reset all config to defaults\n")
		fmt.Printf("   Config saved to: %s\n", configPath)
		return nil
	}

	// Reset specific key
	keyPath := args[0]
	if err := resetConfigValue(cfg, defaults, keyPath); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Reset %s to default\n", keyPath)
	fmt.Printf("   Run 'exarp-go config validate' to verify\n")

	return nil
}

// resetConfigValue resets a specific key to its default value.
func resetConfigValue(cfg, defaults *config.FullConfig, keyPath string) error {
	keys := strings.Split(keyPath, ".")
	if len(keys) == 0 {
		return fmt.Errorf("invalid key path: %s", keyPath)
	}

	switch keys[0] {
	case "version":
		cfg.Version = defaults.Version
	case "timeouts":
		return resetTimeoutsValue(&cfg.Timeouts, &defaults.Timeouts, keys[1:])
	case "thresholds":
		return resetThresholdsValue(&cfg.Thresholds, &defaults.Thresholds, keys[1:])
	case "tasks":
		return resetTasksValue(&cfg.Tasks, &defaults.Tasks, keys[1:])
	case "project":
		return resetProjectValue(&cfg.Project, &defaults.Project, keys[1:])
	case "database":
		return resetDatabaseValue(&cfg.Database, &defaults.Database, keys[1:])
	case "security":
		cfg.Security = defaults.Security
	case "logging":
		cfg.Logging = defaults.Logging
	case "tools":
		cfg.Tools = defaults.Tools
	case "workflow":
		cfg.Workflow = defaults.Workflow
	case "memory":
		cfg.Memory = defaults.Memory
	default:
		return fmt.Errorf("unsupported config section: %s", keys[0])
	}
	return nil
}

func resetTimeoutsValue(cfg, defaults *config.TimeoutsConfig, keys []string) error {
	if len(keys) == 0 {
		*cfg = *defaults
		return nil
	}
	switch keys[0] {
	case "task_lock_lease":
		cfg.TaskLockLease = defaults.TaskLockLease
	case "task_lock_renewal":
		cfg.TaskLockRenewal = defaults.TaskLockRenewal
	case "stale_lock_threshold":
		cfg.StaleLockThreshold = defaults.StaleLockThreshold
	case "tool_default":
		cfg.ToolDefault = defaults.ToolDefault
	case "tool_scorecard":
		cfg.ToolScorecard = defaults.ToolScorecard
	case "tool_linting":
		cfg.ToolLinting = defaults.ToolLinting
	case "tool_testing":
		cfg.ToolTesting = defaults.ToolTesting
	case "tool_report":
		cfg.ToolReport = defaults.ToolReport
	case "ollama_download":
		cfg.OllamaDownload = defaults.OllamaDownload
	case "ollama_generate":
		cfg.OllamaGenerate = defaults.OllamaGenerate
	case "http_client":
		cfg.HTTPClient = defaults.HTTPClient
	case "database_retry":
		cfg.DatabaseRetry = defaults.DatabaseRetry
	case "context_summarize":
		cfg.ContextSummarize = defaults.ContextSummarize
	case "context_budget":
		cfg.ContextBudget = defaults.ContextBudget
	default:
		return fmt.Errorf("unknown timeout key: %s", keys[0])
	}
	return nil
}

func resetThresholdsValue(cfg, defaults *config.ThresholdsConfig, keys []string) error {
	if len(keys) == 0 {
		*cfg = *defaults
		return nil
	}
	switch keys[0] {
	case "similarity_threshold":
		cfg.SimilarityThreshold = defaults.SimilarityThreshold
	case "min_coverage":
		cfg.MinCoverage = defaults.MinCoverage
	case "min_task_confidence":
		cfg.MinTaskConfidence = defaults.MinTaskConfidence
	case "min_test_confidence":
		cfg.MinTestConfidence = defaults.MinTestConfidence
	case "min_description_length":
		cfg.MinDescriptionLength = defaults.MinDescriptionLength
	default:
		return fmt.Errorf("unknown threshold key: %s", keys[0])
	}
	return nil
}

func resetTasksValue(cfg, defaults *config.TasksConfig, keys []string) error {
	if len(keys) == 0 {
		*cfg = *defaults
		return nil
	}
	switch keys[0] {
	case "default_status":
		cfg.DefaultStatus = defaults.DefaultStatus
	case "default_priority":
		cfg.DefaultPriority = defaults.DefaultPriority
	default:
		return fmt.Errorf("unknown task key: %s", keys[0])
	}
	return nil
}

func resetProjectValue(cfg, defaults *config.ProjectConfig, keys []string) error {
	if len(keys) == 0 {
		*cfg = *defaults
		return nil
	}
	switch keys[0] {
	case "task_discovery_ignore_paths":
		cfg.TaskDiscoveryIgnorePaths = defaults.TaskDiscoveryIgnorePaths
	default:
		return fmt.Errorf("unknown project key: %s", keys[0])
	}
	return nil
}

func resetDatabaseValue(cfg, defaults *config.DatabaseConfig, keys []string) error {
	if len(keys) == 0 {
		*cfg = *defaults
		return nil
	}
	switch keys[0] {
	case "sqlite_path":
		cfg.SQLitePath = defaults.SQLitePath
	case "json_fallback_path":
		cfg.JSONFallbackPath = defaults.JSONFallbackPath
	case "backup_path":
		cfg.BackupPath = defaults.BackupPath
	case "max_connections":
		cfg.MaxConnections = defaults.MaxConnections
	case "connection_timeout":
		cfg.ConnectionTimeout = defaults.ConnectionTimeout
	case "query_timeout":
		cfg.QueryTimeout = defaults.QueryTimeout
	case "retry_attempts":
		cfg.RetryAttempts = defaults.RetryAttempts
	case "auto_vacuum":
		cfg.AutoVacuum = defaults.AutoVacuum
	case "wal_mode":
		cfg.WALMode = defaults.WALMode
	default:
		return fmt.Errorf("unknown database key: %s", keys[0])
	}
	return nil
}

// handleConfigDiff shows the diff of config changes.
func handleConfigDiff(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	defaults := config.GetDefaults()

	fmt.Println("### Config Diff (current vs defaults)")
	fmt.Println("")

	showDiff("version", cfg.Version, defaults.Version)
	showTimeoutsDiff(&cfg.Timeouts, &defaults.Timeouts)
	showThresholdsDiff(&cfg.Thresholds, &defaults.Thresholds)
	showTasksDiff(&cfg.Tasks, &defaults.Tasks)
	showProjectDiff(&cfg.Project, &defaults.Project)
	showDatabaseDiff(&cfg.Database, &defaults.Database)

	return nil
}

func showDiff(key, current, defaultVal string) {
	if current != defaultVal {
		fmt.Printf("- %s: %s (default)\n", key, defaultVal)
		fmt.Printf("+ %s: %s (current)\n", key, current)
		fmt.Println("")
	}
}

func showTimeoutsDiff(cfg, defaults *config.TimeoutsConfig) {
	diffs := []struct {
		key     string
		current time.Duration
		def     time.Duration
	}{
		{"timeouts.task_lock_lease", cfg.TaskLockLease, defaults.TaskLockLease},
		{"timeouts.task_lock_renewal", cfg.TaskLockRenewal, defaults.TaskLockRenewal},
		{"timeouts.stale_lock_threshold", cfg.StaleLockThreshold, defaults.StaleLockThreshold},
		{"timeouts.tool_default", cfg.ToolDefault, defaults.ToolDefault},
		{"timeouts.tool_scorecard", cfg.ToolScorecard, defaults.ToolScorecard},
		{"timeouts.tool_linting", cfg.ToolLinting, defaults.ToolLinting},
		{"timeouts.tool_testing", cfg.ToolTesting, defaults.ToolTesting},
		{"timeouts.tool_report", cfg.ToolReport, defaults.ToolReport},
		{"timeouts.ollama_download", cfg.OllamaDownload, defaults.OllamaDownload},
		{"timeouts.ollama_generate", cfg.OllamaGenerate, defaults.OllamaGenerate},
		{"timeouts.http_client", cfg.HTTPClient, defaults.HTTPClient},
		{"timeouts.database_retry", cfg.DatabaseRetry, defaults.DatabaseRetry},
		{"timeouts.context_summarize", cfg.ContextSummarize, defaults.ContextSummarize},
		{"timeouts.context_budget", cfg.ContextBudget, defaults.ContextBudget},
	}
	for _, d := range diffs {
		if d.current != d.def {
			fmt.Printf("- %s: %s (default)\n", d.key, d.def)
			fmt.Printf("+ %s: %s (current)\n", d.key, d.current)
		}
	}
	fmt.Println("")
}

func showThresholdsDiff(cfg, defaults *config.ThresholdsConfig) {
	diffs := []struct {
		key     string
		current interface{}
		def     interface{}
	}{
		{"thresholds.similarity_threshold", cfg.SimilarityThreshold, defaults.SimilarityThreshold},
		{"thresholds.min_coverage", cfg.MinCoverage, defaults.MinCoverage},
		{"thresholds.min_task_confidence", cfg.MinTaskConfidence, defaults.MinTaskConfidence},
		{"thresholds.min_test_confidence", cfg.MinTestConfidence, defaults.MinTestConfidence},
		{"thresholds.min_description_length", cfg.MinDescriptionLength, defaults.MinDescriptionLength},
	}
	for _, d := range diffs {
		if d.current != d.def {
			fmt.Printf("- %s: %v (default)\n", d.key, d.def)
			fmt.Printf("+ %s: %v (current)\n", d.key, d.current)
		}
	}
}

func showTasksDiff(cfg, defaults *config.TasksConfig) {
	diffs := []struct {
		key     string
		current string
		def     string
	}{
		{"tasks.default_status", cfg.DefaultStatus, defaults.DefaultStatus},
		{"tasks.default_priority", cfg.DefaultPriority, defaults.DefaultPriority},
	}
	for _, d := range diffs {
		if d.current != d.def {
			fmt.Printf("- %s: %s (default)\n", d.key, d.def)
			fmt.Printf("+ %s: %s (current)\n", d.key, d.current)
		}
	}
}

func showProjectDiff(cfg, defaults *config.ProjectConfig) {
	if len(cfg.TaskDiscoveryIgnorePaths) != len(defaults.TaskDiscoveryIgnorePaths) ||
		(len(cfg.TaskDiscoveryIgnorePaths) > 0 && len(defaults.TaskDiscoveryIgnorePaths) > 0 &&
			strings.Join(cfg.TaskDiscoveryIgnorePaths, ",") != strings.Join(defaults.TaskDiscoveryIgnorePaths, ",")) {
		fmt.Printf("- project.task_discovery_ignore_paths: %v (default)\n", defaults.TaskDiscoveryIgnorePaths)
		fmt.Printf("+ project.task_discovery_ignore_paths: %v (current)\n", cfg.TaskDiscoveryIgnorePaths)
	}
}

func showDatabaseDiff(cfg, defaults *config.DatabaseConfig) {
	diffs := []struct {
		key     string
		current string
		def     string
	}{
		{"database.sqlite_path", cfg.SQLitePath, defaults.SQLitePath},
		{"database.json_fallback_path", cfg.JSONFallbackPath, defaults.JSONFallbackPath},
		{"database.backup_path", cfg.BackupPath, defaults.BackupPath},
	}
	for _, d := range diffs {
		if d.current != d.def {
			fmt.Printf("- %s: %s (default)\n", d.key, d.def)
			fmt.Printf("+ %s: %s (current)\n", d.key, d.current)
		}
	}
}

// handleConfigHistory shows config change history.
func handleConfigHistory(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	historyPath := filepath.Join(projectRoot, ".exarp", "config.history")

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No config history found.")
			return nil
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	// Show last N entries (default 10)
	limit := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &limit)
	}

	lines := strings.Split(string(data), "\n")
	start := len(lines) - limit
	if start < 0 {
		start = 0
	}

	fmt.Println("### Config History (last", limit, "changes)")
	fmt.Println("")
	for i := start; i < len(lines); i++ {
		if lines[i] != "" {
			fmt.Println(lines[i])
		}
	}

	return nil
}

// handleConfigTemplate applies or lists config templates.
func handleConfigTemplate(args []string) error {
	if len(args) < 1 {
		return listConfigTemplates()
	}

	templateName := args[0]

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	templates := getConfigTemplates()
	tpl, ok := templates[templateName]
	if !ok {
		return fmt.Errorf("unknown template: %s (available: dev, prod, minimal)", templateName)
	}

	// Apply template
	if err := applyConfigTemplate(cfg, tpl); err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Applied template: %s\n", templateName)
	fmt.Printf("   Run 'exarp-go config diff' to see changes\n")

	// Record in history
	recordConfigHistory(projectRoot, fmt.Sprintf("Applied template: %s at %s", templateName, time.Now().Format(time.RFC3339)))

	return nil
}

func listConfigTemplates() error {
	templates := getConfigTemplates()

	fmt.Println("### Available Config Templates")
	fmt.Println("")
	for name, tpl := range templates {
		fmt.Printf("- %s: %s\n", name, tpl.description)
	}
	fmt.Println("")
	fmt.Println("Usage: exarp-go config template <name>")

	return nil
}

type configTemplate struct {
	description string
	apply       func(cfg *config.FullConfig) error
}

func getConfigTemplates() map[string]configTemplate {
	return map[string]configTemplate{
		"dev": {
			description: "Development settings (verbose logging, longer timeouts)",
			apply: func(cfg *config.FullConfig) error {
				cfg.Logging.Level = "debug"
				cfg.Logging.ColorOutput = true
				cfg.Timeouts.ToolDefault = 5 * time.Minute
				cfg.Timeouts.ToolScorecard = 10 * time.Minute
				cfg.Thresholds.MinCoverage = 50
				return nil
			},
		},
		"prod": {
			description: "Production settings (minimal logging, shorter timeouts)",
			apply: func(cfg *config.FullConfig) error {
				cfg.Logging.Level = "warn"
				cfg.Logging.ColorOutput = false
				cfg.Timeouts.ToolDefault = 2 * time.Minute
				cfg.Timeouts.ToolScorecard = 5 * time.Minute
				cfg.Thresholds.MinCoverage = 80
				return nil
			},
		},
		"minimal": {
			description: "Minimal settings (fast, low resource usage)",
			apply: func(cfg *config.FullConfig) error {
				cfg.Logging.Level = "error"
				cfg.Logging.ColorOutput = false
				cfg.Timeouts.ToolDefault = 1 * time.Minute
				cfg.Timeouts.ToolScorecard = 2 * time.Minute
				cfg.Timeouts.ToolLinting = 30 * time.Second
				cfg.Timeouts.ToolTesting = 1 * time.Minute
				cfg.Thresholds.MinCoverage = 0
				return nil
			},
		},
	}
}

func applyConfigTemplate(cfg *config.FullConfig, tpl configTemplate) error {
	return tpl.apply(cfg)
}

func recordConfigHistory(projectRoot, entry string) {
	historyPath := filepath.Join(projectRoot, ".exarp", "config.history")

	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.WriteString(entry + "\n")
}

// handleConfigReload reloads the config (validates it).
func handleConfigReload() error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		fmt.Printf("❌ Config reload failed:\n")
		fmt.Printf("   %v\n", err)
		return err
	}

	fmt.Printf("✅ Config reloaded successfully\n")
	fmt.Printf("   Version: %s\n", cfg.Version)

	return nil
}

// getConfigValue gets a config value by key path.
func getConfigValue(cfg *config.FullConfig, keyPath string) (string, error) {
	keys := strings.Split(keyPath, ".")
	if len(keys) == 0 {
		return "", fmt.Errorf("invalid key path: %s", keyPath)
	}

	switch keys[0] {
	case "version":
		if len(keys) != 1 {
			return "", fmt.Errorf("version is a top-level key")
		}
		return cfg.Version, nil
	case "timeouts":
		return getTimeoutsValue(&cfg.Timeouts, keys[1:])
	case "thresholds":
		return getThresholdsValue(&cfg.Thresholds, keys[1:])
	case "tasks":
		return getTasksValue(&cfg.Tasks, keys[1:])
	case "project":
		return getProjectValue(&cfg.Project, keys[1:])
	case "database":
		return getDatabaseValue(&cfg.Database, keys[1:])
	case "security":
		return getSecurityValue(&cfg.Security, keys[1:])
	case "logging":
		return getLoggingValue(&cfg.Logging, keys[1:])
	case "tools":
		return getToolsValue(&cfg.Tools, keys[1:])
	case "workflow":
		return getWorkflowValue(&cfg.Workflow, keys[1:])
	case "memory":
		return getMemoryValue(&cfg.Memory, keys[1:])
	default:
		return "", fmt.Errorf("unsupported config section: %s (supported: version, timeouts, thresholds, tasks, project, database, security, logging, tools, workflow, memory)", keys[0])
	}
}

// setConfigValue sets a config value by key path (e.g., "timeouts.task_lock_lease").
func setConfigValue(cfg *config.FullConfig, keyPath, value string) error {
	keys := strings.Split(keyPath, ".")
	if len(keys) == 0 {
		return fmt.Errorf("invalid key path: %s", keyPath)
	}

	// Handle top-level keys
	switch keys[0] {
	case "version":
		if len(keys) != 1 {
			return fmt.Errorf("version is a top-level key")
		}

		cfg.Version = value

		return nil
	case "timeouts":
		return setTimeoutsValue(&cfg.Timeouts, keys[1:], value)
	case "thresholds":
		return setThresholdsValue(&cfg.Thresholds, keys[1:], value)
	case "tasks":
		return setTasksValue(&cfg.Tasks, keys[1:], value)
	case "project":
		return setProjectValue(&cfg.Project, keys[1:], value)
	case "database":
		return setDatabaseValue(&cfg.Database, keys[1:], value)
	case "security":
		return nil // Security is complex, skip for now
	case "logging":
		return setLoggingValue(&cfg.Logging, keys[1:], value)
	case "tools":
		return nil // Tools is complex, skip for now
	case "workflow":
		return setWorkflowValue(&cfg.Workflow, keys[1:], value)
	case "memory":
		return setMemoryValue(&cfg.Memory, keys[1:], value)
	default:
		return fmt.Errorf("unsupported config section: %s (supported: version, timeouts, thresholds, tasks, project, database, logging, workflow, memory)", keys[0])
	}
}

// setTimeoutsValue sets a timeout value.
func setTimeoutsValue(timeouts *config.TimeoutsConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("timeout keys must be one level deep (e.g., timeouts.task_lock_lease)")
	}

	// Parse duration value
	duration, err := parseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration value %q: %w (use format like 30m, 1h, 45s)", value, err)
	}

	switch keys[0] {
	case "task_lock_lease":
		timeouts.TaskLockLease = duration
	case "task_lock_renewal":
		timeouts.TaskLockRenewal = duration
	case "stale_lock_threshold":
		timeouts.StaleLockThreshold = duration
	case "tool_default":
		timeouts.ToolDefault = duration
	case "tool_scorecard":
		timeouts.ToolScorecard = duration
	case "tool_linting":
		timeouts.ToolLinting = duration
	case "tool_testing":
		timeouts.ToolTesting = duration
	case "tool_report":
		timeouts.ToolReport = duration
	case "ollama_download":
		timeouts.OllamaDownload = duration
	case "ollama_generate":
		timeouts.OllamaGenerate = duration
	case "http_client":
		timeouts.HTTPClient = duration
	case "database_retry":
		timeouts.DatabaseRetry = duration
	case "context_summarize":
		timeouts.ContextSummarize = duration
	case "context_budget":
		timeouts.ContextBudget = duration
	default:
		return fmt.Errorf("unknown timeout key: %s", keys[0])
	}

	return nil
}

// setThresholdsValue sets a threshold value.
func setThresholdsValue(thresholds *config.ThresholdsConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("threshold keys must be one level deep (e.g., thresholds.similarity_threshold)")
	}

	// Parse float value
	floatVal, err := parseFloat(value)
	if err != nil {
		return fmt.Errorf("invalid float value %q: %w", value, err)
	}

	switch keys[0] {
	case "similarity_threshold":
		thresholds.SimilarityThreshold = floatVal
	case "min_coverage":
		thresholds.MinCoverage = int(floatVal)
	case "min_task_confidence":
		thresholds.MinTaskConfidence = floatVal
	case "min_test_confidence":
		thresholds.MinTestConfidence = floatVal
	case "min_description_length":
		thresholds.MinDescriptionLength = int(floatVal)
	default:
		return fmt.Errorf("unknown threshold key: %s", keys[0])
	}

	return nil
}

// setTasksValue sets a task config value.
func setTasksValue(tasks *config.TasksConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("task keys must be one level deep (e.g., tasks.default_status)")
	}

	switch keys[0] {
	case "default_status":
		tasks.DefaultStatus = value
	case "default_priority":
		tasks.DefaultPriority = value
	default:
		return fmt.Errorf("unknown task key: %s (supported: default_status, default_priority)", keys[0])
	}

	return nil
}

func setProjectValue(project *config.ProjectConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("project keys must be one level deep (e.g., project.task_discovery_ignore_paths)")
	}

	switch keys[0] {
	case "task_discovery_ignore_paths":
		project.TaskDiscoveryIgnorePaths = normalizeCSVList(value)
	case "name":
		project.Name = value
	case "type":
		project.Type = value
	case "language":
		project.Language = value
	default:
		return fmt.Errorf("unknown project key: %s", keys[0])
	}

	return nil
}

func setDatabaseValue(database *config.DatabaseConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("database keys must be one level deep (e.g., database.sqlite_path)")
	}

	switch keys[0] {
	case "sqlite_path":
		database.SQLitePath = value
	case "json_fallback_path":
		database.JSONFallbackPath = value
	case "backup_path":
		database.BackupPath = value
	case "max_connections":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		database.MaxConnections = n
	case "connection_timeout":
		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		database.ConnectionTimeout = d
	case "query_timeout":
		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		database.QueryTimeout = d
	case "retry_attempts":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		database.RetryAttempts = n
	case "auto_vacuum":
		database.AutoVacuum = value == "true"
	case "wal_mode":
		database.WALMode = value == "true"
	default:
		return fmt.Errorf("unknown database key: %s", keys[0])
	}

	return nil
}

func setLoggingValue(logging *config.LoggingConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("logging keys must be one level deep (e.g., logging.level)")
	}

	switch keys[0] {
	case "level":
		logging.Level = value
	case "format":
		logging.Format = value
	case "log_dir":
		logging.LogDir = value
	case "log_file":
		logging.LogFile = value
	case "color_output":
		logging.ColorOutput = value == "true"
	case "include_timestamps":
		logging.IncludeTimestamps = value == "true"
	case "include_caller":
		logging.IncludeCaller = value == "true"
	case "retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		logging.RetentionDays = n
	case "auto_cleanup":
		logging.AutoCleanup = value == "true"
	default:
		return fmt.Errorf("unknown logging key: %s", keys[0])
	}

	return nil
}

func setWorkflowValue(workflow *config.WorkflowConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("workflow keys must be one level deep (e.g., workflow.default_mode)")
	}

	switch keys[0] {
	case "default_mode":
		workflow.DefaultMode = value
	case "auto_detect_mode":
		workflow.AutoDetectMode = value == "true"
	default:
		return fmt.Errorf("unknown workflow key: %s", keys[0])
	}

	return nil
}

func setMemoryValue(memory *config.MemoryConfig, keys []string, value string) error {
	if len(keys) != 1 {
		return fmt.Errorf("memory keys must be one level deep (e.g., memory.storage_path)")
	}

	switch keys[0] {
	case "storage_path":
		memory.StoragePath = value
	case "session_log_path":
		memory.SessionLogPath = value
	case "retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		memory.RetentionDays = n
	case "auto_cleanup":
		memory.AutoCleanup = value == "true"
	case "max_memories":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		memory.MaxMemories = n
	default:
		return fmt.Errorf("unknown memory key: %s", keys[0])
	}

	return nil
}

func normalizeCSVList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		result = append(result, part)
	}

	return result
}

func getTimeoutsValue(timeouts *config.TimeoutsConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("timeout keys must be one level deep (e.g., timeouts.task_lock_lease)")
	}

	switch keys[0] {
	case "task_lock_lease":
		return timeouts.TaskLockLease.String(), nil
	case "task_lock_renewal":
		return timeouts.TaskLockRenewal.String(), nil
	case "stale_lock_threshold":
		return timeouts.StaleLockThreshold.String(), nil
	case "tool_default":
		return timeouts.ToolDefault.String(), nil
	case "tool_scorecard":
		return timeouts.ToolScorecard.String(), nil
	case "tool_linting":
		return timeouts.ToolLinting.String(), nil
	case "tool_testing":
		return timeouts.ToolTesting.String(), nil
	case "tool_report":
		return timeouts.ToolReport.String(), nil
	case "ollama_download":
		return timeouts.OllamaDownload.String(), nil
	case "ollama_generate":
		return timeouts.OllamaGenerate.String(), nil
	case "http_client":
		return timeouts.HTTPClient.String(), nil
	case "database_retry":
		return timeouts.DatabaseRetry.String(), nil
	case "context_summarize":
		return timeouts.ContextSummarize.String(), nil
	case "context_budget":
		return timeouts.ContextBudget.String(), nil
	default:
		return "", fmt.Errorf("unknown timeout key: %s", keys[0])
	}
}

func getThresholdsValue(thresholds *config.ThresholdsConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("threshold keys must be one level deep (e.g., thresholds.similarity_threshold)")
	}

	switch keys[0] {
	case "similarity_threshold":
		return fmt.Sprintf("%f", thresholds.SimilarityThreshold), nil
	case "min_coverage":
		return fmt.Sprintf("%d", thresholds.MinCoverage), nil
	case "min_task_confidence":
		return fmt.Sprintf("%f", thresholds.MinTaskConfidence), nil
	case "min_test_confidence":
		return fmt.Sprintf("%f", thresholds.MinTestConfidence), nil
	case "min_description_length":
		return fmt.Sprintf("%d", thresholds.MinDescriptionLength), nil
	default:
		return "", fmt.Errorf("unknown threshold key: %s", keys[0])
	}
}

func getTasksValue(tasks *config.TasksConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("task keys must be one level deep (e.g., tasks.default_status)")
	}

	switch keys[0] {
	case "default_status":
		return tasks.DefaultStatus, nil
	case "default_priority":
		return tasks.DefaultPriority, nil
	default:
		return "", fmt.Errorf("unknown task key: %s (supported: default_status, default_priority)", keys[0])
	}
}

func getProjectValue(project *config.ProjectConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("project keys must be one level deep (e.g., project.task_discovery_ignore_paths)")
	}

	switch keys[0] {
	case "task_discovery_ignore_paths":
		return strings.Join(project.TaskDiscoveryIgnorePaths, ","), nil
	case "name":
		return project.Name, nil
	case "type":
		return project.Type, nil
	case "language":
		return project.Language, nil
	case "root":
		return project.Root, nil
	default:
		return "", fmt.Errorf("unknown project key: %s", keys[0])
	}
}

func getDatabaseValue(database *config.DatabaseConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("database keys must be one level deep (e.g., database.sqlite_path)")
	}

	switch keys[0] {
	case "sqlite_path":
		return database.SQLitePath, nil
	case "json_fallback_path":
		return database.JSONFallbackPath, nil
	case "backup_path":
		return database.BackupPath, nil
	case "max_connections":
		return fmt.Sprintf("%d", database.MaxConnections), nil
	case "connection_timeout":
		return database.ConnectionTimeout.String(), nil
	case "query_timeout":
		return database.QueryTimeout.String(), nil
	case "retry_attempts":
		return fmt.Sprintf("%d", database.RetryAttempts), nil
	case "auto_vacuum":
		return fmt.Sprintf("%t", database.AutoVacuum), nil
	case "wal_mode":
		return fmt.Sprintf("%t", database.WALMode), nil
	default:
		return "", fmt.Errorf("unknown database key: %s", keys[0])
	}
}

func getSecurityValue(security *config.SecurityConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("security keys must be one level deep (e.g., security.rate_limit.enabled)")
	}

	switch keys[0] {
	case "rate_limit":
		return fmt.Sprintf("enabled=%t, requests=%d, window=%s",
			security.RateLimit.Enabled,
			security.RateLimit.RequestsPerWindow,
			security.RateLimit.WindowDuration.String()), nil
	case "path_validation":
		return fmt.Sprintf("enabled=%t, max_depth=%d",
			security.PathValidation.Enabled,
			security.PathValidation.MaxDepth), nil
	case "file_limits":
		return fmt.Sprintf("max_file_size=%d", security.FileLimits.MaxFileSize), nil
	case "access_control":
		return fmt.Sprintf("enabled=%t, default_policy=%s",
			security.AccessControl.Enabled,
			security.AccessControl.DefaultPolicy), nil
	default:
		return "", fmt.Errorf("unknown security key: %s (supported: rate_limit, path_validation, file_limits, access_control)", keys[0])
	}
}

func getLoggingValue(logging *config.LoggingConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("logging keys must be one level deep (e.g., logging.level)")
	}

	switch keys[0] {
	case "level":
		return logging.Level, nil
	case "format":
		return logging.Format, nil
	case "log_dir":
		return logging.LogDir, nil
	case "log_file":
		return logging.LogFile, nil
	case "color_output":
		return fmt.Sprintf("%t", logging.ColorOutput), nil
	case "retention_days":
		return fmt.Sprintf("%d", logging.RetentionDays), nil
	default:
		return "", fmt.Errorf("unknown logging key: %s", keys[0])
	}
}

func getToolsValue(tools *config.ToolsConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("tools keys must be one level deep (e.g., tools.ollama.default_model)")
	}

	switch keys[0] {
	case "scorecard":
		return fmt.Sprintf("include_wisdom=%t", tools.Scorecard.IncludeWisdom), nil
	case "report":
		return fmt.Sprintf("default_format=%s", tools.Report.DefaultFormat), nil
	case "linting":
		return fmt.Sprintf("default_linter=%s, auto_fix=%t", tools.Linting.DefaultLinter, tools.Linting.AutoFix), nil
	case "testing":
		return fmt.Sprintf("default_framework=%s, min_coverage=%d", tools.Testing.DefaultFramework, tools.Testing.MinCoverage), nil
	case "ollama":
		return fmt.Sprintf("default_model=%s, default_host=%s", tools.Ollama.DefaultModel, tools.Ollama.DefaultHost), nil
	case "context":
		return fmt.Sprintf("default_budget=%d", tools.Context.DefaultBudget), nil
	default:
		return "", fmt.Errorf("unknown tools key: %s", keys[0])
	}
}

func getWorkflowValue(workflow *config.WorkflowConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("workflow keys must be one level deep (e.g., workflow.default_mode)")
	}

	switch keys[0] {
	case "default_mode":
		return workflow.DefaultMode, nil
	case "auto_detect_mode":
		return fmt.Sprintf("%t", workflow.AutoDetectMode), nil
	default:
		return "", fmt.Errorf("unknown workflow key: %s", keys[0])
	}
}

func getMemoryValue(memory *config.MemoryConfig, keys []string) (string, error) {
	if len(keys) != 1 {
		return "", fmt.Errorf("memory keys must be one level deep (e.g., memory.storage_path)")
	}

	switch keys[0] {
	case "storage_path":
		return memory.StoragePath, nil
	case "session_log_path":
		return memory.SessionLogPath, nil
	case "retention_days":
		return fmt.Sprintf("%d", memory.RetentionDays), nil
	case "auto_cleanup":
		return fmt.Sprintf("%t", memory.AutoCleanup), nil
	case "max_memories":
		return fmt.Sprintf("%d", memory.MaxMemories), nil
	default:
		return "", fmt.Errorf("unknown memory key: %s", keys[0])
	}
}

// parseDuration parses a duration string (e.g., "30m", "1h", "45s").
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// parseFloat parses a float string.
func parseFloat(s string) (float64, error) {
	var f float64

	_, err := fmt.Sscanf(s, "%f", &f)

	return f, err
}

// handleConfigExport exports config to different formats.
func handleConfigExport(args []string) error {
	// Parse format argument (yaml, json, protobuf)
	format := "yaml"
	if len(args) > 0 {
		format = strings.ToLower(args[0])
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Export based on format
	switch format {
	case "protobuf", "pb":
		return exportProtobuf(cfg, projectRoot)
	case "yaml", "yml":
		return exportYAML(cfg, projectRoot)
	case "json":
		return exportJSON(cfg, projectRoot)
	default:
		return fmt.Errorf("unknown format: %s (use: yaml, json, protobuf)", format)
	}
}

// handleConfigConvert converts between config formats.
func handleConfigConvert(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: exarp-go config convert <from> <to>")
	}

	fromFormat := strings.ToLower(args[0])
	toFormat := strings.ToLower(args[1])

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Load from source format
	var cfg *config.FullConfig

	switch fromFormat {
	case "yaml", "yml":
		cfg, err = config.LoadConfig(projectRoot) // Loads YAML (or protobuf if exists)
		if err != nil {
			return fmt.Errorf("failed to load YAML config: %w", err)
		}
	case "protobuf", "pb":
		cfg, err = config.LoadConfigProtobuf(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load protobuf config: %w", err)
		}
	default:
		return fmt.Errorf("unknown source format: %s (use: yaml, protobuf)", fromFormat)
	}

	// Save to target format
	switch toFormat {
	case "yaml", "yml":
		return saveYAML(cfg, projectRoot)
	case "protobuf", "pb":
		return saveProtobuf(cfg, projectRoot)
	default:
		return fmt.Errorf("unknown target format: %s (use: yaml, protobuf)", toFormat)
	}
}

// exportProtobuf exports config as protobuf binary.
func exportProtobuf(cfg *config.FullConfig, projectRoot string) error {
	// Convert to protobuf
	pbConfig, err := config.ToProtobuf(cfg)
	if err != nil {
		return fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	// Marshal to binary
	data, err := proto.Marshal(pbConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Write to file
	outputPath := filepath.Join(projectRoot, ".exarp", "config.pb")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write protobuf config: %w", err)
	}

	fmt.Printf("✅ Exported config to protobuf format: %s\n", outputPath)

	return nil
}

// saveProtobuf saves config as protobuf binary (alias for exportProtobuf).
func saveProtobuf(cfg *config.FullConfig, projectRoot string) error {
	return exportProtobuf(cfg, projectRoot)
}

// exportYAML exports config as YAML.
func exportYAML(cfg *config.FullConfig, projectRoot string) error {
	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	outputPath := filepath.Join(projectRoot, ".exarp", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML config: %w", err)
	}

	fmt.Printf("✅ Exported config to YAML format: %s\n", outputPath)

	return nil
}

// saveYAML saves config as YAML (alias for exportYAML).
func saveYAML(cfg *config.FullConfig, projectRoot string) error {
	return exportYAML(cfg, projectRoot)
}

// exportJSON exports config as JSON.
func exportJSON(cfg *config.FullConfig, projectRoot string) error {
	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	outputPath := filepath.Join(projectRoot, ".exarp", "config.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON config: %w", err)
	}

	fmt.Printf("✅ Exported config to JSON format: %s\n", outputPath)

	return nil
}

// printConfigHelp prints help for config command.
func printConfigHelp() error {
	help := `Configuration Management Commands

Usage: exarp-go config <subcommand> [options]

Subcommands:
  init              Generate default .exarp/config.pb file (protobuf)
  validate          Validate the current config file
  show [format]     Display current configuration (yaml or json)
  get <key>         Get a config value by key path
  set <key>=<value> Set a config value (saves to .exarp/config.pb)
  reset [key]       Reset config to defaults (key or all)
  diff              Show diff between current and defaults
  history [n]       Show config change history (last n entries)
  template [name]   Apply/list config templates (dev, prod, minimal)
  reload            Reload and validate config
  export [format]   Export config to format (yaml, json, protobuf)
  convert <from> <to> Convert config between formats (yaml ↔ protobuf)
  help              Show this help message

Examples:
  exarp-go config init
  exarp-go config validate
  exarp-go config show
  exarp-go config show json
  exarp-go config get timeouts.task_lock_lease
  exarp-go config set timeouts.task_lock_lease=45m
  exarp-go config reset timeouts.task_lock_lease
  exarp-go config reset all
  exarp-go config diff
  exarp-go config history
  exarp-go config template dev
  exarp-go config reload
  exarp-go config export yaml
  exarp-go config convert yaml protobuf
  exarp-go config convert protobuf yaml

Templates:
  dev     Development settings (verbose, longer timeouts)
  prod    Production settings (minimal, shorter timeouts)
  minimal Fast, low resource usage settings

Key Paths:
  version, timeouts.<field>, thresholds.<field>, tasks.<field>
  project.<field>, database.<field>, security.<field>, logging.<field>
  tools.<field>, workflow.<field>, memory.<field>

Configuration File (protobuf mandatory):
  Location: .exarp/config.pb (required for file-based config)
  Format: Protobuf binary. Use 'export yaml' to edit as YAML, then 'convert yaml protobuf' to save.
  Defaults: Run without a file uses in-memory defaults.

For more information, see:
  docs/CONFIGURATION_IMPLEMENTATION_PLAN.md
  docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md
`
	fmt.Print(help)

	return nil
}
