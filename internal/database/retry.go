// retry.go — Database operation retry with exponential backoff for transient errors.
package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
)

// SQLite error codes (from sqlite3.h).
const (
	SQLITE_BUSY   = 5 // The database file is locked
	SQLITE_LOCKED = 6 // A table in the database is locked
)

// isTransientError checks if an error is a transient SQLite error that should be retried.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for SQLite error codes
	errStr := err.Error()
	// SQLite errors typically include the error code in the message
	// e.g., "database is locked (5)" or "database is locked (SQLITE_BUSY)"
	if contains(errStr, "database is locked") ||
		contains(errStr, "database locked") ||
		contains(errStr, "SQLITE_BUSY") ||
		contains(errStr, "SQLITE_LOCKED") ||
		contains(errStr, "(5)") ||
		contains(errStr, "(6)") {
		return true
	}

	// Check for context timeout errors (these are not transient, don't retry)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	return false
}

// contains is a simple case-insensitive substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// indexOf finds the first occurrence of substr in s (case-insensitive).
func indexOf(s, substr string) int {
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return i
		}
	}

	return -1
}

// toLower converts a string to lowercase (simple implementation).
func toLower(s string) string {
	result := make([]byte, len(s))

	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}

	return string(result)
}

// retryWithBackoff executes a function with exponential backoff retry for transient errors.
// Uses config for retry attempts, initial delay, max delay, and multiplier when available.
func retryWithBackoff(ctx context.Context, fn func() error) error {
	cfg := config.GetGlobalConfig()

	maxRetries := cfg.Database.RetryAttempts
	if maxRetries <= 0 {
		maxRetries = 3
	}

	initialDelay := cfg.Database.RetryInitialDelay
	if initialDelay <= 0 {
		initialDelay = 100 * time.Millisecond
	}

	maxDelay := cfg.Database.RetryMaxDelay
	if maxDelay <= 0 {
		maxDelay = 5 * time.Second
	}

	multiplier := cfg.Database.RetryMultiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context cancellation before each attempt
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry if it's not a transient error
		if !isTransientError(err) {
			return err
		}

		// Don't retry on the last attempt
		if attempt == maxRetries-1 {
			break
		}

		// Exponential backoff: initialDelay * multiplier^attempt, capped at maxDelay
		backoff := time.Duration(float64(initialDelay) * math.Pow(multiplier, float64(attempt)))
		if backoff > maxDelay {
			backoff = maxDelay
		}

		// Create a timer with context cancellation support
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-timer.C:
			// Continue to next retry attempt
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// withQueryTimeout creates a context with timeout for database queries.
// Uses config.DatabaseConnectionTimeout when available (default 30s).
func withQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := config.DatabaseConnectionTimeout()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return context.WithTimeout(ctx, timeout)
}

// withTransactionTimeout creates a context with timeout for database transactions.
// Uses config.DatabaseQueryTimeout when available (default 60s).
func withTransactionTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := config.DatabaseQueryTimeout()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return context.WithTimeout(ctx, timeout)
}

// ensureContext ensures we have a valid context (uses Background if nil).
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
