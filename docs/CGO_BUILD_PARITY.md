# CGO vs non-CGO build split (intentional)

This document is the **canonical reference** for why exarp-go uses compile-time build constraints alongside shared (`*_common.go` / `*_shared.go`) code, which files participate, and how to keep **darwin/arm64/cgo** and **everything else** aligned.

## Build constraint (exact)

The primary split is **not** “CGO vs no CGO” alone. Go selects one of two variants using:

| Variant | Constraint | Typical environments |
|--------|-------------|----------------------|
| **“CGO native”** | `//go:build darwin && arm64 && cgo` | macOS Apple Silicon, `CGO_ENABLED=1`, full toolchain |
| **“nocgo fallback”** | `//go:build !(darwin && arm64 && cgo)` | Default **`make build` / `make b`** (always `CGO_ENABLED=0`), Linux/Windows, macOS amd64, CI tests (`make test-go`), and any build where the constraint above is false |

**Important:** The usual exarp-go binary from `make build` uses **nocgo fallback** files even on Apple Silicon. FM-enhanced **task_discovery** scanners require a build with **`CGO_ENABLED=1`** on darwin/arm64 (for example `make go-build` when a C compiler is present).

So: **Linux with CGO** still compiles the **nocgo fallback** files for the tools listed below. Naming uses `_nocgo` as “not the Apple-Silicon-CGO variant,” not “CGO disabled globally.”

Verification examples:

```bash
# Default unit tests already use CGO off (same variant as make build)
make test-go

# Explicit no-CGO build (same as make build / make b)
make build
```

On **Apple Silicon**, `make go-build`, `make build-debug`, and `make build-race` enable **CGO** when a C compiler is available, so the **`darwin && arm64 && cgo`** files are selected. That is the path for FM-enhanced **task_discovery** scanners. Apple FM / Swift bridge details: [APPLE_FOUNDATION_MODELS_TESTING.md](APPLE_FOUNDATION_MODELS_TESTING.md).

## Why a compile-time split exists

1. **Symbols that cannot compile on all targets**  
   Apple FM–enhanced task discovery lives in code that assumes the darwin/arm64/cgo stack (see `task_discovery_native_scanners.go`). The nocgo tree uses regex/basic scanners only.

2. **Single `handleXNative` entry per tool**  
   Go requires exactly one definition of `handleTaskDiscoveryNative`, `handleEstimationNative`, and `handleContextSummarizeNative` per build. **Estimation** and **context** define those handlers once in unconstrained `*_shared*.go` files; **task_discovery** still uses a **build-tagged pair** (`native` + `native_nocgo`) plus shared `task_discovery_common.go` for portable pieces like `scanGitJSON`.

3. **Runtime vs compile-time**  
   Most LLM behavior uses **`FMAvailable()`** and **`DefaultFMProvider()`** in **unconstrained** files so one code path runs everywhere. The **task_discovery** scanner stack is the main exception where duplication is traded for compile safety.

## File inventory (`internal/tools`)

| Files | Constraint | Role |
|-------|------------|------|
| `task_discovery_native.go` | `darwin && arm64 && cgo` | `handleTaskDiscoveryNative`; FM-aware `scanComments`, `scanMarkdown`, `findOrphanTasks`, planning scan |
| `task_discovery_native_scanners.go` | `darwin && arm64 && cgo` | Apple FM helpers, `scanPlanningDocs` (git JSON: `task_discovery_common.go`) |
| `task_discovery_native_nocgo.go` | `!(darwin && arm64 && cgo)` | Same tool entry; `*Basic` scanners; no FM enhancement |

**Unconstrained (all builds):** `task_discovery_common.go` (includes **`scanGitJSON`**), `estimation_shared.go`, `estimation_shared_v2.go` (includes **`handleEstimationNative`** shim → `HandleEstimationNative`), `context_shared.go` (includes **`handleContextSummarizeNative`** shim), `fm_chain.go`, `fm_provider.go`, `fm_ollama.go`, `mlx_native_nocgo.go` (single MLX stub handler), and most other tools.

## Shared “common” layer (include here first)

When changing behavior that must match both variants:

| Area | Shared files |
|------|----------------|
| Task discovery | `task_discovery_common.go` — ignore paths, `createTasksFromDiscoveries`, JSON load helpers, types |
| Estimation | `estimation_shared.go`, `estimation_shared_v2.go` — `HandleEstimationNative`, stats, FM at runtime |
| Context | `context_shared.go` — `HandleContextSummarizeShared` (FM chain via `DefaultFMProvider()`; honors `ctx`) |
| FM abstraction | `fm_provider.go`, `fm_chain.go`, `fm_ollama.go` |

New logic that does **not** need Apple-only imports should land in these files (or other untagged helpers), not in the `*_native*.go` pair.

## Behavioral parity checklist

When editing **task discovery**:

- [ ] **CGO variant:** `task_discovery_native.go` + `task_discovery_native_scanners.go`
- [ ] **nocgo variant:** `task_discovery_native_nocgo.go` (`*Basic` functions)
- [ ] **Shared:** `task_discovery_common.go` for anything both need
- [ ] Same actions: `comments`, `markdown`, `orphans`, `git_json`, `planning_links`, `all`
- [ ] Same report output shape: `summary`, optional `create_tasks`, `report_path`
- [ ] `git_json`: **`scanGitJSON`** lives only in `task_discovery_common.go`

**Estimation / context:** `handleEstimationNative` and `handleContextSummarizeNative` live in `estimation_shared_v2.go` and `context_shared.go` (no build-tagged shim files).

## Other build tags (not the darwin/arm64/cgo matrix)

| Location | Tag | Purpose |
|----------|-----|---------|
| `internal/database/firestore.go` | `with_firestore` | Optional Firestore backend |
| `scripts/*.go` | `ignore` | Dev utilities not in normal `go build` |

Optional **llamacpp** / **apple_fm** test tags are documented in [llamacpp-build-requirements.md](llamacpp-build-requirements.md) and [APPLE_FOUNDATION_MODELS_TESTING.md](APPLE_FOUNDATION_MODELS_TESTING.md).

## Feature matrix (summary)

| Feature | CGO darwin/arm64 variant | nocgo / other platforms | Primary mechanism |
|---------|--------------------------|-------------------------|-------------------|
| Task discovery basic scans | Yes | Yes | Shared + Basic scanners |
| Task discovery FM enhancement | Yes | No | Build-tagged scanners |
| Estimation / context | Yes | Yes | Shared handlers + `FMAvailable()` |
| Ollama / LocalAI | Yes | Yes | Untagged |
| MLX generate | Stub / limited | Stub | `mlx_native_nocgo.go`; bridge builds optional |

## Related docs

- [TASK_TOOLS_SHARED_PATTERNS.md](TASK_TOOLS_SHARED_PATTERNS.md) — task-tool conventions and shared layers
- [ARCHITECTURE.md](ARCHITECTURE.md) § Tool handler pattern
- [.cursor/rules/code-map.mdc](../.cursor/rules/code-map.mdc) — file → tool map

## Build Types (Make / daily use)

| Build | Command | CGO | Split tools (`task_discovery`, etc.) |
|-------|---------|-----|--------------------------------------|
| **Default binary** | `make build` / `make b` | **Off** (`CGO_ENABLED=0`) | **nocgo** files on all platforms |
| **go-build** | `make go-build` | **On** on Darwin arm64 if `cc` exists; else off | **cgo** variant on Mac Silicon when CGO on |
| **Debug / race** | `make build-debug` / `make build-race` | Same rule as `go-build` on Darwin arm64 | Same as `go-build` |

Plain `go build ./cmd/server` without setting `CGO_ENABLED` follows the Go default (often CGO on where toolchains exist). Prefer documenting **explicit** `CGO_ENABLED=0` vs `1` when reproducing parity issues.

## Drift minimization (strategy summary)

1. **Prefer shared files** for any logic that does not import platform-only packages.
2. **Use runtime checks** (`FMAvailable()`, `DefaultFMProvider()`) inside shared code when possible (estimation, context, task_analysis, etc.).
3. **Keep build-tagged pairs thin** when they only delegate to shared functions; prefer a single unconstrained file for thin entrypoints (estimation/context pattern).
4. **Reserve `task_discovery`’s split** for true FM/scanner differences; push JSON/git/report aggregation to `task_discovery_common.go`.
