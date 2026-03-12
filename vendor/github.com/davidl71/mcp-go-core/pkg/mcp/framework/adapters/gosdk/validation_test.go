package gosdk

import (
	"testing"
)

func TestValidateToolRegistration(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		description string
		handler     interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid registration",
			toolName:    "test_tool",
			description: "Test tool description",
			handler:     func() {},
			wantErr:     false,
		},
		{
			name:        "empty name",
			toolName:    "",
			description: "Test description",
			handler:     func() {},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "empty description",
			toolName:    "test_tool",
			description: "",
			handler:     func() {},
			wantErr:     true,
			errContains: "description cannot be empty",
		},
		{
			name:        "nil handler",
			toolName:    "test_tool",
			description: "Test description",
			handler:     nil,
			wantErr:     true,
			errContains: "handler cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolRegistration(tt.toolName, tt.description, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("ValidateToolRegistration() error is nil or empty, want error containing %q", tt.errContains)
					return
				}
				errorMsg := err.Error()
				found := false
				for i := 0; i <= len(errorMsg)-len(tt.errContains); i++ {
					if errorMsg[i:i+len(tt.errContains)] == tt.errContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidateToolRegistration() error = %q, want error containing %q", errorMsg, tt.errContains)
				}
			}
		})
	}
}

func TestValidateResourceRegistration(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		toolName    string
		description string
		handler     interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid resource registration",
			uri:         "file:///test",
			toolName:    "test_resource",
			description: "Test resource description",
			handler:     func() {},
			wantErr:     false,
		},
		{
			name:        "empty URI",
			uri:         "",
			toolName:    "test_resource",
			description: "Test description",
			handler:     func() {},
			wantErr:     true,
			errContains: "resource URI cannot be empty",
		},
		{
			name:        "empty name",
			uri:         "file:///test",
			toolName:    "",
			description: "Test description",
			handler:     func() {},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "empty description",
			uri:         "file:///test",
			toolName:    "test_resource",
			description: "",
			handler:     func() {},
			wantErr:     true,
			errContains: "description cannot be empty",
		},
		{
			name:        "nil handler",
			uri:         "file:///test",
			toolName:    "test_resource",
			description: "Test description",
			handler:     nil,
			wantErr:     true,
			errContains: "handler cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceRegistration(tt.uri, tt.toolName, tt.description, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("ValidateResourceRegistration() error is nil, want error containing %q", tt.errContains)
					return
				}
				errorMsg := err.Error()
				found := false
				for i := 0; i <= len(errorMsg)-len(tt.errContains); i++ {
					if errorMsg[i:i+len(tt.errContains)] == tt.errContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidateResourceRegistration() error = %q, want error containing %q", errorMsg, tt.errContains)
				}
			}
		})
	}
}
