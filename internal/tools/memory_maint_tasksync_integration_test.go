// memory_maint_tasksync_integration_test.go — Integration test: memory_maint and task_workflow sync work together.
// Verifies sync (SQLite↔JSON) and memory_maint (health/gc) run in sequence in a shared project root.
package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

// TestMemoryMaintAndTaskSync_Integration verifies memory_maint and task_workflow sync work together:
// same PROJECT_ROOT, DB initialized, sync runs then memory_maint health (and optionally gc) succeed.
func TestMemoryMaintAndTaskSync_Integration(t *testing.T) {
	ctx := context.Background()
	projectRoot := t.TempDir()

	todo2Dir := filepath.Join(projectRoot, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("create .todo2: %v", err)
	}

	origRoot := os.Getenv("PROJECT_ROOT")
	t.Setenv("PROJECT_ROOT", projectRoot)
	defer func() {
		if origRoot != "" {
			_ = os.Setenv("PROJECT_ROOT", origRoot)
		} else {
			_ = os.Unsetenv("PROJECT_ROOT")
		}
	}()

	cfg, err := database.LoadConfig(projectRoot)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	_, self, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(self)))
	cfg.MigrationsDir = filepath.Join(repoRoot, "migrations")
	cfg.AutoMigrate = true
	if err := database.InitWithConfig(cfg); err != nil {
		t.Fatalf("database.InitWithConfig: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	// 1) task_workflow sync (ensures SQLite↔JSON sync works in this root)
	syncResult, err := handleTaskWorkflowNative(ctx, map[string]interface{}{"action": "sync"})
	if err != nil {
		t.Fatalf("task_workflow sync: %v", err)
	}
	if len(syncResult) == 0 {
		t.Fatal("task_workflow sync returned no result")
	}
	var syncData map[string]interface{}
	if err := json.Unmarshal([]byte(syncResult[0].Text), &syncData); err != nil {
		t.Fatalf("sync result JSON: %v", err)
	}
	if syncData["sync_results"] == nil {
		t.Error("expected sync_results in sync response")
	}

	// 2) memory_maint health (same project root; may have 0 memories)
	healthResult, err := handleMemoryMaintNative(ctx, map[string]interface{}{"action": "health"})
	if err != nil {
		t.Fatalf("memory_maint health: %v", err)
	}
	if len(healthResult) == 0 {
		t.Fatal("memory_maint health returned no result")
	}
	var healthData map[string]interface{}
	if err := json.Unmarshal([]byte(healthResult[0].Text), &healthData); err != nil {
		t.Fatalf("memory_maint health result JSON: %v", err)
	}
	if _, ok := healthData["health_score"]; !ok {
		t.Error("expected health_score in memory_maint health response")
	}

	// 3) memory_maint gc dry_run (ensures both tools work in sequence)
	gcResult, err := handleMemoryMaintNative(ctx, map[string]interface{}{"action": "gc", "dry_run": true})
	if err != nil {
		t.Fatalf("memory_maint gc: %v", err)
	}
	if len(gcResult) == 0 {
		t.Fatal("memory_maint gc returned no result")
	}
	var gcData map[string]interface{}
	if err := json.Unmarshal([]byte(gcResult[0].Text), &gcData); err != nil {
		t.Fatalf("memory_maint gc result JSON: %v", err)
	}
	if _, hasDryRun := gcData["dry_run"]; !hasDryRun && gcData["deleted_count"] == nil {
		t.Error("expected dry_run or deleted_count in memory_maint gc response")
	}
}
