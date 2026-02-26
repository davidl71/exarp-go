package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestHandleServerStatusNative(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		args           json.RawMessage
		projectRootEnv string
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent)
	}{
		{
			name:           "default status check",
			args:           fixtures.MustToJSONRawMessage(map[string]interface{}{}),
			projectRootEnv: "",
			wantErr:        false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleServerStatusNative() returned empty result")
					return
				}

				var status map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &status); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if status["status"] != "operational" {
					t.Errorf("expected status 'operational', got %v", status["status"])
				}

				if status["version"] == nil {
					t.Error("expected version field in result")
				}

				if status["project_root"] == nil {
					t.Error("expected project_root field in result")
				}
			},
		},
		{
			name:           "with project root",
			args:           fixtures.MustToJSONRawMessage(map[string]interface{}{}),
			projectRootEnv: "/tmp/test_project",
			wantErr:        false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleServerStatusNative() returned empty result")
					return
				}

				var status map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &status); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if status["project_root"] != "/tmp/test_project" {
					t.Errorf("expected project_root '/tmp/test_project', got %v", status["project_root"])
				}
			},
		},
		{
			name:           "with no PROJECT_ROOT env resolves via FindProjectRoot",
			args:           fixtures.MustToJSONRawMessage(map[string]interface{}{}),
			projectRootEnv: "",
			wantErr:        false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleServerStatusNative() returned empty result")
					return
				}

				var status map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &status); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				pr, ok := status["project_root"].(string)
				if !ok || pr == "" {
					t.Errorf("expected non-empty project_root, got %v", status["project_root"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test PROJECT_ROOT if provided
			if tt.projectRootEnv != "" {
				t.Setenv("PROJECT_ROOT", tt.projectRootEnv)
			} else {
				t.Setenv("PROJECT_ROOT", "")
			}

			result, err := handleServerStatusNative(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleServerStatusNative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}
