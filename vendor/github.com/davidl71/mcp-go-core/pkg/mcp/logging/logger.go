// Package logging provides structured logging functionality with levels and request tracing.
// All logs are written to stderr to maintain MCP protocol compatibility (stdout is for JSON-RPC).
// Uses Go 1.21+ slog standard library for structured logging.
package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message.
// Maintained for backward compatibility with existing code.
type LogLevel int

const (
	// LevelDebug is for detailed debugging information.
	LevelDebug LogLevel = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages that may indicate issues.
	LevelWarn
	// LevelError is for error messages that indicate failures.
	LevelError
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// toSlogLevel converts LogLevel to slog.Level
func (l LogLevel) toSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Logger provides structured logging with levels, timestamps, and context.
// All logs are written to stderr to maintain MCP protocol compatibility.
// Uses slog (Go 1.21+ standard library) for structured logging.
type Logger struct {
	mu            sync.Mutex
	level         LogLevel
	slogLogger    *slog.Logger
	slowThreshold time.Duration // Threshold for performance logging
}

// NewLogger creates a new logger instance.
// The log level is determined by environment variables:
// - If MCP_DEBUG=1, log level is DEBUG (all messages)
// - If GIT_HOOK=1, log level is WARN (suppress INFO messages)
// - Otherwise, log level is INFO (DEBUG messages are suppressed)
// Output format is determined by LOG_FORMAT:
// - If LOG_FORMAT=json, uses JSON output format
// - Otherwise, uses text output format (default)
func NewLogger() *Logger {
	level := LevelInfo
	
	// Check MCP_DEBUG first (for backward compatibility)
	if os.Getenv("MCP_DEBUG") == "1" {
		level = LevelDebug
	}
	
	// GIT_HOOK overrides to WARN (suppress INFO in git hooks)
	if os.Getenv("GIT_HOOK") == "1" || strings.ToLower(os.Getenv("GIT_HOOK")) == "true" {
		level = LevelWarn
	}

	// Determine output format (JSON or text)
	format := os.Getenv("LOG_FORMAT")
	opts := &slog.HandlerOptions{
		Level: level.toSlogLevel(),
	}
	
	var handler slog.Handler
	if format == "json" {
		// Use JSONHandler for machine-readable logs
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		// Use TextHandler for human-readable output to stderr (MCP protocol compatible)
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	
	slogLogger := slog.New(handler)

	return &Logger{
		level:         level,
		slogLogger:    slogLogger,
		slowThreshold: 100 * time.Millisecond, // Log operations taking >100ms
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	// Update slog handler level
	opts := &slog.HandlerOptions{
		Level: level.toSlogLevel(),
	}
	format := os.Getenv("LOG_FORMAT")
	if format == "json" {
		l.slogLogger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
	} else {
		l.slogLogger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
}

// SetSlowThreshold sets the threshold for performance logging.
// Operations taking longer than this threshold will be logged as warnings.
func (l *Logger) SetSlowThreshold(threshold time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slowThreshold = threshold
}

// log writes a log message with the specified level, context, and message.
// Context is optional and can be used for request IDs, operation names, etc.
// Maintains backward compatibility with existing API.
func (l *Logger) log(level LogLevel, context string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if we should log this level
	if level < l.level {
		return
	}

	// Format message
	message := fmt.Sprintf(format, args...)

	// Build structured fields
	fields := []interface{}{"msg", message}
	if context != "" {
		fields = append(fields, "context", context)
	}

	// Log using slog
	switch level {
	case LevelDebug:
		l.slogLogger.Debug(message, fields...)
	case LevelInfo:
		l.slogLogger.Info(message, fields...)
	case LevelWarn:
		l.slogLogger.Warn(message, fields...)
	case LevelError:
		l.slogLogger.Error(message, fields...)
	}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(context string, format string, args ...interface{}) {
	l.log(LevelDebug, context, format, args...)
}

// Info logs an info-level message.
func (l *Logger) Info(context string, format string, args ...interface{}) {
	l.log(LevelInfo, context, format, args...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(context string, format string, args ...interface{}) {
	l.log(LevelWarn, context, format, args...)
}

// Error logs an error-level message.
func (l *Logger) Error(context string, format string, args ...interface{}) {
	l.log(LevelError, context, format, args...)
}

// LogRequest logs the start of a request with the given ID and method.
func (l *Logger) LogRequest(requestID string, method string) {
	l.Info(fmt.Sprintf("req:%s", requestID), "Processing request: %s", method)
}

// LogRequestComplete logs the completion of a request with duration.
func (l *Logger) LogRequestComplete(requestID string, method string, duration time.Duration) {
	context := fmt.Sprintf("req:%s", requestID)
	if duration > l.slowThreshold {
		l.Warn(context, "Slow request: %s took %v", method, duration)
	} else {
		l.Debug(context, "Request completed: %s took %v", method, duration)
	}
}

// LogToolCall logs a tool call with parameters.
func (l *Logger) LogToolCall(requestID string, toolName string, params interface{}) {
	l.Debug(fmt.Sprintf("req:%s", requestID), "Tool call: %s with params: %v", toolName, params)
}

// LogToolCallComplete logs the completion of a tool call with duration.
func (l *Logger) LogToolCallComplete(requestID string, toolName string, duration time.Duration) {
	context := fmt.Sprintf("req:%s", requestID)
	if duration > l.slowThreshold {
		l.Warn(context, "Slow tool call: %s took %v", toolName, duration)
	} else {
		l.Debug(context, "Tool call completed: %s took %v", toolName, duration)
	}
}

// LogError logs an error with context.
func (l *Logger) LogError(requestID string, operation string, err error) {
	l.Error(fmt.Sprintf("req:%s", requestID), "%s failed: %v", operation, err)
}

// LogPerformance logs a performance metric.
func (l *Logger) LogPerformance(context string, operation string, duration time.Duration) {
	if duration > l.slowThreshold {
		l.Warn(context, "Slow operation: %s took %v", operation, duration)
	} else {
		l.Debug(context, "Operation: %s took %v", operation, duration)
	}
}

// WithContext returns a logger that includes context information.
// Extracts request ID, operation name, and other context fields.
func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return l.slogLogger
	}
	
	logger := l.slogLogger
	
	// Extract request ID from context if available
	if requestID := getRequestID(ctx); requestID != "" {
		logger = logger.With("request_id", requestID)
	}
	
	// Extract operation name from context if available
	if operation := getOperation(ctx); operation != "" {
		logger = logger.With("operation", operation)
	}
	
	return logger
}

// requestIDKey is a private type for context keys to avoid collisions
type requestIDKey struct{}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// Check for request ID using the private key type
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	// Check for standard context keys
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

// getOperation extracts operation name from context
func getOperation(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if op, ok := ctx.Value("operation").(string); ok {
		return op
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// WithOperation adds an operation name to the context
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, "operation", operation)
}
