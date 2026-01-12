package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestHandleInferSessionModeNative(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		params         map[string]interface{}
		projectRootEnv string
		wantMode       SessionMode
		wantErr        bool
	}{
		{
			name:           "default inference",
			params:         map[string]interface{}{},
			projectRootEnv: "",
			wantMode:       SessionModeUNKNOWN, // Will be unknown if no project root
			wantErr:        false,
		},
		{
			name:           "force recompute",
			params:         map[string]interface{}{"force_recompute": true},
			projectRootEnv: "",
			wantMode:       SessionModeUNKNOWN,
			wantErr:        false,
		},
		{
			name:           "with project root",
			params:         map[string]interface{}{},
			projectRootEnv: "/tmp",
			wantMode:       SessionModeUNKNOWN, // May be unknown if no tasks
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original PROJECT_ROOT
			originalRoot := os.Getenv("PROJECT_ROOT")
			defer os.Setenv("PROJECT_ROOT", originalRoot)

			// Set test PROJECT_ROOT if provided
			if tt.projectRootEnv != "" {
				os.Setenv("PROJECT_ROOT", tt.projectRootEnv)
			} else {
				os.Unsetenv("PROJECT_ROOT")
			}

			// Clear cache
			lastInferenceCache = nil
			lastInferenceTime = time.Time{}

			result, err := HandleInferSessionModeNative(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleInferSessionModeNative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				if len(result) == 0 {
					t.Error("HandleInferSessionModeNative() returned empty result")
					return
				}

				// Parse result to verify structure
				var inferenceResult ModeInferenceResult
				if err := json.Unmarshal([]byte(result[0].Text), &inferenceResult); err != nil {
					t.Errorf("HandleInferSessionModeNative() returned invalid JSON: %v", err)
					return
				}

				// Verify result structure
				if inferenceResult.Mode == "" {
					t.Error("HandleInferSessionModeNative() returned empty mode")
				}
				if inferenceResult.Confidence < 0 || inferenceResult.Confidence > 1 {
					t.Errorf("HandleInferSessionModeNative() returned invalid confidence: %v", inferenceResult.Confidence)
				}
				if inferenceResult.Timestamp == "" {
					t.Error("HandleInferSessionModeNative() returned empty timestamp")
				}
			}
		})
	}
}

func TestInferModeFromTasks(t *testing.T) {
	tests := []struct {
		name     string
		tasks     []Todo2Task
		wantMode  SessionMode
		wantConf  float64
		checkConf bool
	}{
		{
			name:     "empty tasks",
			tasks:     []Todo2Task{},
			wantMode:  SessionModeASK,
			wantConf:  0.3,
			checkConf: true,
		},
		{
			name: "high in-progress ratio - agent mode",
			tasks: []Todo2Task{
				{ID: "T-1", Status: "In Progress"},
				{ID: "T-2", Status: "In Progress"},
				{ID: "T-3", Status: "In Progress"},
				{ID: "T-4", Status: "Todo"},
				{ID: "T-5", Status: "Todo"},
			},
			wantMode:  SessionModeAGENT,
			wantConf:  0.0, // Don't check exact confidence
			checkConf: false,
		},
		{
			name: "low task count - ask mode",
			tasks: []Todo2Task{
				{ID: "T-1", Status: "Todo"},
				{ID: "T-2", Status: "Done"},
			},
			wantMode:  SessionModeASK,
			wantConf:  0.0,
			checkConf: false,
		},
		{
			name: "high completion ratio - ask mode",
			tasks: []Todo2Task{
				{ID: "T-1", Status: "Done"},
				{ID: "T-2", Status: "Done"},
				{ID: "T-3", Status: "Done"},
				{ID: "T-4", Status: "Done"},
				{ID: "T-5", Status: "Done"},
				{ID: "T-6", Status: "Todo"},
				{ID: "T-7", Status: "Todo"},
			},
			wantMode:  SessionModeASK,
			wantConf:  0.0,
			checkConf: false,
		},
		{
			name: "many tasks with dependencies - agent mode",
			tasks: []Todo2Task{
				{ID: "T-1", Status: "In Progress", Dependencies: []string{"T-2"}},
				{ID: "T-2", Status: "In Progress", Dependencies: []string{"T-3"}},
				{ID: "T-3", Status: "In Progress", Dependencies: []string{"T-4"}},
				{ID: "T-4", Status: "Todo", Dependencies: []string{"T-5"}},
				{ID: "T-5", Status: "Todo"},
			},
			wantMode:  SessionModeAGENT,
			wantConf:  0.0,
			checkConf: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferModeFromTasks(tt.tasks, "/tmp")
			if result.Mode != tt.wantMode {
				t.Errorf("inferModeFromTasks() mode = %v, want %v", result.Mode, tt.wantMode)
			}
			if tt.checkConf && result.Confidence != tt.wantConf {
				t.Errorf("inferModeFromTasks() confidence = %v, want %v", result.Confidence, tt.wantConf)
			}
			if result.Confidence < 0 || result.Confidence > 1 {
				t.Errorf("inferModeFromTasks() returned invalid confidence: %v", result.Confidence)
			}
			if len(result.Reasoning) == 0 && len(tt.tasks) > 0 {
				t.Error("inferModeFromTasks() returned empty reasoning for non-empty tasks")
			}
			if result.Timestamp == "" {
				t.Error("inferModeFromTasks() returned empty timestamp")
			}
		})
	}
}

func TestAnalyzeTaskPatterns(t *testing.T) {
	tasks := []Todo2Task{
		{ID: "T-1", Status: "Todo", Priority: "high", Tags: []string{"bug", "urgent"}},
		{ID: "T-2", Status: "In Progress", Priority: "medium", Tags: []string{"feature"}},
		{ID: "T-3", Status: "Done", Priority: "low", Tags: []string{"documentation"}},
	}

	metrics := analyzeTaskPatterns(tasks)

	// Verify metrics structure
	if metrics["status_distribution"] == nil {
		t.Error("analyzeTaskPatterns() missing status_distribution")
	}
	if metrics["priority_distribution"] == nil {
		t.Error("analyzeTaskPatterns() missing priority_distribution")
	}
	if metrics["tag_counts"] == nil {
		t.Error("analyzeTaskPatterns() missing tag_counts")
	}

	// Verify status distribution
	statusDist, ok := metrics["status_distribution"].(map[string]int)
	if !ok {
		t.Error("analyzeTaskPatterns() status_distribution is not map[string]int")
	} else {
		if statusDist["Todo"] != 1 {
			t.Errorf("analyzeTaskPatterns() status_distribution[Todo] = %v, want 1", statusDist["Todo"])
		}
		if statusDist["In Progress"] != 1 {
			t.Errorf("analyzeTaskPatterns() status_distribution[In Progress] = %v, want 1", statusDist["In Progress"])
		}
		if statusDist["Done"] != 1 {
			t.Errorf("analyzeTaskPatterns() status_distribution[Done] = %v, want 1", statusDist["Done"])
		}
	}
}

func TestMarshalInferenceResult(t *testing.T) {
	result := ModeInferenceResult{
		Mode:       SessionModeAGENT,
		Confidence: 0.85,
		Reasoning:  []string{"High in-progress ratio"},
		Metrics:    map[string]interface{}{"total_tasks": 10},
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	content, err := marshalInferenceResult(result)
	if err != nil {
		t.Fatalf("marshalInferenceResult() error = %v", err)
	}

	if len(content) == 0 {
		t.Fatal("marshalInferenceResult() returned empty content")
	}

	// Verify JSON is valid
	var parsed ModeInferenceResult
	if err := json.Unmarshal([]byte(content[0].Text), &parsed); err != nil {
		t.Fatalf("marshalInferenceResult() returned invalid JSON: %v", err)
	}

	if parsed.Mode != result.Mode {
		t.Errorf("marshalInferenceResult() mode = %v, want %v", parsed.Mode, result.Mode)
	}
	if parsed.Confidence != result.Confidence {
		t.Errorf("marshalInferenceResult() confidence = %v, want %v", parsed.Confidence, result.Confidence)
	}
}

func TestHandleInferSessionModeNative_Cache(t *testing.T) {
	ctx := context.Background()

	// Clear cache
	lastInferenceCache = nil
	lastInferenceTime = time.Time{}

	// First call - should compute
	result1, err := HandleInferSessionModeNative(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("HandleInferSessionModeNative() first call error = %v", err)
	}

	// Second call immediately - should use cache
	result2, err := HandleInferSessionModeNative(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("HandleInferSessionModeNative() second call error = %v", err)
	}

	// Results should be identical (cached)
	if result1[0].Text != result2[0].Text {
		t.Error("HandleInferSessionModeNative() cache not working - results differ")
	}

	// Force recompute - should bypass cache
	result3, err := HandleInferSessionModeNative(ctx, map[string]interface{}{"force_recompute": true})
	if err != nil {
		t.Fatalf("HandleInferSessionModeNative() force recompute error = %v", err)
	}

	// Result should be valid JSON even if different
	var parsed ModeInferenceResult
	if err := json.Unmarshal([]byte(result3[0].Text), &parsed); err != nil {
		t.Errorf("HandleInferSessionModeNative() force recompute returned invalid JSON: %v", err)
	}
}
