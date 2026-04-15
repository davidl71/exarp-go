# MLX Tool: Options for Native Go

**Status:** The `mlx` tool is **hybrid** — native Go for **status**, **hardware** (luxfi/mlx on darwin+CGO), and **models** (static list, all builds); Python bridge for **generate** only (and for status/hardware when native unavailable). Report/scorecard AI insights use DefaultReportInsight (FM then MLX).

**Goal:** Document options and current state (Option B spike done).

---

## Option A: Keep Bridge-Only (Current)

- **No code change.** Document mlx as intentionally bridge-only.
- **Pros:** No CGO/MLX dependency in Go; Python mlx package is mature and well-documented.
- **Cons:** Last remaining Python tool; requires Python + mlx in environment for mlx features.

**When to choose:** Default. Use unless you need to remove Python from the critical path for mlx.

---

## Option B: Native Go via luxfi/mlx ✅ Spike done

**Go bindings for Apple MLX:** [github.com/luxfi/mlx](https://pkg.go.dev/github.com/luxfi/mlx) — MIT license, mirrors Python API, includes Metal bindings (`github.com/luxfi/mlx/metal`). Actively maintained (e.g. v0.30.2 Dec 2025).

**Implemented (spike):**
1. Added `github.com/luxfi/mlx` to Go modules.
2. `internal/tools/mlx_native.go` (build tag `darwin && cgo`): `handleMlxNative` implements **status** (Info, GetBackend), **hardware** (GetBackend, GetDevice), and **models** (static recommended list from `mlx_models_data.go`); **generate** returns a clear error so the handler falls back to the Python bridge.
3. `internal/tools/mlx_native_nocgo.go` (build tag `!(darwin || cgo)`): stub returns "MLX native not available".
4. `handleMlx` calls native first; on success returns; for actions models/generate on error falls through to bridge; for other actions returns error.

**Build note:** With `-mod=vendor`, the vendored luxfi package may lack the `metal/` subdirectory (mtl_bridge.m), so CGO build can fail. Use `-mod=mod` when building with CGO on darwin, or see luxfi README for full Metal build requirements.

**References:**
- [luxfi/mlx - Go Packages](https://pkg.go.dev/github.com/luxfi/mlx)
- [MLX LLM inference (Python)](https://ml-explore.github.io/mlx/build/html/examples/llama-inference.html) — patterns to adapt for Go
- [mlx-lm](https://github.com/ml-explore/mlx-lm) — Python LLM with MLX; reference for generate flow

---

## Option C: Alternative — GoMLX

**GoMLX:** [github.com/gomlx](https://github.com/gomlx) — accelerated ML framework for Go (740+ stars), go-huggingface for models/tokenizers. Different API than Apple MLX Python/luxfi.

**Approach:** Use GoMLX for inference instead of Apple MLX; would require an adapter layer and possibly different model format/tool contract.

**Effort:** Higher — different API, more integration work. Consider only if luxfi/mlx is unsuitable.

---

## Comparison: B (luxfi/mlx) vs C (GoMLX)

| Aspect | B: luxfi/mlx | C: GoMLX |
|--------|----------------|----------|
| **What it is** | Go bindings for **Apple MLX** (same framework as Python `mlx`). Array framework + Metal/CUDA/CPU backends. | Full **ML framework** for Go (PyTorch/JAX/TensorFlow-like). Training + inference; OpenXLA + pure Go backends. |
| **API vs current mlx tool** | Mirrors Python MLX API (arrays, `Info()`, `SetBackend`, `GetBackend`, etc.). Same mental model as Python mlx. | Different API. No direct “MLX Python equivalent”; use GoMLX’s own model stack (e.g. Gemma, ONNX). |
| **LLM inference** | **Low-level only** in Go: array ops (Add, MatMul, Zeros, Context). No high-level “load model + generate” in the package. You’d implement or adapt a generate loop (e.g. from mlx-lm patterns) using luxfi arrays. | **High-level path exists**: e.g. [gomlx/gemma](https://github.com/gomlx/gemma) for Gemma; go-huggingface for models/tokenizers; ONNX via onnx-gomlx. LLM inference is a supported use case. |
| **Platforms** | Metal (Apple Silicon), CUDA (NVIDIA), CPU. CGO required for GPU. | XLA: Linux (primary), Mac (XLA CPU for now; GPU TBD), Windows experimental. Pure Go backend: portable (WASM, etc.). |
| **Apple Silicon GPU** | **Yes** — Metal backend, same as Python MLX. | Mac XLA backend: **CPU for now**; GPU pending (see jax-ml/jax#32800). |
| **Model format / ecosystem** | Same as Python MLX (safetensors, etc.); compatible with mlx-lm–style models if you wire inference in Go. | Gemma, HuggingFace (via go-huggingface), ONNX. Not MLX-native; different model formats/tooling. |
| **Effort for “mlx tool” parity** | **Medium:** Add luxfi/mlx; implement status/hardware (Info, GetBackend, GetDevice); generate requires building or porting a small inference loop (no ready-made LLM helper in the Go package). | **Higher:** Different stack. Use Gemma or ONNX for “generate”; would not be “MLX” anymore (different backend/models). Good if you want “LLM in Go” in general, not “MLX in Go” specifically. |
| **License** | MIT | Apache-2.0 |
| **Maturity / activity** | v0.30.2 (Dec 2025); official Go bindings (linked from MLX docs). | ~1.3k stars; v0.26.0 (Dec 2025); active; Gemma, ONNX, HuggingFace ecosystem. |

**Summary:**  
- **B (luxfi/mlx):** Best fit if the goal is “native **MLX** in Go” — same framework as the Python tool, Apple Silicon GPU via Metal. You get status/hardware easily; generate needs a custom or ported inference loop (no high-level LLM API in the Go binding).  
- **C (GoMLX):** Best fit if the goal is “LLM inference in Go” with minimal Python. Richer high-level LLM support (Gemma, ONNX, tokenizers) but different API and model ecosystem; on Mac, XLA is CPU-only for now, so no Apple GPU parity with current MLX Python.

---

## Recommendation

- **Current:** **Option B spike done** — mlx is hybrid: native status/hardware (luxfi/mlx) when built with CGO on darwin; models/generate use Python bridge. See NATIVE_GO_HANDLER_STATUS.md.
- **Short term:** **Option A** — if you prefer no CGO for mlx, keep mlx bridge-only (revert native path).
- **If you want “LLM in Go” regardless of MLX:** **Option C** (GoMLX + Gemma/ONNX) is viable but replaces MLX with a different stack; use when Apple GPU is not required or when you want a single Go-centric LLM stack.

---

## References

- **Handler:** `internal/tools/handlers.go` (`handleMlx`)
- **Insight provider:** `internal/tools/insight_provider.go` (`executeMLXViaBridge`)
- **Report MLX:** `internal/tools/report_mlx.go`
- **Status:** `docs/NATIVE_GO_HANDLER_STATUS.md` (Hybrid: mlx native status/hardware, bridge models/generate)
- **Plan:** `docs/NATIVE_GO_MIGRATION_PLAN.md` (§ mlx: Options for Native Go)
