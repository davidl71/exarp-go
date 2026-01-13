// Package config provides configuration management for the wisdom engine.
// It handles loading and saving configuration from files and environment variables.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds wisdom configuration including source selection and Hebrew text options.
type Config struct {
	Source        string `json:"source"`
	HebrewEnabled bool   `json:"hebrew_enabled"`
	HebrewOnly    bool   `json:"hebrew_only"`
	Disabled      bool   `json:"disabled"`
	configPath    string
}

// NewConfig creates a new config with default values.
// Default source is "pistis_sophia" and Hebrew features are disabled.
func NewConfig() *Config {
	return &Config{
		Source:        "pistis_sophia", // Default source
		HebrewEnabled: false,
		HebrewOnly:    false,
		Disabled:      false,
		configPath:    GetConfigPath(),
	}
}

// Load loads configuration from file or environment variables.
// Environment variables take precedence over file configuration.
// Returns nil if config file doesn't exist (it's optional).
func (c *Config) Load() error {
	// First check environment variables
	if source := os.Getenv("EXARP_WISDOM_SOURCE"); source != "" {
		c.Source = source
	}

	if os.Getenv("EXARP_WISDOM_HEBREW") == "1" {
		c.HebrewEnabled = true
	}

	if os.Getenv("EXARP_WISDOM_HEBREW_ONLY") == "1" {
		c.HebrewOnly = true
	}

	if os.Getenv("EXARP_DISABLE_WISDOM") == "1" {
		c.Disabled = true
	}

	// Check for .exarp_no_wisdom marker file
	if _, err := os.Stat(".exarp_no_wisdom"); err == nil {
		c.Disabled = true
	}

	// Try to load from config file
	if _, err := os.Stat(c.configPath); err == nil {
		data, err := os.ReadFile(c.configPath)
		if err == nil {
			if err := json.Unmarshal(data, c); err == nil {
				// Config loaded successfully
				return nil
			}
		}
	}

	return nil // Config file is optional
}

// Save saves configuration to file in JSON format.
// Creates the config directory if it doesn't exist.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON for file %q: %w", c.configPath, err)
	}

	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %q: %w", configDir, err)
	}

	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %q: %w", c.configPath, err)
	}

	return nil
}

// GetConfigPath returns the path to the config file.
// Checks home directory first, then falls back to current directory.
func GetConfigPath() string {
	// Look for .exarp_wisdom_config in current directory
	// TODO: Also check home directory
	home, err := os.UserHomeDir()
	if err == nil {
		// Check home first
		homeConfig := filepath.Join(home, ".exarp_wisdom_config")
		if _, err := os.Stat(homeConfig); err == nil {
			return homeConfig
		}
	}

	// Default to current directory
	return filepath.Join(".", ".exarp_wisdom_config")
}
