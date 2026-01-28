# Shared Patterns in Native LLM Abstractions

**Date:** 2026-01-28  
**Scope:** FM (Apple Foundation Models), Report Insights (MLX then FM), Ollama — the three LLM integrations and their native abstractions.

---

## The Three Abstractions

| Abstraction | Purpose | Primary backend | Fallback | Bridge used? |
|-------------|---------|-----------------|----------|--------------|
| **FMProvider** (`fm_provider.go`) | Foundation model text generation (task_analysis, context, estimation, task_workflow, task_discovery) | Apple FM (darwin/arm64/cgo) | None (stub returns ErrFMNotSupported) | No |
| **ReportInsightProvider** (`insight_provider.go`) | Long-form AI insights for report/scorecard | MLX (Python bridge) | DefaultFM | Yes (MLX only) |
| **OllamaProvider** (`ollama_provider.go`) | Ollama tool (status, models, generate, pull, hardware, docs, quality, summary) | Native Go (HTTP to Ollama API) | Python bridge | Yes (fallback only) |

---

## Shared Patterns

### 1. Default instance + init

All three use a **package-level default** set in `init()` so callers don’t construct backends:

- **FM:** `var DefaultFM FMProvider` — set in `fm_apple.go` (CGO) or `fm_stub.go` (non-CGO).
- **Insight:** `var defaultReportInsight ReportInsightProvider` — set in `insight_provider.go` to `&compositeReportInsight{}`.
- **Ollama:** `var defaultOllama OllamaProvider` — set in `ollama_provider.go` to `&compositeOllama{}`.

**Accessor:** All three use an accessor for call sites: `DefaultFMProvider()` (returns DefaultFM), `DefaultReportInsight()`, `DefaultOllama()`. Init still sets the package var `DefaultFM`; fm_apple.go and fm_stub.go are the only places that assign to it.

### 2. Composite: try primary then fallback

Two of three use a **composite** that tries one backend, then another on failure:

- **ReportInsight:** Tries MLX (bridge) first; on failure or empty, uses `DefaultFMProvider().Generate(...)`. Same operation (generate text), two backends.
- **Ollama:** Tries native `handleOllamaNative(ctx, params)` first; on error, uses `invokeOllamaViaBridge(ctx, params)`. Same operation (invoke tool), two backends.

**FM** does not use a composite: it’s either Apple or stub depending on build tags, not “try A then B”.

### 3. Bridge behind abstraction

Handlers and report code **never call the bridge directly** for these LLM flows:

- **Ollama:** `handleOllama` calls `DefaultOllama().Invoke(ctx, params)`; bridge is used only inside `invokeOllamaViaBridge` in `ollama_provider.go`.
- **Report insights:** `enhanceReportWithMLX` calls `DefaultReportInsight().Generate(...)`; bridge is used only inside `tryMLXReportInsight` / `executeMLXViaBridge` in `insight_provider.go`.
- **FM:** No bridge; native Apple or stub only.

So pattern: **single bridge call site per integration**, behind the provider interface.

### 4. Text generation signature (FM and ReportInsight)

**TextGenerator** is the shared interface for “generate text from prompt + options”:

```go
type TextGenerator interface {
    Supported() bool
    Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error)
}
```

- **FMProvider** embeds TextGenerator (Apple or stub).
- **ReportInsightProvider** embeds TextGenerator (composite MLX then FM).

The composite report insight uses a **TextGenerator** for the FM fallback, so code that only needs “generate text” can accept either `DefaultFMProvider()` or `DefaultReportInsight()` via this interface.

### 5. Tool invocation vs text generation

- **OllamaProvider** is **tool-shaped:** `Invoke(ctx, params) ([]framework.TextContent, error)`. It forwards a params map and returns MCP-style text content. Multiple actions (status, models, generate, …) live inside native/bridge.
- **FMProvider** and **ReportInsightProvider** are **text-generation-shaped:** `Generate(ctx, prompt, maxTokens, temperature) (string, error)`. Single operation, same signature.

So we have two patterns:

- **Text generation:** `Supported() bool` + `Generate(ctx, prompt, maxTokens, temp) (string, error)` — FM, ReportInsight.
- **Tool invocation:** `Invoke(ctx, params) ([]framework.TextContent, error)` — Ollama.

---

## Summary Table

| Pattern | FM | ReportInsight | Ollama |
|--------|----|----------------|--------|
| Default var + init | ✅ DefaultFM | ✅ defaultReportInsight | ✅ defaultOllama |
| Accessor function | ✅ DefaultFMProvider() | ✅ DefaultReportInsight() | ✅ DefaultOllama() |
| Composite (try A then B) | No | ✅ MLX then FM | ✅ Native then bridge |
| Bridge behind abstraction | N/A (no bridge) | ✅ in insight_provider | ✅ in ollama_provider |
| Interface shape | Generate(...) (string, error) | Generate(...) (string, error) | Invoke(...) ([]TextContent, error) |
| Supported() | ✅ | ✅ | No (composite always “tries”) |

---

## Streamlining Opportunities

### Implemented

- **FMAvailable()** — Call sites use `FMAvailable()` or `!FMAvailable()` instead of repeating nil/Supported checks. Defined in `fm_provider.go`.
- **DefaultFMProvider() accessor** — For consistency with `DefaultReportInsight()` and `DefaultOllama()`. All FM call sites that read the provider use `DefaultFMProvider()`; only fm_apple.go and fm_stub.go set `DefaultFM` in init.
- **TextGenerator interface** — Shared contract in `fm_provider.go`: `Supported() bool` + `Generate(...) (string, error)`. FMProvider and ReportInsightProvider both embed it. The report-insight composite uses a `TextGenerator` for the FM fallback.

### Optional (not implemented)
- **Composite helper** — ReportInsight and Ollama both use "try A then B". Shared machinery would add indirection for little gain; keep as-is.

---

## Files Reference

| File | Role |
|------|------|
| `internal/tools/fm_provider.go` | TextGenerator interface, FMProvider (embeds TextGenerator), ErrFMNotSupported, DefaultFM var, DefaultFMProvider(), FMAvailable() |
| `internal/tools/fm_apple.go` | Apple FM implementation (darwin, arm64, cgo) |
| `internal/tools/fm_stub.go` | Stub implementation (other platforms) |
| `internal/tools/insight_provider.go` | ReportInsightProvider (embeds TextGenerator), composite (MLX then TextGenerator fallback), DefaultReportInsight(), MLX bridge helpers |
| `internal/tools/ollama_provider.go` | OllamaProvider, composite (native then bridge), DefaultOllama(), bridge helper |

Handlers and report code use **only** the default providers and never touch the bridge or platform for these three LLM integrations.
