package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"gopkg.in/yaml.v3"
	"google.golang.org/protobuf/proto"
)

// handleConfigCommand handles the config subcommand
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		return printConfigHelp()
	}

	subcommand := args[0]
	switch subcommand {
	case "init":
		return handleConfigInit(args[1:])
	case "validate":
		return handleConfigValidate(args[1:])
	case "show":
		return handleConfigShow(args[1:])
	case "set":
		return handleConfigSet(args[1:])
	case "export":
		return handleConfigExport(args[1:])
	case "convert":
		return handleConfigConvert(args[1:])
	case "help", "--help", "-h":
		return printConfigHelp()
	default:
		return fmt.Errorf("unknown config subcommand: %s (use: init, validate, show, set, export, convert, help)", subcommand)
	}
}

// handleConfigInit generates a default config file
func handleConfigInit(args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	configDir := filepath.Join(projectRoot, ".exarp")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Config file already exists: %s\n", configPath)
		fmt.Printf("   Use 'exarp-go config show' to view current config\n")
		fmt.Printf("   Or delete the file and run 'init' again\n")
		return nil
	}

	// Create .exarp directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Get defaults and marshal to YAML
	defaults := config.GetDefaults()
	data, err := yaml.Marshal(defaults)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Created default config file: %s\n", configPath)
	fmt.Printf("   Edit this file to customize configuration for your project\n")
	return nil
}

// handleConfigValidate validates the config file
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
		fmt.Printf("   Config loaded from: %s/.exarp/config.yaml\n", projectRoot)
	}
	return nil
}

// handleConfigShow displays the current configuration
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

// handleConfigSet sets a config value (simple key=value format)
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

	// Load current config
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set the value based on key path
	if err := setConfigValue(cfg, key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Save back to YAML
	configPath := filepath.Join(projectRoot, ".exarp", "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Set %s = %s\n", key, value)
	fmt.Printf("   Config saved to: %s\n", configPath)
	fmt.Printf("   Run 'exarp-go config validate' to verify\n")
	return nil
}

// setConfigValue sets a config value by key path (e.g., "timeouts.task_lock_lease")
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
	default:
		return fmt.Errorf("unsupported config section: %s (supported: version, timeouts, thresholds, tasks)", keys[0])
	}
}

// setTimeoutsValue sets a timeout value
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

// setThresholdsValue sets a threshold value
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

// setTasksValue sets a task config value
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

// parseDuration parses a duration string (e.g., "30m", "1h", "45s")
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// parseFloat parses a float string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// handleConfigExport exports config to different formats
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

// handleConfigConvert converts between config formats
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

// exportProtobuf exports config as protobuf binary
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

// saveProtobuf saves config as protobuf binary (alias for exportProtobuf)
func saveProtobuf(cfg *config.FullConfig, projectRoot string) error {
	return exportProtobuf(cfg, projectRoot)
}

// exportYAML exports config as YAML
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

// saveYAML saves config as YAML (alias for exportYAML)
func saveYAML(cfg *config.FullConfig, projectRoot string) error {
	return exportYAML(cfg, projectRoot)
}

// exportJSON exports config as JSON
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

// printConfigHelp prints help for config command
func printConfigHelp() error {
	help := `Configuration Management Commands

Usage: exarp-go config <subcommand> [options]

Subcommands:
  init              Generate default .exarp/config.yaml file
  validate          Validate the current config file
  show [format]     Display current configuration (yaml or json)
  set <key>=<value> Set a config value
  export [format]   Export config to format (yaml, json, protobuf)
  convert <from> <to> Convert config between formats (yaml ↔ protobuf)
  help              Show this help message

Examples:
  exarp-go config init
  exarp-go config validate
  exarp-go config show
  exarp-go config show json
  exarp-go config export protobuf
  exarp-go config convert yaml protobuf
  exarp-go config convert protobuf yaml
  exarp-go config set timeouts.task_lock_lease=45m

Configuration File:
  Location: .exarp/config.yaml (primary) or .exarp/config.pb (optional)
  Format: YAML (human-readable) or Protobuf Binary (type-safe)
  Defaults: All defaults match current hard-coded behavior

For more information, see:
  docs/CONFIGURATION_IMPLEMENTATION_PLAN.md
  docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md
`
	fmt.Print(help)
	return nil
}
