// tasks_lock_concurrency_test.go — Concurrent ClaimTaskForAgent: exclusive claim under SQLite transactions.
package database

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestConcurrentClaimSameTask_ExactlyOneSuccess(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()

	tmpDir := t.TempDir()
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer Close()

	taskID := "T-2000901"
	if err := CreateTask(context.Background(), &models.Todo2Task{
		ID:       taskID,
		Content:  "concurrent claim target",
		Status:   StatusTodo,
		Priority: "medium",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	const n = 32
	var wg sync.WaitGroup
	var successCount int32

	ctx := context.Background()
	lease := 5 * time.Minute

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			agentID := "concurrent-agent-" + strconv.Itoa(idx)
			res, err := ClaimTaskForAgent(ctx, taskID, agentID, lease)
			if err == nil && res != nil && res.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&successCount) != 1 {
		t.Fatalf("expected exactly 1 successful claim, got %d", successCount)
	}

	stored, err := GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if stored.Status != StatusInProgress {
		t.Fatalf("expected status %q after claim, got %q", StatusInProgress, stored.Status)
	}
}

func TestConcurrentClaimSameTask_SecondWaveAfterRelease(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()

	tmpDir := t.TempDir()
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer Close()

	taskID := "T-2000902"
	if err := CreateTask(context.Background(), &models.Todo2Task{
		ID:       taskID,
		Content:  "release then re-claim",
		Status:   StatusTodo,
		Priority: "medium",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	ctx := context.Background()
	lease := 5 * time.Minute

	first, err := ClaimTaskForAgent(ctx, taskID, "agent-round-a", lease)
	if err != nil || first == nil || !first.Success {
		t.Fatalf("first claim: err=%v success=%v", err, first != nil && first.Success)
	}

	if err := ReleaseTask(ctx, taskID, "agent-round-a"); err != nil {
		t.Fatalf("ReleaseTask: %v", err)
	}

	const n = 24
	var wg sync.WaitGroup
	var successCount int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			agentID := "reclaim-agent-" + strconv.Itoa(idx)
			res, err := ClaimTaskForAgent(ctx, taskID, agentID, lease)
			if err == nil && res != nil && res.Success {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&successCount) != 1 {
		t.Fatalf("expected exactly 1 successful reclaim, got %d", successCount)
	}
}
