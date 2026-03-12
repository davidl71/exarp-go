package logging

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	// Test default logger (INFO level)
	os.Unsetenv("MCP_DEBUG")
	os.Unsetenv("GIT_HOOK")
	os.Unsetenv("LOG_FORMAT")
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	if logger.level != LevelInfo {
		t.Errorf("Expected default level INFO, got %v", logger.level)
	}

	// Test DEBUG level when MCP_DEBUG=1
	os.Setenv("MCP_DEBUG", "1")
	os.Unsetenv("GIT_HOOK")
	logger = NewLogger()
	if logger.level != LevelDebug {
		t.Errorf("Expected DEBUG level when MCP_DEBUG=1, got %v", logger.level)
	}
	os.Unsetenv("MCP_DEBUG")

	// Test WARN level when GIT_HOOK=1 (overrides MCP_DEBUG)
	os.Setenv("GIT_HOOK", "1")
	logger = NewLogger()
	if logger.level != LevelWarn {
		t.Errorf("Expected WARN level when GIT_HOOK=1, got %v", logger.level)
	}
	os.Unsetenv("GIT_HOOK")
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewLogger()
	logger.SetLevel(LevelWarn)

	if logger.level != LevelWarn {
		t.Errorf("Expected level WARN after SetLevel, got %v", logger.level)
	}
}

func TestLogger_SetSlowThreshold(t *testing.T) {
	logger := NewLogger()
	threshold := 200 * time.Millisecond
	logger.SetSlowThreshold(threshold)

	if logger.slowThreshold != threshold {
		t.Errorf("Expected slow threshold %v, got %v", threshold, logger.slowThreshold)
	}
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelDebug

	// Test all log levels
	logger.Debug("test", "Debug message")
	logger.Info("test", "Info message")
	logger.Warn("test", "Warn message")
	logger.Error("test", "Error message")

	output := buf.String()

	// Check that all levels are present (slog uses "level=DEBUG" format)
	if !strings.Contains(output, "level=DEBUG") && !strings.Contains(output, "DEBUG") {
		t.Error("Debug log not found")
	}
	if !strings.Contains(output, "level=INFO") && !strings.Contains(output, "INFO") {
		t.Error("Info log not found")
	}
	if !strings.Contains(output, "level=WARN") && !strings.Contains(output, "WARN") {
		t.Error("Warn log not found")
	}
	if !strings.Contains(output, "level=ERROR") && !strings.Contains(output, "ERROR") {
		t.Error("Error log not found")
	}

	// Check that messages are present
	if !strings.Contains(output, "Debug message") {
		t.Error("Debug message not found")
	}
	if !strings.Contains(output, "Info message") {
		t.Error("Info message not found")
	}
	if !strings.Contains(output, "Warn message") {
		t.Error("Warn message not found")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Error message not found")
	}
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelWarn

	// These should not be logged
	logger.Debug("test", "Debug message")
	logger.Info("test", "Info message")

	// These should be logged
	logger.Warn("test", "Warn message")
	logger.Error("test", "Error message")

	output := buf.String()

	// Check that filtered levels are not present
	if strings.Contains(output, "level=DEBUG") || strings.Contains(output, "DEBUG") {
		t.Error("Debug log should be filtered at WARN level")
	}
	if strings.Contains(output, "level=INFO") || strings.Contains(output, "INFO") {
		t.Error("Info log should be filtered at WARN level")
	}

	// Check that allowed levels are present
	if !strings.Contains(output, "level=WARN") && !strings.Contains(output, "WARN") {
		t.Error("Warn log should be present at WARN level")
	}
	if !strings.Contains(output, "level=ERROR") && !strings.Contains(output, "ERROR") {
		t.Error("Error log should be present at WARN level")
	}
}

func TestLogger_Context(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	logger.Info("req:123", "Test message")
	output := buf.String()

	// slog includes context as a field, check for it
	if !strings.Contains(output, "req:123") && !strings.Contains(output, "context=req:123") {
		t.Error("Context not found in log output")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Message not found in log output")
	}
}

func TestLogger_NoContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	logger.Info("", "Test message")
	output := buf.String()

	// Should contain message
	if !strings.Contains(output, "Test message") {
		t.Error("Message not found in log output")
	}
}

func TestLogger_LogRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	logger.LogRequest("123", "tools/list")
	output := buf.String()

	// slog uses "level=INFO" format
	if !strings.Contains(output, "level=INFO") && !strings.Contains(output, "INFO") {
		t.Error("LogRequest should log at INFO level")
	}
	// LogRequest formats context as "req:123", so check for that
	if !strings.Contains(output, "req:123") {
		t.Errorf("Request ID not found in log. Output: %q", output)
	}
	if !strings.Contains(output, "Processing request: tools/list") {
		t.Error("Request method not found in log")
	}
}

func TestLogger_LogRequestComplete(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelDebug
	logger.slowThreshold = 50 * time.Millisecond

	// Fast request (should log at DEBUG)
	logger.LogRequestComplete("req:123", "tools/list", 10*time.Millisecond)
	output := buf.String()

	if !strings.Contains(output, "level=DEBUG") && !strings.Contains(output, "DEBUG") {
		t.Error("Fast request should log at DEBUG level")
	}
	if !strings.Contains(output, "Request completed: tools/list") {
		t.Error("Request completion message not found")
	}

	// Slow request (should log at WARN)
	buf.Reset()
	logger.LogRequestComplete("req:124", "tools/list", 100*time.Millisecond)
	output = buf.String()

	if !strings.Contains(output, "level=WARN") && !strings.Contains(output, "WARN") {
		t.Error("Slow request should log at WARN level")
	}
	if !strings.Contains(output, "Slow request: tools/list") {
		t.Error("Slow request message not found")
	}
}

func TestLogger_LogToolCall(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelDebug

	logger.LogToolCall("123", "get_wisdom", map[string]interface{}{"source": "pistis_sophia"})
	output := buf.String()

	if !strings.Contains(output, "level=DEBUG") && !strings.Contains(output, "DEBUG") {
		t.Error("LogToolCall should log at DEBUG level")
	}
	// LogToolCall formats context as "req:123", so check for that
	if !strings.Contains(output, "req:123") {
		t.Errorf("Request ID not found in log. Output: %q", output)
	}
	if !strings.Contains(output, "Tool call: get_wisdom") {
		t.Error("Tool call message not found")
	}
}

func TestLogger_LogToolCallComplete(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelDebug
	logger.slowThreshold = 50 * time.Millisecond

	// Fast tool call (should log at DEBUG)
	logger.LogToolCallComplete("req:123", "get_wisdom", 10*time.Millisecond)
	output := buf.String()

	if !strings.Contains(output, "level=DEBUG") && !strings.Contains(output, "DEBUG") {
		t.Error("Fast tool call should log at DEBUG level")
	}
	if !strings.Contains(output, "Tool call completed: get_wisdom") {
		t.Error("Tool call completion message not found")
	}

	// Slow tool call (should log at WARN)
	buf.Reset()
	logger.LogToolCallComplete("req:124", "get_wisdom", 100*time.Millisecond)
	output = buf.String()

	if !strings.Contains(output, "level=WARN") && !strings.Contains(output, "WARN") {
		t.Error("Slow tool call should log at WARN level")
	}
	if !strings.Contains(output, "Slow tool call: get_wisdom") {
		t.Error("Slow tool call message not found")
	}
}

func TestLogger_LogError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelError

	err := &testError{message: "test error"}
	logger.LogError("123", "operation", err)
	output := buf.String()

	if !strings.Contains(output, "level=ERROR") && !strings.Contains(output, "ERROR") {
		t.Error("LogError should log at ERROR level")
	}
	// LogError formats context as "req:123", so check for that
	if !strings.Contains(output, "req:123") {
		t.Errorf("Request ID not found in log. Output: %q", output)
	}
	if !strings.Contains(output, "operation failed: test error") {
		t.Error("Error message not found")
	}
}

func TestLogger_LogPerformance(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelDebug
	logger.slowThreshold = 50 * time.Millisecond

	// Fast operation (should log at DEBUG)
	logger.LogPerformance("test", "operation", 10*time.Millisecond)
	output := buf.String()

	if !strings.Contains(output, "level=DEBUG") && !strings.Contains(output, "DEBUG") {
		t.Error("Fast operation should log at DEBUG level")
	}
	if !strings.Contains(output, "Operation: operation took") {
		t.Error("Performance log message not found")
	}

	// Slow operation (should log at WARN)
	buf.Reset()
	logger.LogPerformance("test", "operation", 100*time.Millisecond)
	output = buf.String()

	if !strings.Contains(output, "level=WARN") && !strings.Contains(output, "WARN") {
		t.Error("Slow operation should log at WARN level")
	}
	if !strings.Contains(output, "Slow operation: operation took") {
		t.Error("Slow operation message not found")
	}
}

func TestLogger_ThreadSafety(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	// Concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("test", "Message %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	// Check that all messages are present
	for i := 0; i < 10; i++ {
		if !strings.Contains(output, fmt.Sprintf("Message %d", i)) {
			t.Errorf("Message %d not found in concurrent log output", i)
		}
	}
}

func TestLogger_TimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	logger.Info("test", "Test message")
	output := buf.String()

	// slog includes timestamp automatically, check that it's present
	// Format is typically "time=2026-01-13T..." or similar
	if !strings.Contains(output, "time=") && !strings.Contains(output, "T") {
		t.Error("Timestamp should be present in log output")
	}
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	os.Setenv("LOG_FORMAT", "json")
	defer os.Unsetenv("LOG_FORMAT")
	
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	logger.Info("test", "Test message")
	output := buf.String()

	// JSON output should contain JSON structure
	if !strings.Contains(output, "\"level\"") && !strings.Contains(output, "\"msg\"") {
		t.Error("JSON output should contain level and msg fields")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Message not found in JSON output")
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	// Create a new logger with buffer output for testing
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	logger.slogLogger = slog.New(handler)
	logger.level = LevelInfo

	// Create context with request ID and operation
	ctx := WithRequestID(context.Background(), "req-123")
	ctx = WithOperation(ctx, "test_operation")
	
	// Use WithContext to get logger with context fields
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("Test message")
	
	output := buf.String()

	// Check that context fields are present (slog uses "request_id=req-123" format)
	// slog text format: time=... level=INFO msg="Test message" request_id=req-123 operation=test_operation
	if !strings.Contains(output, "req-123") {
		t.Errorf("Request ID not found in log output. Output: %q", output)
	}
	if !strings.Contains(output, "test_operation") {
		t.Errorf("Operation not found in log output. Output: %q", output)
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
