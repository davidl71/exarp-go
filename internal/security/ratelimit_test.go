package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 3)

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		if !rl.Allow("client1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("client1") {
		t.Error("4th request should be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should allow requests again
	if !rl.Allow("client1") {
		t.Error("Request after window should be allowed")
	}

	rl.Stop()
}

func TestRateLimiterMultipleClients(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 2)

	// Client 1 should be allowed
	if !rl.Allow("client1") {
		t.Error("Client1 request should be allowed")
	}

	// Client 2 should be allowed (separate limit)
	if !rl.Allow("client2") {
		t.Error("Client2 request should be allowed")
	}

	// Client 1 should still be allowed (different client)
	if !rl.Allow("client1") {
		t.Error("Client1 second request should be allowed")
	}

	// Client 1 should be denied (exceeded limit)
	if rl.Allow("client1") {
		t.Error("Client1 third request should be denied")
	}

	// Client 2 should still be allowed
	if !rl.Allow("client2") {
		t.Error("Client2 second request should be allowed")
	}

	rl.Stop()
}

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1)

	// First request should be allowed
	if !rl.Allow("client1") {
		t.Error("First request should be allowed")
	}

	// Second request should be denied
	if rl.Allow("client1") {
		t.Error("Second request should be denied")
	}

	// Wait should succeed after window expires
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx, "client1")
	if err != nil {
		t.Errorf("Wait should succeed: %v", err)
	}

	rl.Stop()
}

func TestRateLimiterGetRemaining(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 5)

	// Should start with 5 remaining
	if remaining := rl.GetRemaining("client1"); remaining != 5 {
		t.Errorf("Expected 5 remaining, got %d", remaining)
	}

	// Make 2 requests
	rl.Allow("client1")
	rl.Allow("client1")

	// Should have 3 remaining
	if remaining := rl.GetRemaining("client1"); remaining != 3 {
		t.Errorf("Expected 3 remaining, got %d", remaining)
	}

	rl.Stop()
}

func TestCheckRateLimit(t *testing.T) {
	// Create a new limiter for this test
	rl := NewRateLimiter(100*time.Millisecond, 2)

	// First request should succeed
	if !rl.Allow("test-client") {
		t.Error("First request should succeed")
	}

	// Second request should succeed
	if !rl.Allow("test-client") {
		t.Error("Second request should succeed")
	}

	// Third request should fail
	if rl.Allow("test-client") {
		t.Error("Third request should fail")
	}

	// Check remaining
	remaining := rl.GetRemaining("test-client")
	if remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}

	rl.Stop()
}

// TestRequestLimitEnforcement verifies that CheckRateLimit enforces the limit and returns
// RateLimitError when the request limit is exceeded (T-288).
func TestRequestLimitEnforcement(t *testing.T) {
	cfg := config.GetDefaults()
	cfg.Security.RateLimit.Enabled = true
	cfg.Security.RateLimit.RequestsPerWindow = 2
	cfg.Security.RateLimit.WindowDuration = 1 * time.Minute
	cfg.Security.RateLimit.BurstSize = 2
	config.SetGlobalConfig(cfg)
	// Reset default rate limiter so it picks up our config (test is the first to use it)
	// GetDefaultRateLimiter() uses sync.Once; we set config before any CheckRateLimit call
	// so the default limiter is created with our test config.

	clientID := "enforce-test-client"

	// First two requests should be allowed
	for i := 0; i < 2; i++ {
		err := CheckRateLimit(clientID)
		if err != nil {
			t.Errorf("Request %d should be allowed, got error: %v", i+1, err)
		}
	}

	// Third request should be denied (enforcement)
	err := CheckRateLimit(clientID)
	if err == nil {
		t.Error("Third request should be denied (rate limit enforced), got nil error")
	}

	var rateErr *RateLimitError
	if err != nil && !errors.As(err, &rateErr) {
		t.Errorf("Expected *RateLimitError, got %T: %v", err, err)
	}

	if rateErr != nil && rateErr.ClientID != clientID {
		t.Errorf("RateLimitError.ClientID = %s, want %s", rateErr.ClientID, clientID)
	}
}

// TestSlidingWindowAccuracy verifies that the sliding window correctly allows new requests
// as old ones expire (T-289). Window=100ms, max=2: first 2 allowed, 3rd denied; after 100ms
// the first request slides out, so one more should be allowed.
func TestSlidingWindowAccuracy(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 2)

	// First two requests allowed
	if !rl.Allow("client1") || !rl.Allow("client1") {
		t.Error("First two requests should be allowed")
	}
	// Third denied
	if rl.Allow("client1") {
		t.Error("Third request should be denied (window full)")
	}

	// Wait for the first request to slide out of the window (>100ms)
	time.Sleep(110 * time.Millisecond)

	// One slot should be free (oldest request expired); one new request allowed
	if !rl.Allow("client1") {
		t.Error("Request after window slide should be allowed (one slot free)")
	}
	// Second after slide allowed (window now has 2)
	if !rl.Allow("client1") {
		t.Error("Second request after slide should be allowed")
	}
	// Third after slide denied (window full again)
	if rl.Allow("client1") {
		t.Error("Third request after slide should be denied")
	}

	rl.Stop()
}

// TestConcurrentRequests verifies rate limiting under concurrent access (T-291).
func TestConcurrentRequests(t *testing.T) {
	rl := NewRateLimiter(1*time.Second, 10)

	const goroutines = 20

	allowed := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			allowed <- rl.Allow("concurrent-client")
		}()
	}

	allowedCount := 0

	for i := 0; i < goroutines; i++ {
		if <-allowed {
			allowedCount++
		}
	}

	if allowedCount > 10 {
		t.Errorf("Concurrent: expected at most 10 allowed, got %d", allowedCount)
	}

	rl.Stop()
}

// TestLimitResetBehavior verifies that after the window expires, limits reset (T-292).
func TestLimitResetBehavior(t *testing.T) {
	rl := NewRateLimiter(80*time.Millisecond, 2)

	// Exhaust limit
	if !rl.Allow("reset-client") || !rl.Allow("reset-client") {
		t.Error("First two requests should be allowed")
	}

	if rl.Allow("reset-client") {
		t.Error("Third request should be denied")
	}

	// Wait for full window expiry
	time.Sleep(100 * time.Millisecond)

	// Limit should be reset; fresh window
	if !rl.Allow("reset-client") || !rl.Allow("reset-client") {
		t.Error("After reset: first two requests should be allowed")
	}

	if rl.Allow("reset-client") {
		t.Error("After reset: third request should be denied")
	}

	rl.Stop()
}
