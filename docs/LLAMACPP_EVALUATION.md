# llama.cpp Go bindings evaluation

**Tag hints:** `#research` `#llamacpp` `#local-ai`

Evaluation for exarp-go: choose Go bindings for llama.cpp to support on-device GGUF inference (and optional reuse of Ollama blob storage). Task: T-1771367215872916000.

---

## 1. go-skynet/go-llama.cpp (recommended)

| Criterion        | Assessment |
|-----------------|------------|
| **Maturity**    | High — 871 stars, active issues/PRs, used by go-skynet ecosystem (e.g. llama-cli). |
| **Format**      | GGUF only (post-merge [#180](https://github.com/go-skynet/go-llama.cpp/pull/180)); GGML no longer supported. |
| **Performance** | Work kept in C/C++ via CGO; high-level Go API. Supports Metal (Apple Silicon), CuBLAS, OpenBLAS, ROCm, OpenCL. |
| **Build**       | CGO required. Clone with submodules, build `libbinding.a` (e.g. `BUILD_TYPE=metal make libbinding.a` on macOS/arm64). |
| **exarp-go fit**| Implements a `TextGenerator` that loads a GGUF path and calls the binding; can plug into FM chain (e.g. Apple → LlamaCpp → Ollama → stub) or as `provider=llamacpp`. |

**Links:** [GitHub](https://github.com/go-skynet/go-llama.cpp), [pkg.go.dev](https://pkg.go.dev/github.com/go-skynet/go-llama.cpp).

**Recommendation:** Primary choice for exarp-go. Use CGO build tag (e.g. `llamacpp`) and Metal on darwin/arm64; document build deps (submodule, `libbinding.a`, LIBRARY_PATH/C_INCLUDE_PATH).

---

## 2. swdunlop/llm-go (embedded / lightweight)

| Criterion        | Assessment |
|-----------------|------------|
| **Maturity**    | Lower — ~4 stars; derived from early Ollama, “easier to maintain” focus. |
| **API**         | GGUF load, token-by-token `Next()`, Encode/Predict; HTTP server optional. |
| **Build**       | Wraps llama.cpp (CGO); less documented than go-skynet for multi-platform builds. |
| **exarp-go fit**| Could implement `TextGenerator` by iterating `Next()` until max tokens or stop. Lighter surface, less community traction. |

**Links:** [GitHub](https://github.com/swdunlop/llm-go), [pkg.go.dev](https://pkg.go.dev/github.com/swdunlop/llm-go).

**Recommendation:** Viable for experimentation or minimal embedding; prefer go-skynet for production and alignment with existing ecosystem.

---

## 3. dianlight/gollama.cpp (purego)

Could not find an active “dianlight/gollama.cpp” or “gollama.cpp purego” project in 2024–2025. Searches returned no matching repo or docs. Treat as unavailable unless a current URL is found.

---

## 4. exarp-go integration summary

- **Contract:** `TextGenerator`: `Supported() bool`, `Generate(ctx, prompt, maxTokens, temperature) (string, error)` (see `internal/tools/fm_provider.go`).
- **Placement:** New provider (e.g. `llamacpp_provider.go`) implementing `TextGenerator`; register in `text_generate` as `provider=llamacpp` and optionally in FM chain (e.g. after Apple FM, before Ollama) for “no server” GGUF use.
- **Model path:** Initially config or env (path to `.gguf`); follow-up tasks can add “use Ollama blob storage” (same GGUF files Ollama uses) for discovery and path resolution.
- **Build:** CGO + build tag; `DefaultLlamaCppProvider()` returns nil / `Supported() = false` when not built (same pattern as Apple FM / MLX).

---

## 5. Conclusion

| Option              | Use as primary | Notes |
|---------------------|----------------|-------|
| go-skynet/go-llama.cpp | **Yes**        | Best maturity, performance, and ecosystem fit; CGO + Metal on darwin/arm64. |
| swdunlop/llm-go     | Optional       | Lightweight / experimental only. |
| dianlight/gollama.cpp | No           | Not found; skip unless repo appears. |

**Next steps (from backlog):** Test go-skynet build on macOS/arm64 (T-1771367222946694000), then design tool schema and API (T-1771367225957468000) and Ollama manifest parsing (T-1771367227570627000) for model discovery and GGUF path resolution.
