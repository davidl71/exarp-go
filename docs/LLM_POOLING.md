# LLM Usage Pooling

**Summary:** LLM-related calls avoid "instance per call" via **(1) shared Ollama HTTP client** (native Go). Python bridge pool was removed 2026-01-29; bridge uses one-shot `execute_tool.py` per call.

---

## 1. ~~Python bridge pool~~ (Removed 2026-01-29)

Bridge-based tool calls (e.g. **MLX** generate when native unavailable) now use **one-shot subprocess** (`bridge/execute_tool.py`) per call. The persistent Python daemon (`execute_tool_daemon.py`) and Go pool (`internal/bridge/pool.go`) were removed. See `internal/bridge/python.go` — `ExecutePythonTool` calls `executePythonToolSubprocess` only.

---

## 2. Shared Ollama HTTP client (native Go)

Native Ollama calls (HTTP to `ollama serve`) use a **shared `http.Client`** (`ollamaHTTPClient` in `internal/tools/ollama_native.go`) instead of creating a new client per request.

- **Benefits:** Connection reuse (Keep-Alive), fewer TCP handshakes.
- **Timeouts:** Per-request via `context.WithTimeout` (e.g. 2s availability check, 5–10s status/models, config-based for generate/pull).

---

## 3. What is not pooled

- **Apple FM:** In-process; no subprocess or HTTP.
- **Ollama server:** We never start it; we call an existing `ollama serve` over HTTP.

---

## Quick reference

| Path | Pooling | Notes |
|------|---------|--------|
| MLX, ollama (bridge), estimation, etc. | Python process pool | Default on; disable with `EXARP_PYTHON_POOL_ENABLED=false` |
| Ollama native (HTTP) | Shared HTTP client | Always used |
| Apple FM | N/A | In-process |
