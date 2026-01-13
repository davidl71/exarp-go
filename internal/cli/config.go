package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"gopkg.in/yaml.v3"
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
	case "help", "--help", "-h":
		return printConfigHelp()
	default:
		return fmt.Errorf("unknown config subcommand: %s (use: init, validate, show, set, help)", subcommand)
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

	fmt.Printf("✅ Config file is valid\n")
	fmt.Printf("   Version: %s\n", cfg.Version)
	fmt.Printf("   Config loaded from: %s/.exarp/config.yaml\n", projectRoot)
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

	fmt.Printf("⚠️  Config set command is not yet fully implemented\n")
	fmt.Printf("   Key: %s\n", key)
	fmt.Printf("   Value: %s\n", value)
	fmt.Printf("   For now, edit %s/.exarp/config.yaml directly\n", projectRoot)
	fmt.Printf("   Then run 'exarp-go config validate' to verify\n")
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
  set <key>=<value> Set a config value (not yet fully implemented)
  help              Show this help message

Examples:
  exarp-go config init
  exarp-go config validate
  exarp-go config show
  exarp-go config show json
  exarp-go config set timeouts.task_lock_lease=45m

Configuration File:
  Location: .exarp/config.yaml (in project root)
  Format: YAML
  Defaults: All defaults match current hard-coded behavior

For more information, see:
  docs/CONFIGURATION_IMPLEMENTATION_PLAN.md
  docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md
`
	fmt.Print(help)
	return nil
}
