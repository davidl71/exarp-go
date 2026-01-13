package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	mu          sync.RWMutex
	requests    map[string][]time.Time // client -> request timestamps
	window      time.Duration          // time window
	maxRequests int                    // max requests per window
	cleanup     *time.Ticker           // periodic cleanup
	stopCleanup chan struct{}
}

// NewRateLimiter creates a new rate limiter
// window: time window (e.g., 1 minute)
// maxRequests: maximum requests allowed in the window
func NewRateLimiter(window time.Duration, maxRequests int) *RateLimiter {
	rl := &RateLimiter{
		requests:    make(map[string][]time.Time),
		window:      window,
		maxRequests: maxRequests,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine to remove old entries
	rl.cleanup = time.NewTicker(window)
	go rl.cleanupOldEntries()

	return rl
}

// Allow checks if a request from the given client should be allowed
// Returns true if allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this client
	requests, exists := rl.requests[clientID]
	if !exists {
		requests = make([]time.Time, 0, rl.maxRequests)
	}

	// Remove requests outside the window
	validRequests := make([]time.Time, 0, len(requests))
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we've exceeded the limit
	if len(validRequests) >= rl.maxRequests {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[clientID] = validRequests

	return true
}

// Wait blocks until a request can be made (or context expires)
func (rl *RateLimiter) Wait(ctx context.Context, clientID string) error {
	for {
		if rl.Allow(clientID) {
			return nil
		}

		// Calculate when the oldest request will expire
		rl.mu.RLock()
		requests := rl.requests[clientID]
		var waitTime time.Duration
		if len(requests) > 0 {
			oldest := requests[0]
			waitTime = rl.window - time.Since(oldest)
			if waitTime < 0 {
				waitTime = 0
			}
		}
		rl.mu.RUnlock()

		// Wait for the oldest request to expire or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// cleanupOldEntries periodically removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanupOldEntries() {
	for {
		select {
		case <-rl.stopCleanup:
			return
		case <-rl.cleanup.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.window)
			for clientID, requests := range rl.requests {
				validRequests := make([]time.Time, 0)
				for _, reqTime := range requests {
					if reqTime.After(cutoff) {
						validRequests = append(validRequests, reqTime)
					}
				}
				if len(validRequests) == 0 {
					delete(rl.requests, clientID)
				} else {
					rl.requests[clientID] = validRequests
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Stop stops the rate limiter and cleans up resources
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
	close(rl.stopCleanup)
}

// GetRemaining returns the number of remaining requests for a client
func (rl *RateLimiter) GetRemaining(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	requests := rl.requests[clientID]
	cutoff := time.Now().Add(-rl.window)
	count := 0
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			count++
		}
	}
	return rl.maxRequests - count
}

// DefaultRateLimiter is the default rate limiter instance
var (
	defaultRateLimiter *RateLimiter
	once               sync.Once
)

// GetDefaultRateLimiter returns the default rate limiter
// Default: 100 requests per minute
func GetDefaultRateLimiter() *RateLimiter {
	once.Do(func() {
		defaultRateLimiter = NewRateLimiter(1*time.Minute, 100)
	})
	return defaultRateLimiter
}

// AllowRequest checks if a request should be allowed using the default rate limiter
func AllowRequest(clientID string) bool {
	return GetDefaultRateLimiter().Allow(clientID)
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	ClientID    string
	RetryAfter  time.Duration
	Remaining   int
	MaxRequests int
	Window      time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for client %s: %d requests in %v (max: %d)",
		e.ClientID, e.MaxRequests-e.Remaining, e.Window, e.MaxRequests)
}

// CheckRateLimit checks rate limit and returns an error if exceeded
func CheckRateLimit(clientID string) error {
	rl := GetDefaultRateLimiter()
	if !rl.Allow(clientID) {
		remaining := rl.GetRemaining(clientID)
		return &RateLimitError{
			ClientID:    clientID,
			Remaining:   remaining,
			MaxRequests: rl.maxRequests,
			Window:      rl.window,
		}
	}
	return nil
}
