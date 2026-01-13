// Package logging provides structured logging functionality with levels and request tracing.
// All logs are written to stderr to maintain MCP protocol compatibility (stdout is for JSON-RPC).
package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message.
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

// Logger provides structured logging with levels, timestamps, and context.
// All logs are written to stderr to maintain MCP protocol compatibility.
type Logger struct {
	mu            sync.Mutex
	level         LogLevel
	output        io.Writer
	slowThreshold time.Duration // Threshold for performance logging
}

// NewLogger creates a new logger instance.
// The log level is determined by the MCP_DEBUG environment variable:
// - If MCP_DEBUG=1, log level is DEBUG (all messages)
// - Otherwise, log level is INFO (DEBUG messages are suppressed)
func NewLogger() *Logger {
	level := LevelInfo
	if os.Getenv("MCP_DEBUG") == "1" {
		level = LevelDebug
	}

	return &Logger{
		level:         level,
		output:        os.Stderr,
		slowThreshold: 100 * time.Millisecond, // Log operations taking >100ms
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
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
func (l *Logger) log(level LogLevel, context string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if we should log this level
	if level < l.level {
		return
	}

	// Format timestamp
	timestamp := time.Now().Format(time.RFC3339)

	// Format message
	message := fmt.Sprintf(format, args...)

	// Build log line: [LEVEL] [TIMESTAMP] [CONTEXT] message
	var logLine string
	if context != "" {
		logLine = fmt.Sprintf("[%s] [%s] [%s] %s\n", level.String(), timestamp, context, message)
	} else {
		logLine = fmt.Sprintf("[%s] [%s] %s\n", level.String(), timestamp, message)
	}

	// Write to stderr
	fmt.Fprint(l.output, logLine)
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
