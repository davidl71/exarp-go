// Package cli provides a logging adapter that uses exarp-go's shared logger
// (internal/logging) so CLI, main, and database share one logger instance.
package cli

import (
	"context"
	"time"

	"github.com/davidl71/exarp-go/internal/logging"
)

// DefaultSlowThreshold is the default threshold for slow operation detection (2 seconds).
var DefaultSlowThreshold = 2 * time.Second

// logInfo logs an info message with structured fields using the shared logger.
func logInfo(ctx context.Context, msg string, args ...interface{}) {
	logging.Default().WithContext(ctx).Info(msg, args...)
}

// logWarn logs a warning message with structured fields using the shared logger.
func logWarn(ctx context.Context, msg string, args ...interface{}) {
	logging.Default().WithContext(ctx).Warn(msg, args...)
}

// logError logs an error message with structured fields using the shared logger.
func logError(ctx context.Context, msg string, args ...interface{}) {
	logging.Default().WithContext(ctx).Error(msg, args...)
}

// logDebug logs a debug message with structured fields using the shared logger.
func logDebug(ctx context.Context, msg string, args ...interface{}) {
	logging.Default().WithContext(ctx).Debug(msg, args...)
}

// PerformanceLogger tracks operation duration and logs slow operations.
type PerformanceLogger struct {
	operation string
	start     time.Time
	ctx       context.Context
	threshold time.Duration
}

// StartPerformanceLogging starts tracking performance for an operation.
func StartPerformanceLogging(ctx context.Context, operation string, threshold time.Duration) *PerformanceLogger {
	logging.SetSlowThreshold(threshold)

	return &PerformanceLogger{
		operation: operation,
		start:     time.Now(),
		ctx:       ctx,
		threshold: threshold,
	}
}

// Finish logs the operation duration and warns if it exceeded the threshold.
func (pl *PerformanceLogger) Finish() {
	duration := time.Since(pl.start)
	log := logging.Default()

	// Build context string for LogPerformance
	contextStr := ""

	if pl.ctx != nil {
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

	log.LogPerformance(contextStr, pl.operation, duration)

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

// FinishWithError logs the operation duration and error.
func (pl *PerformanceLogger) FinishWithError(err error) {
	duration := time.Since(pl.start)
	logError(pl.ctx, "Operation failed",
		"operation", pl.operation,
		"error", err,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
}
