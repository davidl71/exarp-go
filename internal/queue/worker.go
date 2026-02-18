package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/hibiken/asynq"
)

// RunWorker starts the Asynq worker: connects to Redis, registers the task_execute handler,
// and runs until context is cancelled. cfg must be enabled (Redis configured).
func RunWorker(ctx context.Context, cfg Config) error {
	if !cfg.Enabled() {
		return fmt.Errorf("queue not enabled: set REDIS_ADDR")
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      map[string]int{cfg.QueueName: 1},
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTaskExecute, handleTaskExecuteJob)
	return srv.Run(mux)
}

// handleTaskExecuteJob processes one task_execute job: init DB for project, claim task, run flow, release.
func handleTaskExecuteJob(ctx context.Context, t *asynq.Task) error {
	p, err := UnmarshalTaskExecutePayload(t.Payload())
	if err != nil {
		return fmt.Errorf("payload: %w", err)
	}
	if p.TaskID == "" || p.ProjectRoot == "" {
		return fmt.Errorf("task_id and project_root required")
	}

	// Ensure DB is initialized for this project (safe to call per task; re-inits if project changes).
	if fullCfg, err := config.LoadConfig(p.ProjectRoot); err == nil {
		dbCfg := database.DatabaseConfigFields{
			SQLitePath:       fullCfg.Database.SQLitePath,
			JSONFallbackPath: fullCfg.Database.JSONFallbackPath,
			BackupPath:       fullCfg.Database.BackupPath,
			MaxConnections:   fullCfg.Database.MaxConnections,
		}
		_ = database.InitWithCentralizedConfig(p.ProjectRoot, dbCfg)
	} else {
		_ = database.Init(p.ProjectRoot)
	}

	agentID, err := database.GetAgentID()
	if err != nil {
		return fmt.Errorf("agent ID: %w", err)
	}
	leaseDuration := config.TaskLockLease()
	if leaseDuration <= 0 {
		leaseDuration = 30 * 60 // 30 min default
	}

	claimResult, err := database.ClaimTaskForAgent(ctx, p.TaskID, agentID, leaseDuration)
	if err != nil {
		return fmt.Errorf("claim task: %w", err)
	}
	if claimResult != nil && claimResult.WasLocked {
		// Task already locked by another agent; retry later
		return fmt.Errorf("task %s locked by %s", p.TaskID, claimResult.LockedBy)
	}
	if claimResult == nil || !claimResult.Success {
		return fmt.Errorf("claim failed for %s", p.TaskID)
	}

	defer func() {
		if rerr := database.ReleaseTask(ctx, p.TaskID, agentID); rerr != nil {
			log.Printf("queue: release task %s: %v", p.TaskID, rerr)
		}
	}()

	_, err = tools.RunTaskExecutionFlow(ctx, tools.RunTaskExecutionFlowParams{
		TaskID:      p.TaskID,
		ProjectRoot: p.ProjectRoot,
		Apply:       true,
	})
	if err != nil {
		return fmt.Errorf("task_execute: %w", err)
	}
	return nil
}
