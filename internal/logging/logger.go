// Package logging provides a single logging facade for exarp-go.
// It wraps mcp-go-core's logger (which uses slog) so main, database, and cli
// share one logger instance and one configuration (level, GIT_HOOK, LOG_FORMAT).
package logging

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	mcplog "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

// Logger is the mcp-go-core logger (uses slog). Re-exported so callers can use *logging.Logger.
type Logger = mcplog.Logger

var (
	defaultLogger *mcplog.Logger
	initOnce      sync.Once
)

// Init initializes the default logger. Safe to call multiple times; only the first call runs.
// Main should call Init() at startup so all packages use the same logger.
func Init() {
	initOnce.Do(func() {
		defaultLogger = mcplog.NewLogger()
		defaultLogger.SetLevel(mcplog.LevelWarn)
	})
}

// Default returns the shared logger. Call Init() before first use (e.g. from main).
// If never initialized, initializes on first call so database/cli work without explicit Init.
func Default() *Logger {
	initOnce.Do(func() {
		defaultLogger = mcplog.NewLogger()
		defaultLogger.SetLevel(mcplog.LevelWarn)
	})

	return defaultLogger
}

// SetSlowThreshold sets the slow-operation threshold on the default logger.
func SetSlowThreshold(threshold time.Duration) {
	Default().SetSlowThreshold(threshold)
}

// ConfigureFromConfig applies logging configuration from centralized config.
// Call after config.SetGlobalConfig when config is loaded (e.g. in EnsureConfigAndDatabase).
// Level mapping: debug, info, warn, error (case-insensitive). Format still comes from LOG_FORMAT env.
func ConfigureFromConfig(cfg config.LoggingConfig) {
	l := Default()

	level := mcplog.LevelInfo
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = mcplog.LevelDebug
	case "info":
		level = mcplog.LevelInfo
	case "warn", "warning":
		level = mcplog.LevelWarn
	case "error":
		level = mcplog.LevelError
	}

	l.SetLevel(level)
}

// Warn logs a warning using the default logger with context string "".
func Warn(format string, args ...interface{}) {
	Default().Warn("", format, args...)
}

// Error logs an error using the default logger with context string "".
func Error(format string, args ...interface{}) {
	Default().Error("", format, args...)
}

// Fatal logs an error and exits with code 1. Use for startup failures.
func Fatal(format string, args ...interface{}) {
	Default().Error("", format, args...)
	os.Exit(1)
}
