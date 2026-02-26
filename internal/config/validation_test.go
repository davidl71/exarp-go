package config

import (
	"testing"
)

func TestValidateLogging(t *testing.T) {
	tests := []struct {
		name    string
		logging LoggingConfig
		wantErr bool
	}{
		{name: "valid defaults", logging: LoggingConfig{}, wantErr: false},
		{name: "valid debug", logging: LoggingConfig{Level: "debug"}, wantErr: false},
		{name: "valid info", logging: LoggingConfig{Level: "info"}, wantErr: false},
		{name: "valid warn", logging: LoggingConfig{Level: "warn"}, wantErr: false},
		{name: "valid error", logging: LoggingConfig{Level: "error"}, wantErr: false},
		{name: "invalid level", logging: LoggingConfig{Level: "trace"}, wantErr: true},
		{name: "valid json format", logging: LoggingConfig{Format: "json"}, wantErr: false},
		{name: "valid text format", logging: LoggingConfig{Format: "text"}, wantErr: false},
		{name: "invalid format", logging: LoggingConfig{Format: "xml"}, wantErr: true},
		{name: "negative retention", logging: LoggingConfig{RetentionDays: -1}, wantErr: true},
		{name: "zero retention", logging: LoggingConfig{RetentionDays: 0}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogging(tt.logging)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_IncludesLogging(t *testing.T) {
	cfg := GetDefaults()
	cfg.Logging.Level = "invalid_level"

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected validation error for invalid logging level")
	}
}
