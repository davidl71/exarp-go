package config

import (
	"testing"
)

func TestNewConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()
	if builder == nil {
		t.Fatal("NewConfigBuilder() returned nil")
	}
	if builder.config == nil {
		t.Fatal("NewConfigBuilder() config is nil")
	}

	// Check defaults
	if builder.config.Framework != FrameworkGoSDK {
		t.Errorf("NewConfigBuilder().config.Framework = %v, want %v", builder.config.Framework, FrameworkGoSDK)
	}
	if builder.config.Name != "mcp-server" {
		t.Errorf("NewConfigBuilder().config.Name = %q, want %q", builder.config.Name, "mcp-server")
	}
	if builder.config.Version != "1.0.0" {
		t.Errorf("NewConfigBuilder().config.Version = %q, want %q", builder.config.Version, "1.0.0")
	}
}

func TestConfigBuilder_WithFramework(t *testing.T) {
	builder := NewConfigBuilder()
	result := builder.WithFramework(FrameworkGoSDK)

	// Should return self for chaining
	if result != builder {
		t.Error("WithFramework() should return self for chaining")
	}

	// Should update framework
	if builder.config.Framework != FrameworkGoSDK {
		t.Errorf("WithFramework() did not update framework")
	}
}

func TestConfigBuilder_WithName(t *testing.T) {
	builder := NewConfigBuilder()
	result := builder.WithName("custom-server")

	// Should return self for chaining
	if result != builder {
		t.Error("WithName() should return self for chaining")
	}

	// Should update name
	if builder.config.Name != "custom-server" {
		t.Errorf("WithName() did not update name, got %q", builder.config.Name)
	}
}

func TestConfigBuilder_WithVersion(t *testing.T) {
	builder := NewConfigBuilder()
	result := builder.WithVersion("2.0.0")

	// Should return self for chaining
	if result != builder {
		t.Error("WithVersion() should return self for chaining")
	}

	// Should update version
	if builder.config.Version != "2.0.0" {
		t.Errorf("WithVersion() did not update version, got %q", builder.config.Version)
	}
}

func TestConfigBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		builder *ConfigBuilder
		wantErr bool
	}{
		{
			name: "valid config",
			builder: NewConfigBuilder().
				WithName("test-server").
				WithVersion("1.0.0"),
			wantErr: false,
		},
		{
			name: "unsupported framework",
			builder: NewConfigBuilder().
				WithFramework(FrameworkType("unsupported")).
				WithName("test-server").
				WithVersion("1.0.0"),
			wantErr: true,
		},
		{
			name: "empty name",
			builder: NewConfigBuilder().
				WithName("").
				WithVersion("1.0.0"),
			wantErr: true,
		},
		{
			name: "empty version",
			builder: NewConfigBuilder().
				WithName("test-server").
				WithVersion(""),
			wantErr: true,
		},
		{
			name: "fluent API chaining",
			builder: NewConfigBuilder().
				WithFramework(FrameworkGoSDK).
				WithName("fluent-server").
				WithVersion("3.0.0"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := tt.builder.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigBuilder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cfg == nil {
					t.Error("ConfigBuilder.Build() returned nil config")
					return
				}
				// Verify values were set correctly
				if tt.builder.config.Name != "" && cfg.Name != tt.builder.config.Name {
					t.Errorf("ConfigBuilder.Build() config.Name = %q, want %q", cfg.Name, tt.builder.config.Name)
				}
			}
		})
	}
}

func TestConfigBuilder_FluentAPI(t *testing.T) {
	cfg, err := NewConfigBuilder().
		WithFramework(FrameworkGoSDK).
		WithName("fluent-test").
		WithVersion("2.5.0").
		Build()

	if err != nil {
		t.Fatalf("ConfigBuilder.Build() error = %v", err)
	}

	if cfg.Framework != FrameworkGoSDK {
		t.Errorf("config.Framework = %v, want %v", cfg.Framework, FrameworkGoSDK)
	}
	if cfg.Name != "fluent-test" {
		t.Errorf("config.Name = %q, want %q", cfg.Name, "fluent-test")
	}
	if cfg.Version != "2.5.0" {
		t.Errorf("config.Version = %q, want %q", cfg.Version, "2.5.0")
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "framework",
		Value:   "invalid",
		Message: "unsupported framework",
	}

	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("ConfigError.Error() returned empty string")
	}
	if errorMsg == "unsupported framework" {
		t.Error("ConfigError.Error() should include field and value information")
	}
}
