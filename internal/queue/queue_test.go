package queue

import (
	"os"
	"testing"
)

func TestConfigFromEnv_Disabled(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_URL", "")

	cfg := ConfigFromEnv()
	if cfg.Enabled() {
		t.Fatal("expected disabled when REDIS_ADDR is empty")
	}
	if cfg.RedisAddr != "" {
		t.Errorf("expected empty RedisAddr, got %q", cfg.RedisAddr)
	}
}

func TestConfigFromEnv_RedisAddr(t *testing.T) {
	t.Setenv("REDIS_ADDR", "myhost:6380")
	t.Setenv("REDIS_URL", "")
	t.Setenv("ASYNQ_QUEUE", "")
	t.Setenv("ASYNQ_CONCURRENCY", "")

	cfg := ConfigFromEnv()
	if !cfg.Enabled() {
		t.Fatal("expected enabled when REDIS_ADDR is set")
	}
	if cfg.RedisAddr != "myhost:6380" {
		t.Errorf("RedisAddr = %q, want %q", cfg.RedisAddr, "myhost:6380")
	}
	if cfg.QueueName != DefaultQueueName {
		t.Errorf("QueueName = %q, want %q", cfg.QueueName, DefaultQueueName)
	}
	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5", cfg.Concurrency)
	}
}

func TestConfigFromEnv_RedisURL(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_URL", "redis://fallback:6379")

	cfg := ConfigFromEnv()
	if !cfg.Enabled() {
		t.Fatal("expected enabled when REDIS_URL is set")
	}
	if cfg.RedisAddr != "redis://fallback:6379" {
		t.Errorf("RedisAddr = %q, want %q", cfg.RedisAddr, "redis://fallback:6379")
	}
}

func TestConfigFromEnv_CustomQueueAndConcurrency(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("ASYNQ_QUEUE", "exarp-prod")
	t.Setenv("ASYNQ_CONCURRENCY", "10")

	cfg := ConfigFromEnv()
	if cfg.QueueName != "exarp-prod" {
		t.Errorf("QueueName = %q, want %q", cfg.QueueName, "exarp-prod")
	}
	if cfg.Concurrency != 10 {
		t.Errorf("Concurrency = %d, want 10", cfg.Concurrency)
	}
}

func TestConfigFromEnv_InvalidConcurrency(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("ASYNQ_CONCURRENCY", "notanumber")

	cfg := ConfigFromEnv()
	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5 (default on parse error)", cfg.Concurrency)
	}
}

func TestConfigFromEnv_ZeroConcurrency(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("ASYNQ_CONCURRENCY", "0")

	cfg := ConfigFromEnv()
	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5 (default on zero)", cfg.Concurrency)
	}
}

func TestConfigFromEnv_RedisAddrPriority(t *testing.T) {
	t.Setenv("REDIS_ADDR", "primary:6379")
	t.Setenv("REDIS_URL", "fallback:6379")

	cfg := ConfigFromEnv()
	if cfg.RedisAddr != "primary:6379" {
		t.Errorf("REDIS_ADDR should take priority over REDIS_URL: got %q", cfg.RedisAddr)
	}
}

func TestPayloadMarshalRoundtrip(t *testing.T) {
	original := &TaskExecutePayload{
		TaskID:      "T-1234567890",
		ProjectRoot: "/home/user/project",
	}
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	restored, err := UnmarshalTaskExecutePayload(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if restored.TaskID != original.TaskID {
		t.Errorf("TaskID = %q, want %q", restored.TaskID, original.TaskID)
	}
	if restored.ProjectRoot != original.ProjectRoot {
		t.Errorf("ProjectRoot = %q, want %q", restored.ProjectRoot, original.ProjectRoot)
	}
}

func TestPayloadUnmarshalInvalid(t *testing.T) {
	_, err := UnmarshalTaskExecutePayload([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPayloadMarshalEmpty(t *testing.T) {
	p := &TaskExecutePayload{}
	data, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal empty: %v", err)
	}
	restored, err := UnmarshalTaskExecutePayload(data)
	if err != nil {
		t.Fatalf("Unmarshal empty: %v", err)
	}
	if restored.TaskID != "" || restored.ProjectRoot != "" {
		t.Errorf("expected empty fields, got %+v", restored)
	}
}

func TestProducerEnqueueTask_Disabled(t *testing.T) {
	cfg := Config{}
	producer, err := NewProducer(cfg)
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer producer.Close()

	enqueued, err := producer.EnqueueTask(t.Context(), "T-1", "/tmp/proj")
	if err != nil {
		t.Fatalf("EnqueueTask should not error when disabled: %v", err)
	}
	if enqueued != 0 {
		t.Errorf("enqueued = %d, want 0 (disabled)", enqueued)
	}
}

func TestProducerEnqueueWave_Disabled(t *testing.T) {
	cfg := Config{}
	producer, err := NewProducer(cfg)
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer producer.Close()

	enqueued, err := producer.EnqueueWave(t.Context(), "/tmp/proj", 0)
	if err != nil {
		t.Fatalf("EnqueueWave should not error when disabled: %v", err)
	}
	if enqueued != 0 {
		t.Errorf("enqueued = %d, want 0 (disabled)", enqueued)
	}
}

func TestProducerCloseNil(t *testing.T) {
	var p *Producer
	if err := p.Close(); err != nil {
		t.Errorf("Close on nil producer should not error: %v", err)
	}
}

func TestProducerEnqueueTask_NilProducer(t *testing.T) {
	var p *Producer
	enqueued, err := p.EnqueueTask(t.Context(), "T-1", "/tmp")
	if err != nil {
		t.Errorf("EnqueueTask on nil producer should not error: %v", err)
	}
	if enqueued != 0 {
		t.Errorf("enqueued = %d, want 0", enqueued)
	}
}

// Integration tests requiring Redis are skipped unless REDIS_ADDR is set.
func TestProducerEnqueueTask_Integration(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set, skipping integration test")
	}
	cfg := ConfigFromEnv()
	producer, err := NewProducer(cfg)
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer producer.Close()

	enqueued, err := producer.EnqueueTask(t.Context(), "T-integration-test", "/tmp/proj")
	if err != nil {
		t.Fatalf("EnqueueTask: %v", err)
	}
	if enqueued != 1 {
		t.Errorf("enqueued = %d, want 1", enqueued)
	}
}
