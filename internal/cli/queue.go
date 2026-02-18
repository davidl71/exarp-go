// Package cli: queue and worker subcommands for Redis+Asynq task queue.
package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/davidl71/exarp-go/internal/queue"
	"github.com/davidl71/exarp-go/internal/tools"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

// handleQueueCommand handles "exarp-go queue ..." (enqueue-wave N, or help).
func handleQueueCommand(parsed *mcpcli.Args) error {
	subcommand := parsed.Subcommand
	if subcommand == "" && len(parsed.Positional) > 0 {
		subcommand = parsed.Positional[0]
	}
	switch subcommand {
	case "enqueue-wave", "wave":
		return handleQueueEnqueueWave(parsed)
	case "help", "":
		fmt.Fprintln(os.Stderr, `Usage: exarp-go queue <subcommand> [options]
  enqueue-wave [N]   Enqueue all tasks in wave N (0-based). Requires REDIS_ADDR.
  wave [N]            Alias for enqueue-wave.
  help                Show this message.

Environment: REDIS_ADDR (e.g. 127.0.0.1:6379), ASYNQ_QUEUE (default: default).`)
		return nil
	default:
		return fmt.Errorf("unknown queue command: %s (use: enqueue-wave, help)", subcommand)
	}
}

func handleQueueEnqueueWave(parsed *mcpcli.Args) error {
	cfg := queue.ConfigFromEnv()
	if !cfg.Enabled() {
		return fmt.Errorf("queue not enabled: set REDIS_ADDR (e.g. 127.0.0.1:6379)")
	}
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("project root: %w", err)
	}
	waveIndex := 0
	if len(parsed.Positional) > 0 {
		// first positional might be the wave number (e.g. "queue enqueue-wave 0" or "queue wave 1")
		for _, p := range parsed.Positional {
			if n, err := strconv.Atoi(p); err == nil && n >= 0 {
				waveIndex = n
				break
			}
		}
	}
	if parsed.Subcommand == "" && len(parsed.Positional) > 0 {
		if n, err := strconv.Atoi(parsed.Positional[0]); err == nil && n >= 0 {
			waveIndex = n
		}
	}
	// Allow -wave=0 or --wave=0
	if w := parsed.GetFlag("wave", ""); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n >= 0 {
			waveIndex = n
		}
	}

	producer, err := queue.NewProducer(cfg)
	if err != nil {
		return err
	}
	defer producer.Close()

	ctx := context.Background()
	enqueued, err := producer.EnqueueWave(ctx, projectRoot, waveIndex)
	if err != nil {
		return err
	}
	fmt.Printf("Enqueued %d task(s) for wave %d\n", enqueued, waveIndex)
	return nil
}

// handleWorkerCommand runs the Asynq worker until interrupted. Requires REDIS_ADDR.
func handleWorkerCommand(parsed *mcpcli.Args) error {
	cfg := queue.ConfigFromEnv()
	if !cfg.Enabled() {
		return fmt.Errorf("worker not enabled: set REDIS_ADDR (e.g. 127.0.0.1:6379)")
	}
	ctx := context.Background()
	return queue.RunWorker(ctx, cfg)
}
