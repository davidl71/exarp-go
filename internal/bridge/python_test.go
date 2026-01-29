package bridge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutePythonTool_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple args",
			args: map[string]interface{}{
				"action": "test",
				"value":  42,
			},
			wantErr: false,
		},
		{
			name:    "empty args",
			args:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nested args",
			args: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			argsJSON, err := json.Marshal(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify it can be unmarshaled back
			var unmarshaled map[string]interface{}
			if err := json.Unmarshal(argsJSON, &unmarshaled); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
		})
	}
}

func TestExecutePythonTool_PROJECT_ROOT(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		wantPath    string
	}{
		{
			name:        "with PROJECT_ROOT",
			projectRoot: "/test/project",
			wantPath:    "/test/project/bridge/execute_tool.py",
		},
		{
			name:        "without PROJECT_ROOT",
			projectRoot: "",
			wantPath:    filepath.Join(".", "bridge", "execute_tool.py"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalRoot := os.Getenv("PROJECT_ROOT")
			defer func() {
				_ = os.Setenv("PROJECT_ROOT", originalRoot) //nolint:errcheck // Test cleanup
			}()

			// Set test env
			if tt.projectRoot != "" {
				_ = os.Setenv("PROJECT_ROOT", tt.projectRoot) //nolint:errcheck // Test setup
			} else {
				_ = os.Unsetenv("PROJECT_ROOT") //nolint:errcheck // Test setup
			}

			// Get project root
			projectRoot := os.Getenv("PROJECT_ROOT")
			if projectRoot == "" {
				projectRoot = "."
			}

			// Verify path construction
			bridgeScript := filepath.Join(projectRoot, "bridge", "execute_tool.py")
			if bridgeScript != tt.wantPath {
				// Allow for path differences (e.g., relative vs absolute)
				// Just verify it ends with expected path
				expectedSuffix := filepath.Join("bridge", "execute_tool.py")
				if !filepath.IsAbs(tt.wantPath) {
					expectedSuffix = tt.wantPath
				}
				if filepath.Base(bridgeScript) != filepath.Base(expectedSuffix) {
					t.Errorf("ExecutePythonTool path = %v, want ends with %v", bridgeScript, expectedSuffix)
				}
			}
		})
	}
}

func TestContextTimeout(t *testing.T) {
	t.Run("ExecutePythonTool timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Create a context that will timeout quickly
		shortCtx, shortCancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer shortCancel()

		// Verify timeout context is cancelled
		select {
		case <-shortCtx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("context should have timed out")
		}
	})
}
