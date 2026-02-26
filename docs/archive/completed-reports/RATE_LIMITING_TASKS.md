# Rate Limiting Tasks (Archived)

**Date archived:** 2026-02-11  
**Reason:** Per-tool limits deferred until HTTP/SSE transport; base rate limiting complete.

---

## STDIO Irrelevance

**exarp-go uses STDIO transport exclusively** (Cursor IDE requirement). Per-tool and per-client limits are **not relevant for STDIO**:

- **STDIO = single client per process.** Cursor spawns one exarp-go process; one stdio connection. No multiple clients to throttle.
- **Per-tool limits** (T-276, T-290): Only matter when multiple HTTP/SSE clients can hammer different tools. With one client, global limits suffice.
- **Removed from Todo2:** T-276 (Add per-tool limits), T-290 (Test per-tool limits) — deleted 2026-02-11 as stdio-irrelevant.

**When to revisit:** When HTTP/SSE transport is implemented (`docs/FUTURE_SSE_IMPLEMENTATION.md`).

---

## Overview

Rate limiting tasks originated from `docs/SECURITY_AUDIT.md` Phase 2. Base implementation is complete (sliding window, per-client). See:

- **Implementation:** `internal/security/ratelimit.go`, `internal/security/ratelimit_test.go`
- **Config:** `internal/config/schema.go` (`RateLimitConfig`), `docs/CONFIGURATION_REFERENCE.md`
- **Status:** `docs/SECURITY_IMPLEMENTATION_STATUS.md` (Phase 2: Rate Limiting ✅ COMPLETE)

---

## Archived / Removed Tasks

| ID | Name | Status | Reason |
|----|------|--------|--------|
| T-276 | Add per-tool limits | Archived | STDIO-irrelevant; deferred until HTTP/SSE (not in Todo2) |
| T-290 | Test per-tool limits | Deleted 2026-02-11 | STDIO-irrelevant; tests deferred per-tool feature |

### T-276: Add per-tool limits

**Scope:** Configure different rate limits per tool (e.g., `ollama` 10 req/min, `mlx` 5 req/min).

**Current state:** Global per-client limits only. No per-tool config.

**Deferral (from `docs/FUTURE_SSE_IMPLEMENTATION.md`):**

> Per-tool rate/limits matter when multiple clients connect over HTTP/SSE; STDIO has a single client per process.

**When to revisit:** When HTTP/SSE transport is implemented and multiple clients can connect. Implementation would extend `RateLimitConfig` with `tool_limits map[string]int` and pass tool name to `CheckRateLimit(clientID, toolName)`.

---

## Related Tasks (Not Archived)

- T-273: Implement sliding window rate limiter — ✅ Done  
- ~~T-274~~ *(removed)* Add middleware to framework — **Future improvement**: documented in `docs/SECURITY_AUDIT.md`. Defer until HTTP/SSE or multi-transport; mcp-go-core has middleware support.  
- T-275: Configure default limits — ✅ Done via config  
- T-277: Add tests for rate limiting — ✅ Done  
- T-288, T-289, T-291, T-292: Base rate limit tests (T-290 per-tool deleted)  

---

## References

- **Source:** `docs/SECURITY_AUDIT.md` Phase 2
- **Backlog:** `docs/BACKLOG_EXECUTION_PLAN.md` (T-273 through T-292)
- **SSE deferral:** `docs/FUTURE_SSE_IMPLEMENTATION.md` — T-276 deferred until multi-client transport
