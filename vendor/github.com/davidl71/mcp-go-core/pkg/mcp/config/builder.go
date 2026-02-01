package config

// ConfigBuilder provides a fluent API for building BaseConfig programmatically.
// Complements LoadBaseConfig() for tests and code that constructs config without env.
type ConfigBuilder struct {
	framework string
	name      string
	version   string
}

// NewConfigBuilder returns a new ConfigBuilder with default values.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		framework: DefaultFramework,
		name:      DefaultName,
		version:   DefaultVersion,
	}
}

// WithFramework sets the framework (e.g. "go-sdk").
func (b *ConfigBuilder) WithFramework(framework string) *ConfigBuilder {
	b.framework = framework
	return b
}

// WithName sets the server name.
func (b *ConfigBuilder) WithName(name string) *ConfigBuilder {
	b.name = name
	return b
}

// WithVersion sets the server version.
func (b *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
	b.version = version
	return b
}

// Build returns a BaseConfig and validates it. Returns an error if any required field is empty.
func (b *ConfigBuilder) Build() (*BaseConfig, error) {
	cfg := &BaseConfig{
		Framework: b.framework,
		Name:      b.name,
		Version:   b.version,
	}
	if err := validateBaseConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
