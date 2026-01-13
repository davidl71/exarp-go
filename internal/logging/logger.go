package logging

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// Logger is a unified logger wrapper that uses slog
var Logger *slog.Logger

// InitLogger initializes the unified logger
// If GIT_HOOK=1 is set, log level is set to WARN to suppress INFO messages
// If LOG_FORMAT=json is set, uses JSON output format
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: getLogLevel(),
	}

	// Determine output format (JSON or text)
	format := os.Getenv("LOG_FORMAT")
	if format == "json" {
		// Use JSONHandler for machine-readable logs
		handler := slog.NewJSONHandler(os.Stderr, opts)
		Logger = slog.New(handler)
	} else {
		// Use TextHandler for human-readable output to stderr (MCP protocol compatible)
		handler := slog.NewTextHandler(os.Stderr, opts)
		Logger = slog.New(handler)
	}
}

// getLogLevel returns the appropriate log level based on environment
func getLogLevel() slog.Level {
	gitHook := os.Getenv("GIT_HOOK")
	if gitHook == "1" || gitHook == "true" {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

// Info logs an info message with structured fields
func Info(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Info(msg, args...)
}

// InfoContext logs an info message with context and structured fields
func InfoContext(ctx context.Context, msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning message with structured fields
func Warn(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Warn(msg, args...)
}

// WarnContext logs a warning message with context and structured fields
func WarnContext(ctx context.Context, msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.WarnContext(ctx, msg, args...)
}

// Error logs an error message with structured fields
func Error(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Error(msg, args...)
}

// ErrorContext logs an error message with context and structured fields
func ErrorContext(ctx context.Context, msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.ErrorContext(ctx, msg, args...)
}

// Debug logs a debug message with structured fields
func Debug(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Debug(msg, args...)
}

// DebugContext logs a debug message with context and structured fields
func DebugContext(ctx context.Context, msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.DebugContext(ctx, msg, args...)
}

// With returns a logger that includes the given attributes
func With(args ...any) *slog.Logger {
	if Logger == nil {
		InitLogger()
	}
	return Logger.With(args...)
}

// WithContext returns a logger that includes context information
// Extracts request ID, operation name, and other context fields
func WithContext(ctx context.Context) *slog.Logger {
	if Logger == nil {
		InitLogger()
	}
	logger := Logger
	
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

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// Check for common request ID context keys
	type requestIDKey struct{}
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
	type requestIDKey struct{}
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// WithOperation adds an operation name to the context
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, "operation", operation)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	return getRequestID(ctx)
}

// GetOperation extracts operation name from context
func GetOperation(ctx context.Context) string {
	return getOperation(ctx)
}

// PerformanceLogger tracks operation duration and logs slow operations
type PerformanceLogger struct {
	operation string
	start     time.Time
	ctx       context.Context
	threshold time.Duration
}

// StartPerformanceLogging starts tracking performance for an operation
// Automatically logs slow operations if duration exceeds threshold
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
	
	logger := WithContext(pl.ctx)
	
	// Always log duration for performance tracking
	logger.Info("Operation completed",
		"operation", pl.operation,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
	
	// Warn if operation was slow
	if pl.threshold > 0 && duration > pl.threshold {
		logger.Warn("Slow operation detected",
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
	
	logger := WithContext(pl.ctx)
	
	logger.Error("Operation failed",
		"operation", pl.operation,
		"error", err,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
}

// LogDuration logs operation duration with optional slow operation detection
func LogDuration(ctx context.Context, operation string, duration time.Duration, threshold time.Duration) {
	logger := WithContext(ctx)
	
	logger.Info("Operation duration",
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"duration", duration.String(),
	)
	
	if threshold > 0 && duration > threshold {
		logger.Warn("Slow operation",
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
			"threshold_ms", threshold.Milliseconds(),
			"threshold", threshold.String(),
		)
	}
}

// DefaultSlowThreshold is the default threshold for slow operation detection (2 seconds)
const DefaultSlowThreshold = 2 * time.Second

// LogSlowOperation logs a warning if the operation took longer than the threshold
func LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration) {
	if threshold == 0 {
		threshold = DefaultSlowThreshold
	}
	
	if duration > threshold {
		logger := WithContext(ctx)
		logger.Warn("Slow operation",
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
			"threshold_ms", threshold.Milliseconds(),
			"threshold", threshold.String(),
			"exceeded_by_ms", (duration - threshold).Milliseconds(),
		)
	}
}
