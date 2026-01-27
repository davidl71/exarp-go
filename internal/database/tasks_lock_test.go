package database

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

// initLockTestDB sets up a temp DB with migrations from the repo so lock columns (assignee, etc.) exist.
// Uses InitWithConfig with MigrationsDir from repo root (relative to this test file).
func initLockTestDB(t *testing.T) {
	t.Helper()
	_, self, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(self)))
	migrationsDir := filepath.Join(repoRoot, "migrations")
	tmpDir := t.TempDir()
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	cfg.Driver = DriverSQLite
	cfg.DSN = filepath.Join(tmpDir, ".todo2", "todo2.db")
	cfg.MigrationsDir = migrationsDir
	cfg.AutoMigrate = true // ensure migrations run regardless of DB_AUTO_MIGRATE
	err = InitWithConfig(cfg)
	if err != nil {
		t.Fatalf("InitWithConfig() error = %v (migrationsDir=%s)", err, migrationsDir)
	}
	// Require schema ≥ 2 so assignee/lock_until exist
	ver, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion() after init: %v", err)
	}
	if ver < 2 {
		t.Fatalf("schema not updated to 002: GetCurrentVersion()=%d (migrationsDir=%s, AutoMigrate=%v)",
			ver, migrationsDir, cfg.AutoMigrate)
	}
	t.Cleanup(func() { Close() })
}

func TestClaimTaskForAgent(t *testing.T) {
	// Setup (use repo migrations so assignee/lock_until columns exist)
	initLockTestDB(t)

	// Create a task first
	task := &models.Todo2Task{
		ID:       "T-LOCK-1",
		Content:  "Task for locking test",
		Status:   StatusTodo,
		Priority: "medium",
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Test claiming task
	agentID := "agent-test-1"
	leaseDuration := 30 * time.Minute

	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-1", agentID, leaseDuration)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %v", result.Error)
	}

	if result.Task == nil {
		t.Fatal("Expected task in result, got nil")
	}

	if result.Task.ID != "T-LOCK-1" {
		t.Errorf("Expected task ID T-LOCK-1, got %s", result.Task.ID)
	}

	if result.Task.Status != StatusInProgress {
		t.Errorf("Expected status %s, got %s", StatusInProgress, result.Task.Status)
	}

	// Verify lock was set in database
	retrieved, err := GetTask(context.Background(), "T-LOCK-1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	// Note: We can't check assignee/lock_until directly from GetTask since
	// those fields aren't in the Todo2Task model. But we can verify status changed.
	if retrieved.Status != StatusInProgress {
		t.Errorf("Expected status %s in database, got %s", StatusInProgress, retrieved.Status)
	}
}

func TestClaimTaskForAgent_AlreadyLocked(t *testing.T) {
	initLockTestDB(t)

	// Create a task
	task := &models.Todo2Task{
		ID:      "T-LOCK-2",
		Content: "Task already locked",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// First agent claims task
	agentID1 := "agent-test-1"
	result1, err := ClaimTaskForAgent(context.Background(), "T-LOCK-2", agentID1, 30*time.Minute)
	if err != nil {
		t.Fatalf("First ClaimTaskForAgent() error = %v", err)
	}
	if !result1.Success {
		t.Fatalf("First claim should succeed, got error: %v", result1.Error)
	}

	// Second agent tries to claim same task (should fail)
	agentID2 := "agent-test-2"
	result2, err := ClaimTaskForAgent(context.Background(), "T-LOCK-2", agentID2, 30*time.Minute)
	if err == nil {
		t.Error("Expected error when claiming already locked task, got nil")
	}

	if result2 != nil && !result2.WasLocked {
		t.Error("Expected WasLocked=true, got false")
	}

	if result2 != nil && result2.LockedBy != agentID1 {
		t.Errorf("Expected LockedBy=%s, got %s", agentID1, result2.LockedBy)
	}
}

func TestReleaseTask(t *testing.T) {
	initLockTestDB(t)

	// Create and claim task
	task := &models.Todo2Task{
		ID:      "T-LOCK-3",
		Content: "Task to release",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	agentID := "agent-test-1"
	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-3", agentID, 30*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Claim should succeed, got error: %v", result.Error)
	}

	// Release task
	err = ReleaseTask(context.Background(), "T-LOCK-3", agentID)
	if err != nil {
		t.Fatalf("ReleaseTask() error = %v", err)
	}

	// Verify task can be claimed again
	result2, err := ClaimTaskForAgent(context.Background(), "T-LOCK-3", "agent-test-2", 30*time.Minute)
	if err != nil {
		t.Fatalf("Second ClaimTaskForAgent() error = %v", err)
	}
	if !result2.Success {
		t.Fatalf("Task should be claimable after release, got error: %v", result2.Error)
	}
}

func TestReleaseTask_WrongAgent(t *testing.T) {
	initLockTestDB(t)

	// Create and claim task
	task := &models.Todo2Task{
		ID:      "T-LOCK-4",
		Content: "Task locked by agent1",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	agentID1 := "agent-test-1"
	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-4", agentID1, 30*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Claim should succeed, got error: %v", result.Error)
	}

	// Wrong agent tries to release
	agentID2 := "agent-test-2"
	err = ReleaseTask(context.Background(), "T-LOCK-4", agentID2)
	if err == nil {
		t.Error("Expected error when wrong agent releases task, got nil")
	}
}

func TestRenewLease(t *testing.T) {
	initLockTestDB(t)

	// Create and claim task
	task := &models.Todo2Task{
		ID:      "T-LOCK-5",
		Content: "Task for lease renewal",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	agentID := "agent-test-1"
	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-5", agentID, 10*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Claim should succeed, got error: %v", result.Error)
	}

	// Renew lease
	err = RenewLease(context.Background(), "T-LOCK-5", agentID, 30*time.Minute)
	if err != nil {
		t.Fatalf("RenewLease() error = %v", err)
	}
}

func TestCleanupExpiredLocks(t *testing.T) {
	initLockTestDB(t)

	// Create a task
	task := &models.Todo2Task{
		ID:      "T-LOCK-6",
		Content: "Task with expired lock",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// We need to manually set an expired lock in the database
	// Since we can't directly manipulate lock_until through the API,
	// we'll use a short lease and wait, or use SQL directly
	// For this test, let's use SQL to set an expired lock
	db, err := GetDB()
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}

	now := time.Now().Unix()
	expiredTime := now - 300 // 5 minutes ago

	_, err = db.Exec(`
		UPDATE tasks SET
			assignee = ?,
			assigned_at = ?,
			lock_until = ?,
			status = ?,
			version = version + 1
		WHERE id = ?
	`, "expired-agent", now-600, expiredTime, StatusInProgress, "T-LOCK-6")
	if err != nil {
		t.Fatalf("Failed to set expired lock: %v", err)
	}

	// Cleanup expired locks
	cleaned, err := CleanupExpiredLocks(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpiredLocks() error = %v", err)
	}

	if cleaned < 1 {
		t.Errorf("Expected at least 1 lock cleaned, got %d", cleaned)
	}

	// Verify task can be claimed again
	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-6", "agent-test-1", 30*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() after cleanup error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Task should be claimable after cleanup, got error: %v", result.Error)
	}
}

func TestDetectStaleLocks(t *testing.T) {
	initLockTestDB(t)

	// Create tasks
	var err error
	tasks := []*models.Todo2Task{
		{ID: "T-LOCK-7", Content: "Task 1", Status: StatusTodo},
		{ID: "T-LOCK-8", Content: "Task 2", Status: StatusTodo},
	}
	for _, task := range tasks {
		err = CreateTask(context.Background(), task)
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	// Claim one task (valid lock)
	_, err = ClaimTaskForAgent(context.Background(), "T-LOCK-7", "agent-1", 30*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}

	// Manually set expired lock on second task
	db, err := GetDB()
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}

	now := time.Now().Unix()
	expiredTime := now - 600 // 10 minutes ago (stale)

	_, err = db.Exec(`
		UPDATE tasks SET
			assignee = ?,
			assigned_at = ?,
			lock_until = ?,
			status = ?,
			version = version + 1
		WHERE id = ?
	`, "expired-agent", now-1200, expiredTime, StatusInProgress, "T-LOCK-8")
	if err != nil {
		t.Fatalf("Failed to set expired lock: %v", err)
	}

	// Detect stale locks
	info, err := DetectStaleLocks(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("DetectStaleLocks() error = %v", err)
	}

	if info.ExpiredCount < 1 {
		t.Errorf("Expected at least 1 expired lock, got %d", info.ExpiredCount)
	}

	if info.StaleCount < 1 {
		t.Errorf("Expected at least 1 stale lock, got %d", info.StaleCount)
	}

	if len(info.Locks) < 2 {
		t.Errorf("Expected at least 2 locked tasks, got %d", len(info.Locks))
	}
}

func TestGetLockStatus(t *testing.T) {
	initLockTestDB(t)

	// Create and claim task
	task := &models.Todo2Task{
		ID:      "T-LOCK-9",
		Content: "Task for status check",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	agentID := "agent-test-1"
	result, err := ClaimTaskForAgent(context.Background(), "T-LOCK-9", agentID, 30*time.Minute)
	if err != nil {
		t.Fatalf("ClaimTaskForAgent() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Claim should succeed, got error: %v", result.Error)
	}

	// Get lock status
	status, err := GetLockStatus(context.Background(), "T-LOCK-9")
	if err != nil {
		t.Fatalf("GetLockStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("Expected lock status, got nil")
	}

	if status.TaskID != "T-LOCK-9" {
		t.Errorf("Expected TaskID T-LOCK-9, got %s", status.TaskID)
	}

	if status.Assignee != agentID {
		t.Errorf("Expected Assignee %s, got %s", agentID, status.Assignee)
	}

	if status.IsExpired {
		t.Error("Expected lock not to be expired")
	}

	if status.IsStale {
		t.Error("Expected lock not to be stale")
	}
}

func TestGetLockStatus_NotLocked(t *testing.T) {
	initLockTestDB(t)

	// Create task (not locked)
	task := &models.Todo2Task{
		ID:      "T-LOCK-10",
		Content: "Unlocked task",
		Status:  StatusTodo,
	}
	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Get lock status (should return nil for unlocked task)
	status, err := GetLockStatus(context.Background(), "T-LOCK-10")
	if err != nil {
		t.Fatalf("GetLockStatus() error = %v", err)
	}

	if status != nil {
		t.Errorf("Expected nil for unlocked task, got status: %+v", status)
	}
}
