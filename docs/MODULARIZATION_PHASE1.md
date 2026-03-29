# Modularization Phase 1 (in progress)

**Related:** [MODULARIZATION_PACKAGE_MAP.md](./MODULARIZATION_PACKAGE_MAP.md), [plans/mcp-go-core-extraction.plan.md](./plans/mcp-go-core-extraction.plan.md)

## Done

| Step | Detail |
|------|--------|
| **Cache extraction** | `FileCache`, `TTLCache`, and `GetGlobalFileCache` live in **`mcp-go-core/pkg/mcp/cache`**. exarp-go **`internal/cache`** re-exports them and keeps **`GetScorecardCache()`** (scorecard singleton). |
| **Local `replace`** | `go.mod` includes `replace github.com/davidl71/mcp-go-core => ../mcp-go-core` for sibling-repo dev. Remove after publishing a tagged **mcp-go-core** release and `go get` bump. |

## Ratelimit (already complete)

`internal/security/ratelimit.go` is a thin wrapper over **`mcp-go-core/pkg/mcp/security`** (see extraction plan T-1772056740723802000).

## Next (Phase 1 remainder)

1. **ToolError / response compact** — move types and helpers per extraction plan (medium touch).
2. **FileLock** — optional `pkg/mcp/filelock` in core (build tags).
3. **Publish mcp-go-core** — tag patch release, drop `replace` in exarp-go, document in release notes.

## Verify

```bash
cd ../mcp-go-core && go test ./pkg/mcp/cache/...
cd ../exarp-go && make test-go
```
