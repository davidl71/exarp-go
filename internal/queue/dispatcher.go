// dispatcher.go — Periodically finds and enqueues the next wave of tasks.
package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidl71/exarp-go/internal/tools"
)

// RunDispatcher periodically checks for the next wave with unfinished tasks
// and enqueues it via the Producer. It runs until ctx is cancelled.
func RunDispatcher(ctx context.Context, cfg Config, projectRoot string, interval time.Duration) error {
	if !cfg.Enabled() {
		return fmt.Errorf("queue not enabled: set REDIS_ADDR (e.g. 127.0.0.1:6379)")
	}
	producer, err := NewProducer(cfg)
	if err != nil {
		return fmt.Errorf("creating producer: %w", err)
	}
	defer producer.Close()

	log.Printf("dispatcher: starting (interval=%s, project=%s)", interval, projectRoot)

	if err := dispatchOnce(ctx, producer, projectRoot); err != nil {
		log.Printf("dispatcher: initial dispatch: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("dispatcher: stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := dispatchOnce(ctx, producer, projectRoot); err != nil {
				log.Printf("dispatcher: %v", err)
			}
		}
	}
}

// dispatchOnce finds the lowest wave index that has unfinished (Todo/In Progress) tasks
// and enqueues that wave. If all waves are done, it logs and returns nil.
func dispatchOnce(ctx context.Context, producer *Producer, projectRoot string) error {
	waves, err := tools.GetExecutionWaves(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("get waves: %w", err)
	}
	if len(waves) == 0 {
		log.Printf("dispatcher: no waves found (no unfinished tasks)")
		return nil
	}

	nextWave := lowestWaveIndex(waves)
	if nextWave < 0 {
		log.Printf("dispatcher: no waves to enqueue")
		return nil
	}

	enqueued, err := producer.EnqueueWave(ctx, projectRoot, nextWave)
	if err != nil {
		return fmt.Errorf("enqueue wave %d: %w", nextWave, err)
	}
	log.Printf("dispatcher: enqueued %d task(s) for wave %d", enqueued, nextWave)
	return nil
}

// lowestWaveIndex returns the smallest key from the waves map, or -1 if empty.
func lowestWaveIndex(waves map[int][]string) int {
	min := -1
	for k := range waves {
		if min < 0 || k < min {
			min = k
		}
	}
	return min
}
