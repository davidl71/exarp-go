package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// SQLite error codes (from sqlite3.h)
const (
	SQLITE_BUSY   = 5 // The database file is locked
	SQLITE_LOCKED = 6 // A table in the database is locked
)

// maxRetries is the maximum number of retry attempts for transient errors
const maxRetries = 5

// initialBackoff is the initial backoff duration in milliseconds
const initialBackoff = 10 * time.Millisecond

// isTransientError checks if an error is a transient SQLite error that should be retried
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

// contains is a simple case-insensitive substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// indexOf finds the first occurrence of substr in s (case-insensitive)
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

// toLower converts a string to lowercase (simple implementation)
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

// retryWithBackoff executes a function with exponential backoff retry for transient errors
// It retries up to maxRetries times with exponential backoff: 10ms, 20ms, 40ms, 80ms, 160ms
func retryWithBackoff(ctx context.Context, fn func() error) error {
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

		// Calculate exponential backoff: 10ms * 2^attempt
		backoff := initialBackoff * time.Duration(math.Pow(2, float64(attempt)))

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

	// Return the last error after all retries exhausted
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// withQueryTimeout creates a context with timeout for database queries (30 seconds)
func withQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, 30*time.Second)
}

// withTransactionTimeout creates a context with timeout for database transactions (60 seconds)
func withTransactionTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, 60*time.Second)
}

// ensureContext ensures we have a valid context (uses Background if nil)
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
