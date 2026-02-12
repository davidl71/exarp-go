# Future Authentication Implementation

**Status:** Deferred for Future Implementation  
**Created:** 2026-02-11  
**Reason:** STDIO transport (Cursor IDE) has no client identity; authentication applies when HTTP/SSE transport is used

## Overview

Authentication tasks (OAuth configuration, token handling) are **not required for STDIO MCP** (exarp-go's current deployment). They matter when the server is exposed over HTTP/SSE and clients need to authenticate (e.g. web-based MCP clients, multi-tenant cloud deployments).

These tasks are documented here as **future implementation** and have been removed from the Todo2 backlog.

## Deferred Tasks

### T-144: Configure OAuth if needed

**Priority:** Medium  
**Status:** Deferred  
**Former Status:** Todo

**Description:**
Configure OAuth2 flow for MCP server when HTTP/SSE transport is used and client authentication is required.

**Why Deferred:**
- STDIO transport (Cursor IDE) has a single client per process; no client identity or auth flow
- OAuth applies when the server is exposed over HTTP/SSE and multiple clients connect
- See `docs/FUTURE_SSE_IMPLEMENTATION.md` for SSE/HTTP prerequisite

**Required Work (when HTTP/SSE is implemented):**
1. OAuth2 provider configuration (client ID, client secret, redirect URIs)
2. Authorization code flow vs client credentials flow
3. Token exchange and validation
4. Integration with MCP SDK auth interfaces (e.g. `github.com/modelcontextprotocol/go-sdk/auth`, `oauthex`)

**References:**
- MCP SDK: `vendor/modules.txt` — `github.com/modelcontextprotocol/go-sdk/auth`, `github.com/modelcontextprotocol/go-sdk/oauthex`
- `golang.org/x/oauth2` package for OAuth2 flows
- `docs/FUTURE_SSE_IMPLEMENTATION.md` Phase 3: Production Features (Authentication/authorization)

---

### T-146: Handle authentication tokens

**Priority:** Medium  
**Status:** Deferred  
**Former Status:** Todo

**Description:**
Handle authentication tokens (Bearer tokens, session tokens, OAuth access tokens) for HTTP/SSE MCP connections.

**Why Deferred:**
- STDIO transport has no token-based auth; Cursor spawns the process directly
- Token handling applies when clients connect over HTTP/SSE and send `Authorization` headers
- See `docs/FUTURE_SSE_IMPLEMENTATION.md` for SSE/HTTP prerequisite

**Required Work (when HTTP/SSE is implemented):**
1. Parse `Authorization: Bearer <token>` (and similar) from HTTP requests
2. Validate token (JWT validation, OAuth introspection, session lookup)
3. Attach identity/claims to request context for downstream handlers
4. Integrate with access control (`docs/ACCESS_CONTROL_MODEL.md`) when client identity is available

**References:**
- `docs/FUTURE_SSE_IMPLEMENTATION.md` — Phase 3: Authentication/authorization
- `docs/ACCESS_CONTROL_TASKS.md` — T-278 (access control model) matters when transport has client identity

---

## Relationship to Other Deferred Work

| Document | Scope | Relationship |
|----------|--------|---------------|
| `FUTURE_SSE_IMPLEMENTATION.md` | HTTP server, SSE transport, multi-client | Authentication (T-144, T-146) is a Phase 3 production feature for SSE/HTTP |
| `ACCESS_CONTROL_MODEL.md` | Permission system, tool/resource access | Authorization uses client identity; auth (tokens, OAuth) provides that identity when HTTP/SSE is used |
| `RATE_LIMITING_TASKS.md` | Per-tool limits, request limits | Rate limiting can be per-client when auth is available |

## Trigger for Implementation

**Implement when:**
- HTTP/SSE transport is prioritized (see `FUTURE_SSE_IMPLEMENTATION.md`)
- Web-based or multi-tenant MCP clients need to connect
- Access control (T-278, T-281) requires client identity for enforcement

**Current workaround:**
- STDIO transport; no authentication required
- Cursor IDE and local CLI tools do not use tokens

## Decision Log

**2026-02-11:** Documented T-144 and T-146 as future implementation; removed from Todo2 backlog
- **Reason:** Focus on STDIO transport; auth not needed for current use cases
- **Impact:** No change to current behavior
- **Reference:** `docs/FUTURE_SSE_IMPLEMENTATION.md` (optional / lower priority note)
