// Package cli provides a logging adapter to bridge exarp-go's logging API
// with mcp-go-core's logger. This allows gradual migration while maintaining
// the same API surface.
package cli

import (
	"context"
	"time"

	"github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

var (
	// logger is the shared mcp-go-core logger instance
	logger *logging.Logger

	// DefaultSlowThreshold is the default threshold for slow operation detection (2 seconds)
	DefaultSlowThreshold = 2 * time.Second
)

// initLogger initializes the mcp-go-core logger if not already initialized
func initLogger() {
	if logger == nil {
		logger = logging.NewLogger()
		logger.SetSlowThreshold(DefaultSlowThreshold)
	}
}

// Info logs an info message with structured fields
// Adapts exarp-go's logging.Info(msg, args...) to mcp-go-core's logger
func logInfo(ctx context.Context, msg string, args ...interface{}) {
	initLogger()
	
	// Always use slog logger for structured logging (supports key-value pairs)
	slogLogger := logger.WithContext(ctx)
	slogLogger.Info(msg, args...)
}

// Warn logs a warning message with structured fields
func logWarn(ctx context.Context, msg string, args ...interface{}) {
	initLogger()
	
	slogLogger := logger.WithContext(ctx)
	slogLogger.Warn(msg, args...)
}

// Error logs an error message with structured fields
func logError(ctx context.Context, msg string, args ...interface{}) {
	initLogger()
	
	slogLogger := logger.WithContext(ctx)
	slogLogger.Error(msg, args...)
}

// Debug logs a debug message with structured fields
func logDebug(ctx context.Context, msg string, args ...interface{}) {
	initLogger()
	
	slogLogger := logger.WithContext(ctx)
	slogLogger.Debug(msg, args...)
}

// PerformanceLogger tracks operation duration and logs slow operations
type PerformanceLogger struct {
	operation string
	start     time.Time
	ctx       context.Context
	threshold time.Duration
}

// StartPerformanceLogging starts tracking performance for an operation
func StartPerformanceLogging(ctx context.Context, operation string, threshold time.Duration) *PerformanceLogger {
	return &PerformanceLogger{
		operation: operation,
		start:     time.Now(),
		ctx:       ctx,
		threshold: threshold,
	}
}

// Finish logs the operation duration and warns if it exceeded the threshold
func (pl *PerformanceLogger) Finish() {
	duration := time.Since(pl.start)
	initLogger()
	
	// Build context string for LogPerformance
	contextStr := ""
	if pl.ctx != nil {
		// Extract operation and request ID from context
		type requestIDKey struct{}
		if reqID, ok := pl.ctx.Value(requestIDKey{}).(string); ok {
			contextStr = "req:" + reqID
		}
		if op, ok := pl.ctx.Value("operation").(string); ok {
			if contextStr != "" {
				contextStr = contextStr + " operation:" + op
			} else {
				contextStr = "operation:" + op
			}
		}
	}
	
	// Always log duration using LogPerformance
	logger.LogPerformance(contextStr, pl.operation, duration)
	
	// Also log with structured fields if slow
	if pl.threshold > 0 && duration > pl.threshold {
		logWarn(pl.ctx, "Slow operation detected",
			"operation", pl.operation,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
			"threshold_ms", pl.threshold.Milliseconds(),
			"threshold", pl.threshold.String(),
			"exceeded_by_ms", (duration - pl.threshold).Milliseconds(),
		)
	}
}

// FinishWithError logs the operation duration and error
func (pl *PerformanceLogger) FinishWithError(err error) {
	duration := time.Since(pl.start)
	logError(pl.ctx, "Operation failed",
		"operation", pl.operation,
		"error", err,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
}
