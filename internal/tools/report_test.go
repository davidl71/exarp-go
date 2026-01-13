package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleReportOverview(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "overview with text format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "text",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result should be text format
			},
		},
		{
			name: "overview with json format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "overview with markdown format",
			params: map[string]interface{}{
				"action":        "overview",
				"output_format": "markdown",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result should be markdown format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleReportOverview(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReportOverview() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleReportPRD(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
			},
		},
		{
			name: "prd with project_name",
			params: map[string]interface{}{
				"action":       "prd",
				"project_name": "test-project",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result should contain project name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleReportPRD(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReportPRD() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleReport(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "overview action",
			params: map[string]interface{}{
				"action": "overview",
			},
			wantError: false,
		},
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
			},
			wantError: false,
		},
		{
			name: "scorecard action",
			params: map[string]interface{}{
				"action": "scorecard",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleReport(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleReport() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
