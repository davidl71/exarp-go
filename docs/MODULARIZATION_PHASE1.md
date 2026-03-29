# Modularization Phase 1 (in progress)

**Related:** [MODULARIZATION_PACKAGE_MAP.md](./MODULARIZATION_PACKAGE_MAP.md), [plans/mcp-go-core-extraction.plan.md](./plans/mcp-go-core-extraction.plan.md)

## Done

| Step | Detail |
|------|--------|
| **Cache extraction** | `FileCache`, `TTLCache`, and `GetGlobalFileCache` live in **`mcp-go-core/pkg/mcp/cache`**. exarp-go **`internal/cache`** re-exports them and keeps **`GetScorecardCache()`** (scorecard singleton). |
| **Published core** | **mcp-go-core `v0.4.2`** (tag on GitHub). exarp-go **`require github.com/davidl71/mcp-go-core v0.4.2`** — no `replace` for core in `go.mod`. For local core hacking, add a temporary `replace` line. |

## Ratelimit (already complete)

`internal/security/ratelimit.go` is a thin wrapper over **`mcp-go-core/pkg/mcp/security`** (see extraction plan T-1772056740723802000).

## Next (Phase 1 remainder)

1. **ToolError / response compact** — move types and helpers per extraction plan (medium touch).
2. **FileLock** — optional `pkg/mcp/filelock` in core (build tags).

## Verify

```bash
cd ../mcp-go-core && go test ./pkg/mcp/cache/...
cd ../exarp-go && make test-go
```
