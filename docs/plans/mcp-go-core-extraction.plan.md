---
title: Extract reusable MCP infrastructure to mcp-go-core
epic: T-1772056718478802000
status: Todo
updated: 2026-02-25
---

# Extract reusable MCP infrastructure to mcp-go-core

Move generic, domain-agnostic MCP infrastructure from exarp-go into mcp-go-core
so it can be shared across any MCP server.

## Background

Architectural analysis identified four extraction targets and one duplicate cleanup.
The principle: if it has zero exarp-specific dependencies and any MCP server would
want it, it belongs in mcp-go-core.

## Tasks

- [ ] T-1772056740723802000 — Remove duplicate ratelimit.go, use mcp-go-core version
- [ ] T-1772056739613403000 — Extract TTLCache and FileCache to mcp-go-core/pkg/mcp/cache
- [ ] T-1772056740718615000 — Extract ToolError types to mcp-go-core/pkg/mcp/framework
- [ ] T-1772056740722069000 — Add FormatResultCompact to mcp-go-core/pkg/mcp/response
- [ ] T-1772056740720274000 — Extract FileLock to mcp-go-core/pkg/mcp/filelock

## Execution order

**Phase 1 — Low risk, high value (no new packages in core)**

1. `T-1772056740723802000` **Remove duplicate ratelimit.go**
   - Pure deletion in exarp-go; mcp-go-core already has the type
   - Wire `Security.RateLimit.*` config values into the core constructor
   - Risk: low — unit tests cover rate limiting

2. `T-1772056739613403000` **TTLCache / FileCache → mcp-go-core/pkg/mcp/cache**
   - New package in mcp-go-core; zero deps to carry over
   - Update exarp-go import paths; keep `GetScorecardCache()` local
   - Risk: low — types are self-contained

**Phase 2 — Medium complexity (touches tool error handling)**

3. `T-1772056740718615000` **ToolError types → mcp-go-core/pkg/mcp/framework**
   - Add to existing framework package; no new package needed
   - Many call sites in exarp-go tools — update imports project-wide
   - Risk: medium — broad but mechanical import-path change

4. `T-1772056740722069000` **FormatResultCompact → mcp-go-core/pkg/mcp/response**
   - Add alongside existing `FormatResult`; keep token-estimate logic local
   - Risk: low — additive change

**Phase 3 — Lower priority (OS-level, fcntl)**

5. `T-1772056740720274000` **FileLock → mcp-go-core/pkg/mcp/filelock**
   - New package; requires unix build tags in mcp-go-core
   - Keep `TaskLock`, `StateFileLock`, `WithGitLock` local
   - Risk: medium — build tag handling across module boundary

## Workflow per task

1. Make change in mcp-go-core, bump patch version
2. Run `make test` in mcp-go-core
3. Update `go.mod` replace directive version in exarp-go (already uses local replace)
4. Update exarp-go import paths
5. `make b && make test-go` in exarp-go
6. Commit both repos

## What stays in exarp-go

| Package | Reason |
|---|---|
| `internal/tools/task_*` | Todo2 domain |
| `internal/database/` | SQLite/Todo2 storage |
| `internal/models/` | Todo2 protobuf models |
| `internal/config/` | Exarp-specific config schema |
| `internal/cli/` | TUI + cursor integration |
| `internal/tools/llm_backends.go` | FM/Ollama/MLX orchestration |
| `GetScorecardCache()` | Uses exarp config |
| `TaskLock`, `WithGitLock` | .todo2 path specifics |
| `AddTokenEstimateToResult` | Uses `config.TokensPerChar()` |
