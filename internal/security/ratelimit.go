// Package security provides rate limiting and access checks.
//
// Rate limiter: This package wraps the core sliding-window rate limiter and adds
// config-aware helpers that read from exarp's centralized security config.
// For a single-resource or global token-bucket limit (no per-client map), consider
// golang.org/x/time/rate instead (rate.Limiter). See docs/VENDOR_USAGE_AND_IMPROVEMENTS.md.
package security

import (
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	coresecurity "github.com/davidl71/mcp-go-core/pkg/mcp/security"
)

// Re-export core types so callers don't need to change imports.
type (
	RateLimiter    = coresecurity.RateLimiter
	RateLimitError = coresecurity.RateLimitError
)

// NewRateLimiter creates a new rate limiter (re-exported from core).
var NewRateLimiter = coresecurity.NewRateLimiter

var (
	defaultRateLimiter *RateLimiter
	once               sync.Once
)

// GetDefaultRateLimiter returns a rate limiter configured from exarp's security config.
// Uses centralized config if available, otherwise defaults to 100 requests per minute.
func GetDefaultRateLimiter() *RateLimiter {
	once.Do(func() {
		cfg := config.GetGlobalConfig()
		if cfg.Security.RateLimit.Enabled {
			defaultRateLimiter = NewRateLimiter(
				cfg.Security.RateLimit.WindowDuration,
				cfg.Security.RateLimit.RequestsPerWindow,
			)
		} else {
			// Rate limiting disabled — create a permissive limiter
			defaultRateLimiter = NewRateLimiter(1*time.Minute, 1_000_000)
		}
	})
	return defaultRateLimiter
}

// AllowRequest checks if a request should be allowed using the default rate limiter.
func AllowRequest(clientID string) bool {
	return GetDefaultRateLimiter().Allow(clientID)
}

// CheckRateLimit checks rate limit and returns an error if exceeded.
// Respects config enable flag — returns nil immediately when rate limiting is disabled.
func CheckRateLimit(clientID string) error {
	cfg := config.GetGlobalConfig()
	if !cfg.Security.RateLimit.Enabled {
		return nil
	}
	rl := GetDefaultRateLimiter()
	if !rl.Allow(clientID) {
		remaining := rl.GetRemaining(clientID)
		return &RateLimitError{
			ClientID:    clientID,
			Remaining:   remaining,
			MaxRequests: cfg.Security.RateLimit.RequestsPerWindow,
			Window:      cfg.Security.RateLimit.WindowDuration,
		}
	}
	return nil
}
