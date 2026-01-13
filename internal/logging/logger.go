package logging

import (
	"log/slog"
	"os"
)

// Logger is a unified logger wrapper that uses slog
var Logger *slog.Logger

// InitLogger initializes the unified logger
// If GIT_HOOK=1 is set, log level is set to WARN to suppress INFO messages
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: getLogLevel(),
	}

	// Use TextHandler for human-readable output to stderr (MCP protocol compatible)
	handler := slog.NewTextHandler(os.Stderr, opts)
	Logger = slog.New(handler)
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

// Warn logs a warning message with structured fields
func Warn(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Warn(msg, args...)
}

// Error logs an error message with structured fields
func Error(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Error(msg, args...)
}

// Debug logs a debug message with structured fields
func Debug(msg string, args ...any) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Debug(msg, args...)
}
