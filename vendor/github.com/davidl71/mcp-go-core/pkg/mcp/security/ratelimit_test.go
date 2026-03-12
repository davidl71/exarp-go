package security

import (
	"context"
	"sync"
	"testing"
	"time"
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

// TestRateLimiterConcurrentRequests tests T-291: concurrent requests from multiple goroutines.
// Verifies rate limiter is safe under concurrent access and enforces limits correctly.
func TestRateLimiterConcurrentRequests(t *testing.T) {
	rl := NewRateLimiter(1*time.Second, 10)
	defer rl.Stop()

	var allowed, denied int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok := rl.Allow("concurrent-client")
			mu.Lock()
			if ok {
				allowed++
			} else {
				denied++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Should allow exactly 10, deny 20
	if allowed != 10 {
		t.Errorf("expected 10 allowed, got %d", allowed)
	}
	if denied != 20 {
		t.Errorf("expected 20 denied, got %d", denied)
	}
}

// TestRateLimiterSlidingWindowAccuracy tests T-289: sliding window accuracy.
// Verifies that requests are correctly expired as the window slides.
func TestRateLimiterSlidingWindowAccuracy(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 2)
	defer rl.Stop()

	// Use 2 requests
	if !rl.Allow("slide") {
		t.Error("request 1 should be allowed")
	}
	if !rl.Allow("slide") {
		t.Error("request 2 should be allowed")
	}
	if rl.Allow("slide") {
		t.Error("request 3 should be denied")
	}

	// Wait 60ms - oldest request should still be in window
	time.Sleep(60 * time.Millisecond)
	if rl.Allow("slide") {
		t.Error("request before window slides should be denied")
	}

	// Wait another 50ms - oldest request expires (100ms total from first)
	time.Sleep(50 * time.Millisecond)
	if !rl.Allow("slide") {
		t.Error("request after window slides should be allowed")
	}
}

// TestRateLimiterResetBehavior tests T-292: limit reset after window expires.
func TestRateLimiterResetBehavior(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 2)
	defer rl.Stop()

	// Exhaust limit
	rl.Allow("reset")
	rl.Allow("reset")
	if rl.Allow("reset") {
		t.Error("third request should be denied")
	}

	// Wait for full window expiry
	time.Sleep(60 * time.Millisecond)

	// Limit should reset - 2 more allowed
	if !rl.Allow("reset") {
		t.Error("first request after reset should be allowed")
	}
	if !rl.Allow("reset") {
		t.Error("second request after reset should be allowed")
	}
	if rl.Allow("reset") {
		t.Error("third request after reset should be denied")
	}
}
