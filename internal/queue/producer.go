package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/davidl71/exarp-go/internal/tools"
)

// Producer enqueues task_execute jobs. When config is not enabled, Enqueue methods return nil without error.
type Producer struct {
	client *asynq.Client
	cfg    Config
}

// NewProducer creates a producer from config. If config is not enabled, returns nil (caller may still call Enqueue; they no-op).
func NewProducer(cfg Config) (*Producer, error) {
	if !cfg.Enabled() {
		return &Producer{cfg: cfg}, nil
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	return &Producer{client: client, cfg: cfg}, nil
}

// Close closes the Redis client. Safe to call if client is nil.
func (p *Producer) Close() error {
	if p == nil || p.client == nil {
		return nil
	}
	return p.client.Close()
}

// EnqueueTask enqueues one task_execute job for the given task_id and project_root.
// If producer is not enabled (no Redis), returns (0, nil).
func (p *Producer) EnqueueTask(ctx context.Context, taskID, projectRoot string) (enqueued int, err error) {
	if p == nil || !p.cfg.Enabled() || p.client == nil {
		return 0, nil
	}
	payload, err := (&TaskExecutePayload{TaskID: taskID, ProjectRoot: projectRoot}).Marshal()
	if err != nil {
		return 0, err
	}
	task := asynq.NewTask(TypeTaskExecute, payload, asynq.Queue(p.cfg.QueueName))
	_, err = p.client.EnqueueContext(ctx, task)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// EnqueueWave enqueues task_execute jobs for all task IDs in the given wave index.
// Wave 0 = first wave (no dependencies), wave 1 = next, etc. Uses tools.GetExecutionWaves.
// If producer is not enabled, returns (0, nil). projectRoot is used to compute waves and as payload.
func (p *Producer) EnqueueWave(ctx context.Context, projectRoot string, waveIndex int) (enqueued int, err error) {
	if p == nil || !p.cfg.Enabled() || p.client == nil {
		return 0, nil
	}
	waves, err := tools.GetExecutionWaves(ctx, projectRoot)
	if err != nil {
		return 0, fmt.Errorf("get waves: %w", err)
	}
	taskIDs, ok := waves[waveIndex]
	if !ok || len(taskIDs) == 0 {
		return 0, fmt.Errorf("wave %d has no tasks (waves: %d levels)", waveIndex, len(waves))
	}
	for _, taskID := range taskIDs {
		payload, err := (&TaskExecutePayload{TaskID: taskID, ProjectRoot: projectRoot}).Marshal()
		if err != nil {
			return enqueued, err
		}
		task := asynq.NewTask(TypeTaskExecute, payload, asynq.Queue(p.cfg.QueueName))
		_, err = p.client.EnqueueContext(ctx, task)
		if err != nil {
			return enqueued, err
		}
		enqueued++
	}
	return enqueued, nil
}
