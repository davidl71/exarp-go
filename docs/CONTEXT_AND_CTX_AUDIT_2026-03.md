# Context tool and `context.Context` audit (2026-03)

Canonical notes from the 2026-03 review of the MCP **`context`** tool and request-scoped cancellation.

## Request context (`ctx`)

- **Source:** MCP tool invocations receive `context.Context` from the Go SDK. `internal/factory` wraps it with `ctxcache.NewContextWithCache(ctx)` for per-request memoization (`toolContextCacheMiddleware`).
- **Handlers:** `handleContext` in `handlers_ai.go` passes `ctx` into summarize, budget, batch, and count paths.

## Changes applied (code)

| Location | Behavior |
|----------|----------|
| `context.go` — `handleContextBudget` | `ctx.Err()` at entry; periodic checks while scanning items; checks during token sort loop. |
| `context.go` — `handleContextBatchNative` | `ctx.Err()` before each item’s summarize/FM work. |
| `context.go` — `handleContextCount` | Accepts `ctx`; fails fast if already canceled. |
| `context_shared.go` — `HandleContextSummarizeShared` | `ctx.Err()` at entry; error copy uses neutral “FM provider / Ollama” wording instead of Apple-only. |
| `handlers_ai.go` — `handleContext` | Removed redundant `FMAvailable()` gate before summarize (errors come from `Generate`); count passes `ctx`. |

Tests: `TestHandleContextCountCancelled` in `context_test.go`.

## FM chain vs messaging

- **`DefaultFM`** is set in `fm_chain.go` as a **chain** (currently **Ollama → stub** in the default `init`; Apple may be added on some builds).
- **`FMAvailable()`** is truthful for the stock chain: `chainFMProvider.Supported()` is true only when **`OllamaReachableForFM()`** succeeds (cached **GET /api/tags**, 2s timeout per probe, **15s TTL**). The chain **stub** is ignored for availability.
- **`ollamaTextGenerator.Supported()`** stays **always true** so `Generate` still attempts HTTP even if a probe was stale (Ollama came up after a failed probe).
- **Discovery:** `LLMBackendStatus()` / `stdio://models` expose **`fm_available`** (same as `FMAvailable()`) and **`ollama_reachable`** (explicit Ollama probe; matches `FMAvailable` when the default chain is Ollama-only).

Implementation: `ollamaPingTagsAPI` in `ollama_native.go`, cache in `fm_ollama.go`, `chainFMProvider.Supported()` in `fm_chain.go`.

## Documentation drift

- **`docs/CONTEXT_TOOLS_COMPARISON.md`** is a **short** comparison + workflow; historical copy is under **`docs/archive/context-tools/`**. This file remains the audit / source of truth for `ctx` and FM behavior.

## Suggestions (exarp tasks)

| Task ID | Topic |
|---------|--------|
| T-1774976982390803000 | Done: trimmed `CONTEXT_TOOLS_COMPARISON.md`; legacy → `archive/context-tools/CONTEXT_TOOLS_COMPARISON_legacy_2026-03-31.md`. |
| T-1774976992707995000 | Done: `FMAvailable` / chain `Supported()` use Ollama `/api/tags` probe + cache; `ollama_reachable` in `LLMBackendStatus`. |
| T-1774976992710110000 | Optional: `sort.Slice` in `handleContextBudget`. *(Created as **medium** in DB; lower to **low** manually if `task update --new-priority` rejects.)* |

**CLI note:** `task create --tags` with comma-separated values may hit a proto/JSON parse bug (unexpected `#tag` token); create without `--tags` or fix CLI separately.

## Related files

- `internal/tools/context.go`, `context_shared.go`, `handlers_ai.go`
- `internal/tools/fm_chain.go`, `fm_provider.go`, `fm_ollama.go`, `ollama_native.go` (`ollamaPingTagsAPI`)
- `internal/factory/server.go` (context cache middleware)
