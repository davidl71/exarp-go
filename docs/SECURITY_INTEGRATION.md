# Security Integration

This document describes the security features integrated into exarp-go.

## Rate Limiting

Rate limiting is integrated into the MCP server via middleware in `internal/factory/server.go`.

### Configuration

Enable rate limiting in your config:

```yaml
security:
  rate_limit:
    enabled: true          # Enable rate limiting
    window_duration: 1m    # Time window (1m = 1 minute)
    requests_per_window: 100  # Max requests per window
```

### How It Works

1. **Middleware**: `toolRateLimitMiddleware` wraps every tool call
2. **Check**: Calls `security.CheckRateLimit(clientID)` before tool execution
3. **Config-aware**: Respects `security.rate_limit.enabled` flag (disabled by default)
4. **Default**: When disabled, allows 1,000,000 requests/minute (effectively unlimited)

### Implementation

- **Middleware**: `internal/factory/server.go:86-106`
- **Core**: `internal/security/ratelimit.go`
- **Config**: Reads from `config.GetGlobalConfig().Security.RateLimit`

### Usage

```go
// Check rate limit for a client
err := security.CheckRateLimit("client-123")
if err != nil {
    // Rate limit exceeded
}

// Or use the convenience function
if !security.AllowRequest("client-123") {
    // Rate limit exceeded
}
```

## Access Control

Access control is defined in `internal/security/access.go` but **not yet integrated** into the MCP server.

### Implementation Status

| Feature | Implemented | Integrated |
|---------|-------------|------------|
| Rate Limiting | ✓ | ✓ |
| Access Control | ✓ | ✗ |

### Future: Access Control Integration

To integrate access control:

1. Create `toolAccessControlMiddleware` similar to rate limiter
2. Check `security.CheckTool(toolName)` before execution
3. Add config for tool permissions

See task T-1773310480291859000: "Integrate security access control into MCP server"
