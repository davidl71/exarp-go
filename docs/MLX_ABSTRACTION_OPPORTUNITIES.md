# MLX Abstraction Opportunities

**Date:** 2026-01-28  
**Context:** MLX has no Go bindings; all MLX usage goes through the Python bridge. This doc outlines abstraction opportunities so that Go code does not depend on the bridge directly where optional behavior (e.g. report insights) can be satisfied by another backend (e.g. Apple FM).

---

## Where MLX Necessitates the Python Bridge

| Location | Purpose | Bridge call |
|----------|---------|-------------|
| **handlers.go** `handleMlx` | Direct `mlx` tool (status, hardware, models, generate) | `bridge.ExecutePythonTool(ctx, "mlx", params)` |
| **report_mlx.go** `enhanceReportWithMLX` | Report/scorecard AI insights (build prompt → generate text → attach as `ai_insights`) | `bridge.ExecutePythonTool(ctx, "mlx", mlxParams)` with `action=generate` |

The `mlx` tool itself (status, hardware, models, generate) remains bridge-only by design (see `docs/MLX_TOOL_RETENTION_RATIONALE.md`). The **report enhancement** path is the one that benefits from abstraction: same “generate from prompt” contract can be fulfilled by MLX (bridge) or by the existing FM provider (Apple FM when available).

---

## Abstraction 1: Report Insights Provider (recommended)

**Idea:** Treat “generate long-form insights for report/scorecard” as a single capability behind an interface. Report code calls the provider; it does not call the bridge or MLX by name.

**Interface (e.g. in `internal/tools/insight_provider.go`):**

```go
// ReportInsightProvider generates AI insights for report/scorecard content.
// Implementations: MLX via bridge, or DefaultFM when MLX unavailable.
type ReportInsightProvider interface {
    Supported() bool
    Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error)
}
```

**Implementations:**

1. **MLXBridge** – Calls `bridge.ExecutePythonTool(ctx, "mlx", params)` with `action=generate`, prompt, model, max_tokens, temperature. Parses bridge response and returns generated text. `Supported()` can be true when bridge is available (or we keep “always try, return error if bridge fails”).
2. **FMReportInsight** – Wraps `DefaultFM`: `Supported() = DefaultFM != nil && DefaultFM.Supported()`, `Generate = DefaultFM.Generate`. Use when MLX is not desired or bridge fails.

**Strategy:** Single default provider that **tries MLX first, then FM** (or config-driven: prefer_mlx vs prefer_fm). So:
- `DefaultReportInsight` = composite: try MLX bridge, on failure or unavailability call `DefaultFM.Generate`.
- `enhanceReportWithMLX` becomes `enhanceReportWithInsights`: build prompt (unchanged), then call `DefaultReportInsight.Generate(ctx, prompt, maxTokens, temp)`; no direct `bridge.ExecutePythonTool` in report_mlx.go.

**Benefits:**
- Report/scorecard code no longer depends on the bridge or “mlx” by name.
- On machines without Python/MLX, report insights can still be provided by Apple FM (when available).
- Same pattern as FMProvider: one capability, multiple backends, clear fallback.

---

## Abstraction 2: MLX Tool Adapter (optional, lower impact)

**Idea:** The `mlx` tool (status, hardware, models, generate) could sit behind an interface so that `handleMlx` does not call `bridge.ExecutePythonTool` directly.

**Interface (e.g. `MLXProvider`):**

```go
type MLXProvider interface {
    Invoke(ctx context.Context, params map[string]interface{}) (string, error)
}
```

**Implementation:** Single implementation that forwards to `bridge.ExecutePythonTool(ctx, "mlx", params)`. No other backend (no Go MLX bindings).

**Benefits:**
- Handlers don’t reference the bridge for the mlx tool; they call `DefaultMLX.Invoke(params)`.
- Easier to test (mock provider) and to swap later if a native MLX backend ever appears.

**Drawbacks:** Thin wrapper; the tool remains bridge-only. Main gain is consistency and testability, not new functionality.

---

## Recommendation

1. **Implement Abstraction 1 (Report Insights Provider)**  
   - Add `ReportInsightProvider` and a composite default that tries MLX (bridge) then DefaultFM.  
   - Change `enhanceReportWithMLX` → `enhanceReportWithInsights` and use the provider so that report code never calls the bridge directly for insights.  
   - Keeps MLX-only tool as-is; only the “report enhancement” path is abstracted and gains an FM fallback.

2. **Abstraction 2 (MLX tool adapter)**  
   - Optional follow-up: introduce `MLXProvider` and use it in `handleMlx` so that the only bridge call for the `mlx` tool lives inside the adapter.

---

## Files to Touch for Abstraction 1

- **New:** `internal/tools/insight_provider.go` – interface, composite “try MLX then FM” implementation, `var DefaultReportInsight ReportInsightProvider`.
- **Update:** `internal/tools/report_mlx.go` – replace direct `bridge.ExecutePythonTool(ctx, "mlx", ...)` with `DefaultReportInsight.Generate(...)`; optionally rename to `enhanceReportWithInsights` and keep existing prompt-building helpers.
- **Update:** `internal/tools/scorecard_mlx.go` / handlers – only if we rename “MLX” in user-facing strings to “AI insights” or “insights”; otherwise just wire to the new provider.
- **Docs:** `docs/MLX_TOOL_RETENTION_RATIONALE.md` – add a short note that report insights are now behind ReportInsightProvider (MLX bridge + FM fallback).

---

## Summary

| Area | Current | After abstraction |
|------|---------|-------------------|
| **Report/scorecard insights** | `report_mlx.go` calls `bridge.ExecutePythonTool(ctx, "mlx", ...)` | Report code calls `DefaultReportInsight.Generate(...)`; provider tries MLX then FM. |
| **mlx tool (status/hardware/models/generate)** | `handleMlx` calls `bridge.ExecutePythonTool(ctx, "mlx", params)` | Unchanged, or optionally behind `MLXProvider` in handleMlx. |

So: **anywhere MLX necessitates the Python bridge today, the main abstraction opportunity is report insights** — same “generate text from prompt” contract, with an FM fallback and no direct bridge dependency in report code. The standalone `mlx` tool can stay bridge-only, with an optional thin adapter for consistency.
