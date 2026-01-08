package tools

import (
	"testing"
)

// Test helper functions that don't require Foundation Models API
// These can run on all platforms

func TestGetTemperature(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected float32
	}{
		{
			name:     "temperature provided",
			params:   map[string]interface{}{"temperature": 0.5},
			expected: 0.5,
		},
		{
			name:     "temperature as float64",
			params:   map[string]interface{}{"temperature": float64(0.8)},
			expected: 0.8,
		},
		{
			name:     "no temperature (default)",
			params:   map[string]interface{}{},
			expected: 0.7,
		},
		{
			name:     "zero temperature",
			params:   map[string]interface{}{"temperature": 0.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTemperature(tt.params)
			if result != tt.expected {
				t.Errorf("getTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetMaxTokens(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected int
	}{
		{
			name:     "max_tokens provided",
			params:   map[string]interface{}{"max_tokens": 256},
			expected: 256,
		},
		{
			name:     "max_tokens as float64",
			params:   map[string]interface{}{"max_tokens": float64(512)},
			expected: 512,
		},
		{
			name:     "no max_tokens (default)",
			params:   map[string]interface{}{},
			expected: 512,
		},
		{
			name:     "zero max_tokens",
			params:   map[string]interface{}{"max_tokens": 0},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxTokens(tt.params)
			if result != tt.expected {
				t.Errorf("getMaxTokens() = %v, want %v", result, tt.expected)
			}
		})
	}
}

