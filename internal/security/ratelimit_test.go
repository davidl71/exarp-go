package security

import (
	"context"
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
