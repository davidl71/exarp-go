# Modularization Phase 1 (in progress)

**Related:** [MODULARIZATION_PACKAGE_MAP.md](./MODULARIZATION_PACKAGE_MAP.md), [plans/mcp-go-core-extraction.plan.md](./plans/mcp-go-core-extraction.plan.md)

## Done

| Step | Detail |
|------|--------|
| **Cache extraction** | `FileCache`, `TTLCache`, and `GetGlobalFileCache` live in **`mcp-go-core/pkg/mcp/cache`**. exarp-go **`internal/cache`** re-exports them and keeps **`GetScorecardCache()`** (scorecard singleton). |
| **Response + tool errors** | **`FormatResult`** and **`FormatResultCompact`** delegate to **`mcp-go-core/pkg/mcp/response`** from exarp **`internal/framework/response.go`**; **`ConvertToMap`** stays in exarp (JSON round-trip semantics). **`ErrToolFailed`** / **`IsToolFailed`** in core **`pkg/mcp/framework`**, re-exported from exarp **`internal/framework`**, used in **`internal/tools/handlers_wrap.go`**. |
| **Published core** | **mcp-go-core `v0.4.3`** (tag on GitHub). exarp-go **`require github.com/davidl71/mcp-go-core v0.4.3`** — no `replace` for core in `go.mod`. **`vendor/`** is gitignored; **release order:** tag and push **`v0.4.3`** on **mcp-go-core** first so **`go mod download`**, **`make go-mod-verify`**, and CI can resolve the module. |

## Ratelimit (already complete)

`internal/security/ratelimit.go` is a thin wrapper over **`mcp-go-core/pkg/mcp/security`** (see extraction plan T-1772056740723802000).

## Next (Phase 1 remainder)

1. **FileLock** — optional `pkg/mcp/filelock` in core (build tags).

## Verify

```bash
cd ../mcp-go-core && go test ./pkg/mcp/cache/...
cd ../exarp-go && make test-go
```
