# Ollama Abstraction Opportunities

**Date:** 2026-01-28  
**Context:** Ollama is hybrid: native Go first (HTTP client to Ollama API), Python bridge fallback when native fails. This doc outlines abstraction opportunities so handlers don't depend on the bridge by name and the flow is consistent with FM/MLX patterns.

---

## Where Ollama Touches the Python Bridge

| Location | Purpose | Bridge use |
|----------|---------|------------|
| **handlers.go** `handleOllama` | Direct `ollama` tool (status, models, generate, pull, hardware, docs, quality, summary) | Fallback only: `bridge.ExecutePythonTool(ctx, "ollama", params)` when **native returns an error** |

**Native coverage:** All actions are implemented in `internal/tools/ollama_native.go` (status, models, generate, pull, hardware, docs, quality, summary). The bridge is used only when native fails (e.g. Ollama server unreachable, or action-specific error). There are no Python-only actions.

---

## Abstraction Opportunity: OllamaProvider (optional, low impact)

**Idea:** Hide "try native then bridge" behind an interface so `handleOllama` does not call `bridge.ExecutePythonTool` directly.

**Interface (e.g. in `internal/tools/ollama_provider.go`):**

```go
// OllamaProvider invokes the ollama tool (native first, then bridge fallback).
type OllamaProvider interface {
    Invoke(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error)
}
```

**Implementations:**

1. **NativeOllama** – Calls existing `handleOllamaNative(ctx, params)`; returns its result or error.
2. **BridgeOllama** – Calls `bridge.ExecutePythonTool(ctx, "ollama", params)`, parses response into `[]framework.TextContent`.
3. **CompositeOllama** – Tries NativeOllama first; on error, calls BridgeOllama. This is the current handler logic in one place.

**Default:** `DefaultOllama = &CompositeOllama{}`. `handleOllama` becomes: parse args → `DefaultOllama.Invoke(ctx, params)` → return. No `bridge` import in handlers for ollama.

**Benefits:**

- Handlers don't reference the bridge for ollama; consistent with MLX/insight provider pattern.
- Easier to test (mock OllamaProvider).
- Single place for "native then bridge" logic.

**Drawbacks:** Thin wrapper; behavior unchanged. Main gain is consistency and testability.

---

## Comparison with MLX / Report Insights

| Aspect | MLX | Report insights (fixed) | Ollama |
|--------|-----|--------------------------|--------|
| **Bridge use** | Tool + report enhancement | Enhancement only (now via provider) | Fallback only when native fails |
| **Abstraction done** | Report insights → ReportInsightProvider (MLX then FM) | ✅ | Optional OllamaProvider |
| **Second use** | Report called bridge for "mlx" generate | No longer; report uses DefaultReportInsight | No second use; only handleOllama |
| **Recommendation** | MLX tool stays bridge-only | Done | Optional: add OllamaProvider for consistency |

---

## Recommendation

- **Optional follow-up:** Add `OllamaProvider` and `DefaultOllama` (composite: native then bridge); have `handleOllama` call `DefaultOllama.Invoke(ctx, params)` and remove the direct bridge call from handlers. No new functionality; consistency with FM/MLX/insight patterns and cleaner tests.
- **No "report-style" abstraction:** Unlike MLX, ollama is not used by another tool for a separate feature, so there is no second call site to abstract.

---

## Summary

| Area | Current | After optional abstraction |
|------|--------|----------------------------|
| **handleOllama** | Tries native, then `bridge.ExecutePythonTool(ctx, "ollama", params)` | Tries native, then bridge via `DefaultOllama.Invoke(ctx, params)`; no bridge import in handler |

Ollama already uses the bridge only as a fallback; the only abstraction opportunity is a thin provider for consistency and testability, not for removing or replacing bridge use.
