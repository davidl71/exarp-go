# CGO Build Parity Matrix

This document describes which features are available in CGO vs no-CGO builds, and how we minimize drift between builds.

## Build Types

| Build | Command | Description |
|-------|---------|-------------|
| **CGO (default)** | `make build` or `CGO=1 go build` | Full features including Apple FM |
| **no-CGO** | `make build-no-cgo` or `CGO=0 go build` | Core features without Apple FM |

## Feature Matrix

| Feature | CGO Build | no-CGO Build | Shared Code |
|---------|-----------|--------------|-------------|
| **Task Discovery** | | | |
| Basic scanning (TODO/FIXME/markdown) | ✅ | ✅ | ✅ `task_discovery_common.go` |
| Apple FM semantic enhancement | ✅ | ❌ | Uses runtime `FMAvailable()` |
| **Estimation** | | | |
| Statistical estimation | ✅ | ✅ | ✅ `estimation_shared.go` + `estimation_shared_v2.go` |
| Apple FM estimation | ✅ | ❌ | ✅ Unified via runtime check |
| Ollama estimation | ✅ | ✅ | ✅ Shared |
| MLX estimation | ✅ (optional) | ❌ | ✅ Runtime check |
| **Context** | | | |
| Summarization (Apple FM) | ✅ | ❌ | ✅ `context_shared.go` |
| **LLM Backends** | | | |
| Apple Foundation Models | ✅ | ❌ | darwin/arm64 only |
| Ollama | ✅ | ✅ | All platforms |
| MLX | ✅ (optional) | ❌ | Requires CGO |
| LlamaCpp | ✅ (optional) | ❌ | Requires CGO |
| LocalAI | ✅ | ✅ | All platforms |

## Drift Minimization Strategy

We use several strategies to minimize code drift between CGO and no-CGO builds:

### 1. Shared Files (`*_shared.go`, `*_common.go`)
Common logic lives in shared files used by both builds:
- `task_discovery_common.go` - Scanner logic, ignore paths
- `estimation_shared.go` - Types, statistical estimation
- `estimation_shared_v2.go` - Unified dispatcher with runtime checks
- `context_shared.go` - Unified summarization handler
- `task_analysis_shared.go` - Analysis dispatcher
- `task_workflow_common.go` - Workflow infrastructure

### 2. Runtime Feature Detection
Instead of compile-time build tags, we use runtime checks:
```go
// Both CGO and no-CGO use the same handler
func HandleContextSummarizeShared(...) {
    if !FMAvailable() {
        return error
    }
    // Use Apple FM
}
```

### 3. Unified Dispatchers
Handlers check feature availability at runtime:
```go
func HandleEstimationNative(...) {
    if FMAvailable() {
        // Try Apple FM
    }
    if MLAvailable() {
        // Try MLX
    }
    // Fallback to statistical
}
```

## Platform Requirements

| Feature | Platform | CGO Required |
|---------|----------|--------------|
| Apple FM | darwin/arm64 | Yes |
| MLX | darwin/arm64 | Yes |
| LlamaCpp | All | Optional (CGO) |
| Ollama | All | No |
| LocalAI | All | No |
