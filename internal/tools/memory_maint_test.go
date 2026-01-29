package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestProjectRoot creates a temporary directory with .todo2 marker
// and sets PROJECT_ROOT environment variable for tests
func setupTestProjectRoot(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Create .todo2 directory to satisfy FindProjectRoot()
	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("Failed to create .todo2 directory: %v", err)
	}

	// Set PROJECT_ROOT environment variable
	originalProjectRoot := os.Getenv("PROJECT_ROOT")
	os.Setenv("PROJECT_ROOT", tmpDir)

	// Cleanup function
	cleanup := func() {
		if originalProjectRoot != "" {
			os.Setenv("PROJECT_ROOT", originalProjectRoot)
		} else {
			os.Unsetenv("PROJECT_ROOT")
		}
	}

	return tmpDir, cleanup
}

// createTestMemory creates a test memory file in the project root
// Uses protobuf format (.pb) to match production format
func createTestMemory(t *testing.T, projectRoot string, memory Memory) {
	memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("Failed to create memories directory: %v", err)
	}

	// Save as protobuf binary (.pb)
	memoryPath := filepath.Join(memoriesDir, memory.ID+".pb")
	data, err := SerializeMemoryToProtobuf(&memory)
	if err != nil {
		t.Fatalf("Failed to serialize memory to protobuf: %v", err)
	}

	if err := os.WriteFile(memoryPath, data, 0644); err != nil {
		t.Fatalf("Failed to write memory file: %v", err)
	}
}

// createTestMemoryWithAge creates a test memory with a specific age in days
func createTestMemoryWithAge(t *testing.T, projectRoot string, id, title, category string, ageDays int, linkedTasks []string) {
	createdAt := time.Now().AddDate(0, 0, -ageDays)
	memory := Memory{
		ID:          id,
		Title:       title,
		Content:     "Test content for " + title,
		Category:    category,
		LinkedTasks: linkedTasks,
		CreatedAt:   createdAt.Format(time.RFC3339),
		SessionDate: createdAt.Format("2006-01-02"),
	}
	createTestMemory(t, projectRoot, memory)
}

func TestHandleMemoryMaintNative_Health(t *testing.T) {
	projectRoot, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	tests := []struct {
		name            string
		setupMemories   func(*testing.T)
		wantHealthScore float64
		wantIssues      int
		wantErr         bool
	}{
		{
			name: "empty memories",
			setupMemories: func(t *testing.T) {
				// No memories created
			},
			wantHealthScore: 0,
			wantIssues:      1,
			wantErr:         false,
		},
		{
			name: "healthy memories",
			setupMemories: func(t *testing.T) {
				// Create recent memories in different categories
				createTestMemoryWithAge(t, projectRoot, "m1", "Test Memory 1", "insight", 1, []string{"T-1"})
				createTestMemoryWithAge(t, projectRoot, "m2", "Test Memory 2", "research", 2, []string{})
				createTestMemoryWithAge(t, projectRoot, "m3", "Test Memory 3", "architecture", 3, []string{"T-2"})
			},
			wantHealthScore: 100,
			wantIssues:      0,
			wantErr:         false,
		},
		{
			name: "old memories",
			setupMemories: func(t *testing.T) {
				// Create mostly old memories (>3 months)
				for i := 0; i < 10; i++ {
					createTestMemoryWithAge(t, projectRoot, fmt.Sprintf("old-%d", i), "Old Memory", "insight", 100+i, []string{})
				}
			},
			wantHealthScore: 70, // Should be penalized for old memories (20 points) and category imbalance (10 points)
			wantIssues:      2,  // Old memories + category imbalance
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing memories
			memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
			os.RemoveAll(memoriesDir)

			// Setup test memories
			tt.setupMemories(t)

			// Call health action
			ctx := context.Background()
			params := map[string]interface{}{"action": "health"}
			result, err := handleMemoryMaintHealth(ctx, params)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleMemoryMaintHealth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Parse result
			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Check health score
			if score, ok := resultData["health_score"].(float64); ok {
				if score != tt.wantHealthScore {
					t.Errorf("health_score = %v, want %v", score, tt.wantHealthScore)
				}
			} else {
				t.Error("health_score not found or wrong type")
			}

			// Check issues
			if issues, ok := resultData["issues"].([]interface{}); ok {
				if len(issues) != tt.wantIssues {
					t.Errorf("issues count = %v, want %v", len(issues), tt.wantIssues)
				}
			} else {
				t.Error("issues not found or wrong type")
			}
		})
	}
}

func TestHandleMemoryMaintNative_GC(t *testing.T) {
	projectRoot, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupMemories func(*testing.T)
		params        map[string]interface{}
		wantDeleted   int
		wantErr       bool
	}{
		{
			name: "dry run - old memories",
			setupMemories: func(t *testing.T) {
				// Create old memories (>90 days)
				createTestMemoryWithAge(t, projectRoot, "old1", "Old Memory 1", "insight", 100, []string{})
				createTestMemoryWithAge(t, projectRoot, "old2", "Old Memory 2", "research", 120, []string{})
				// Create recent memory
				createTestMemoryWithAge(t, projectRoot, "recent", "Recent Memory", "insight", 5, []string{"T-1"})
			},
			params: map[string]interface{}{
				"action":       "gc",
				"max_age_days": 90.0,
				"dry_run":      true,
			},
			wantDeleted: 2,
			wantErr:     false,
		},
		{
			name: "delete old memories",
			setupMemories: func(t *testing.T) {
				createTestMemoryWithAge(t, projectRoot, "old1", "Old Memory 1", "insight", 100, []string{})
				createTestMemoryWithAge(t, projectRoot, "recent", "Recent Memory", "insight", 5, []string{"T-1"})
			},
			params: map[string]interface{}{
				"action":       "gc",
				"max_age_days": 90.0,
				"dry_run":      false,
			},
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name: "delete orphaned memories",
			setupMemories: func(t *testing.T) {
				// Create orphaned memory (>30 days, no linked tasks)
				createTestMemoryWithAge(t, projectRoot, "orphaned", "Orphaned Memory", "insight", 40, []string{})
				// Create memory with linked task
				createTestMemoryWithAge(t, projectRoot, "linked", "Linked Memory", "insight", 40, []string{"T-1"})
			},
			params: map[string]interface{}{
				"action":          "gc",
				"delete_orphaned": true,
				"dry_run":         false,
			},
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name: "delete duplicates",
			setupMemories: func(t *testing.T) {
				// Create duplicate memories (same title and category)
				memory1 := Memory{
					ID:          "dup1",
					Title:       "Duplicate Title",
					Content:     "Content 1",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
				}
				memory2 := Memory{
					ID:          "dup2",
					Title:       "Duplicate Title",
					Content:     "Content 2",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
				}
				createTestMemory(t, projectRoot, memory1)
				createTestMemory(t, projectRoot, memory2)
			},
			params: map[string]interface{}{
				"action":            "gc",
				"delete_duplicates": true,
				"dry_run":           false,
			},
			wantDeleted: 1, // Older duplicate should be deleted
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing memories
			memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
			os.RemoveAll(memoriesDir)

			// Setup test memories
			tt.setupMemories(t)

			// Call GC action
			ctx := context.Background()
			result, err := handleMemoryMaintGC(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleMemoryMaintGC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Parse result
			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Check deleted count
			if deletedCount, ok := resultData["deleted_count"].(float64); ok {
				if int(deletedCount) != tt.wantDeleted {
					t.Errorf("deleted_count = %v, want %v", deletedCount, tt.wantDeleted)
				}
			} else {
				t.Error("deleted_count not found or wrong type")
			}
		})
	}
}

func TestHandleMemoryMaintNative_Prune(t *testing.T) {
	projectRoot, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupMemories func(*testing.T)
		params        map[string]interface{}
		wantPruned    int
		wantErr       bool
	}{
		{
			name: "dry run - prune low value memories",
			setupMemories: func(t *testing.T) {
				// Create memories with different characteristics
				// Low value: old, no linked tasks, scorecard type
				oldScorecard := Memory{
					ID:          "scorecard1",
					Title:       "Scorecard Memory",
					Content:     "Scorecard content",
					Category:    "insight",
					LinkedTasks: []string{},
					Metadata:    map[string]interface{}{"type": "scorecard"},
					CreatedAt:   time.Now().AddDate(0, 0, -50).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -50).Format("2006-01-02"),
				}
				createTestMemory(t, projectRoot, oldScorecard)

				// High value: recent, linked task
				createTestMemoryWithAge(t, projectRoot, "high-value", "High Value Memory", "insight", 2, []string{"T-1"})
			},
			params: map[string]interface{}{
				"action":          "prune",
				"value_threshold": 0.3,
				"keep_minimum":    1.0, // Lower keep_minimum so we can prune the scorecard memory
				"dry_run":         true,
			},
			wantPruned: 1,
			wantErr:    false,
		},
		{
			name: "prune with keep minimum",
			setupMemories: func(t *testing.T) {
				// Create exactly keep_minimum memories (should not prune any)
				for i := 0; i < 50; i++ {
					createTestMemoryWithAge(t, projectRoot, fmt.Sprintf("mem-%d", i), "Memory", "insight", 50+i, []string{})
				}
			},
			params: map[string]interface{}{
				"action":          "prune",
				"value_threshold": 0.3,
				"keep_minimum":    50.0,
				"dry_run":         false,
			},
			wantPruned: 0, // Should keep all due to keep_minimum
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing memories
			memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
			os.RemoveAll(memoriesDir)

			// Setup test memories
			tt.setupMemories(t)

			// Call prune action
			ctx := context.Background()
			result, err := handleMemoryMaintPrune(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleMemoryMaintPrune() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Parse result
			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Check pruned count
			if prunedCount, ok := resultData["pruned_count"].(float64); ok {
				if int(prunedCount) != tt.wantPruned {
					t.Errorf("pruned_count = %v, want %v", prunedCount, tt.wantPruned)
				}
			} else {
				t.Error("pruned_count not found or wrong type")
			}
		})
	}
}

func TestHandleMemoryMaintNative_Consolidate(t *testing.T) {
	projectRoot, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupMemories func(*testing.T)
		params        map[string]interface{}
		wantGroups    int
		wantErr       bool
	}{
		{
			name: "dry run - consolidate similar titles",
			setupMemories: func(t *testing.T) {
				// Create similar memories (same category)
				memory1 := Memory{
					ID:          "sim1",
					Title:       "Test Memory",
					Content:     "Content 1",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
				}
				memory2 := Memory{
					ID:          "sim2",
					Title:       "Test Memory", // Same title
					Content:     "Content 2",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
				}
				createTestMemory(t, projectRoot, memory1)
				createTestMemory(t, projectRoot, memory2)
			},
			params: map[string]interface{}{
				"action":               "consolidate",
				"similarity_threshold": 0.85,
				"merge_strategy":       "newest",
				"dry_run":              true,
			},
			wantGroups: 1,
			wantErr:    false,
		},
		{
			name: "consolidate with newest strategy",
			setupMemories: func(t *testing.T) {
				memory1 := Memory{
					ID:          "old",
					Title:       "Duplicate Title",
					Content:     "Old content",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
				}
				memory2 := Memory{
					ID:          "new",
					Title:       "Duplicate Title",
					Content:     "New content",
					Category:    "insight",
					LinkedTasks: []string{},
					CreatedAt:   time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
					SessionDate: time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
				}
				createTestMemory(t, projectRoot, memory1)
				createTestMemory(t, projectRoot, memory2)
			},
			params: map[string]interface{}{
				"action":               "consolidate",
				"similarity_threshold": 0.85,
				"merge_strategy":       "newest",
				"dry_run":              false,
			},
			wantGroups: 1,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing memories
			memoriesDir := filepath.Join(projectRoot, ".exarp", "memories")
			os.RemoveAll(memoriesDir)

			// Setup test memories
			tt.setupMemories(t)

			// Call consolidate action
			ctx := context.Background()
			result, err := handleMemoryMaintConsolidate(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleMemoryMaintConsolidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Parse result
			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Check groups found
			if groupsFound, ok := resultData["groups_found"].(float64); ok {
				if int(groupsFound) != tt.wantGroups {
					t.Errorf("groups_found = %v, want %v", groupsFound, tt.wantGroups)
				}
			} else {
				t.Error("groups_found not found or wrong type")
			}
		})
	}
}

func TestHandleMemoryMaintNative_Dream(t *testing.T) {
	_, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	// Dream action should return error (requires advisor integration)
	ctx := context.Background()
	params := map[string]interface{}{"action": "dream"}

	argsJSON, _ := json.Marshal(params)
	_, err := handleMemoryMaintNative(ctx, argsJSON)

	if err == nil {
		t.Error("handleMemoryMaintNative() with dream action should return error")
	}

	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestHandleMemoryMaintNative_UnknownAction(t *testing.T) {
	_, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	ctx := context.Background()
	params := map[string]interface{}{"action": "unknown_action"}

	argsJSON, _ := json.Marshal(params)
	_, err := handleMemoryMaintNative(ctx, argsJSON)

	if err == nil {
		t.Error("handleMemoryMaintNative() with unknown action should return error")
	}
}

func TestHandleMemoryMaintNative_DefaultAction(t *testing.T) {
	_, cleanup := setupTestProjectRoot(t)
	defer cleanup()

	// Default action should be "health"
	ctx := context.Background()
	params := map[string]interface{}{} // No action specified

	argsJSON, _ := json.Marshal(params)
	result, err := handleMemoryMaintNative(ctx, argsJSON)

	if err != nil {
		t.Errorf("handleMemoryMaintNative() with default action error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Should have health_score (indicates health action was called)
	if _, ok := resultData["health_score"]; !ok {
		t.Error("Expected health_score in result (indicates health action was called)")
	}
}
