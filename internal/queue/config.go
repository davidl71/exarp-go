// Package queue provides optional Redis+Asynq task queue for distributed execution.
// Queue is opt-in: when REDIS_ADDR is not set, producer and worker are no-ops.
package queue

import (
	"os"
	"strconv"
)

const (
	// TypeTaskExecute is the Asynq task type for running task_execute (one Todo2 task).
	TypeTaskExecute = "exarp:task_execute"
	// DefaultQueueName is the default Asynq queue name when ASYNQ_QUEUE is not set.
	DefaultQueueName = "default"
	// DefaultRedisAddr is the default Redis address when REDIS_ADDR is not set.
	DefaultRedisAddr = "127.0.0.1:6379"
)

// Config holds queue/Redis settings from environment.
type Config struct {
	RedisAddr string
	QueueName string
	Concurrency int
}

// ConfigFromEnv returns queue config from environment. Empty RedisAddr means queue is disabled.
func ConfigFromEnv() Config {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = os.Getenv("REDIS_URL") // common alternative
	}
	if addr == "" {
		return Config{}
	}
	queue := os.Getenv("ASYNQ_QUEUE")
	if queue == "" {
		queue = DefaultQueueName
	}
	concurrency := 5
	if s := os.Getenv("ASYNQ_CONCURRENCY"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			concurrency = n
		}
	}
	return Config{
		RedisAddr:   addr,
		QueueName:   queue,
		Concurrency: concurrency,
	}
}

// Enabled returns true if Redis is configured.
func (c Config) Enabled() bool {
	return c.RedisAddr != ""
}
