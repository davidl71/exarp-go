package config

import "fmt"

// ConfigBuilder builds BaseConfig with fluent API
type ConfigBuilder struct {
	config *BaseConfig
}

// NewConfigBuilder creates a new config builder with default values
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &BaseConfig{
			Framework: FrameworkGoSDK,
			Name:      "mcp-server",
			Version:   "1.0.0",
		},
	}
}

// WithFramework sets the framework type
func (b *ConfigBuilder) WithFramework(framework FrameworkType) *ConfigBuilder {
	b.config.Framework = framework
	return b
}

// WithName sets the server name
func (b *ConfigBuilder) WithName(name string) *ConfigBuilder {
	b.config.Name = name
	return b
}

// WithVersion sets the server version
func (b *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
	b.config.Version = version
	return b
}

// Build returns the built configuration
// Returns an error if the configuration is invalid
func (b *ConfigBuilder) Build() (*BaseConfig, error) {
	// Validate framework
	if b.config.Framework != FrameworkGoSDK {
		return nil, &ConfigError{
			Field:   "framework",
			Value:   string(b.config.Framework),
			Message: "unsupported framework",
		}
	}

	// Validate name (non-empty)
	if b.config.Name == "" {
		return nil, &ConfigError{
			Field:   "name",
			Value:   "",
			Message: "server name cannot be empty",
		}
	}

	// Validate version (non-empty)
	if b.config.Version == "" {
		return nil, &ConfigError{
			Field:   "version",
			Value:   "",
			Message: "server version cannot be empty",
		}
	}

	return b.config, nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error: %s (field: %s, value: %q)", e.Message, e.Field, e.Value)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}
