# llama.cpp Tool Schema and API Design

**Tag hints:** `#design` `#llamacpp` `#api`

Design for the exarp-go `llamacpp` MCP tool and provider. Uses go-skynet/go-llama.cpp; integrates as a new tool, `text_generate` provider, and optional FM chain backend.

---

## 1. Tool Overview

| Property | Value |
|----------|-------|
| **Tool name** | `llamacpp` |
| **Library** | github.com/go-skynet/go-llama.cpp |
| **Format** | GGUF only |
| **CGO** | Required |

**Actions:** `status`, `models`, `generate`, `load`, `unload`

---

## 2. Action Parameters and Responses

### 2.1 Common Parameters (all actions)

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `action` | string | `"status"` | One of: `status`, `models`, `generate`, `load`, `unload` |

### 2.2 status

**Purpose:** Check if llama.cpp binding is available and optionally if a model is loaded.

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| — | — | — | No extra params |

**Response (success):**

```json
{
  "success": true,
  "method": "native_go",
  "available": true,
  "loaded": false,
  "loaded_model": null,
  "platform": "darwin/arm64",
  "metal": true,
  "tip": "Use load action to load a GGUF file, then generate"
}
```

**Response (llamacpp not built):**

```json
{
  "success": true,
  "available": false,
  "message": "llamacpp not available (requires CGO + go-llama.cpp build)",
  "tip": "Build with: make build-llamacpp or CGO_ENABLED=1 -tags llamacpp"
}
```

---

### 2.3 models

**Purpose:** List discoverable GGUF models. Sources: config path, env paths, and optionally Ollama blob storage.

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `include_ollama_blobs` | boolean | No | If true, scan Ollama blob dir for GGUF files (default: true when OLLAMA_MODELS set) |

**Response:**

```json
{
  "success": true,
  "method": "native_go",
  "models": [
    {
      "path": "/path/to/gguf/model-Q4_K_M.gguf",
      "name": "model-Q4_K_M",
      "size_bytes": 4200000000,
      "source": "custom"
    },
    {
      "path": "/Users/dlowes/.ollama/models/blobs/sha256-abc123",
      "name": "llama3.2:3b",
      "size_bytes": 2100000000,
      "source": "ollama"
    }
  ],
  "count": 2
}
```

---

### 2.4 load

**Purpose:** Load a GGUF model into memory for subsequent `generate` calls.

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Path to `.gguf` file or alias (e.g. model name from `models`) |
| `num_gpu_layers` | integer | No | GPU layers to offload (default: -1 = auto) |
| `context_size` | integer | No | Context window size (default: 2048) |

**Response:**

```json
{
  "success": true,
  "method": "native_go",
  "model": "/path/to/model.gguf",
  "loaded": true,
  "message": "Model loaded successfully"
}
```

**Error (model already loaded):**

```json
{
  "success": false,
  "error": "model already loaded: llama3.2. Unload first.",
  "loaded_model": "llama3.2"
}
```

---

### 2.5 unload

**Purpose:** Unload the current model from memory.

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| — | — | — | No extra params |

**Response:**

```json
{
  "success": true,
  "method": "native_go",
  "unloaded": "llama3.2",
  "message": "Model unloaded"
}
```

**Response (nothing loaded):**

```json
{
  "success": true,
  "unloaded": null,
  "message": "No model was loaded"
}
```

---

### 2.6 generate

**Purpose:** Generate text from the loaded model.

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | Input prompt |
| `max_tokens` | integer | No | Max tokens to generate (default: 512) |
| `temperature` | number | No | Sampling temperature 0–2 (default: 0.7) |
| `top_k` | integer | No | Top-k sampling (default: 40) |
| `top_p` | number | No | Top-p (nucleus) sampling (default: 0.9) |
| `num_threads` | integer | No | CPU threads (default: runtime.NumCPU()) |

**Response:**

```json
{
  "success": true,
  "method": "native_go",
  "response": "The generated text...",
  "model": "/path/to/model.gguf",
  "prompt_tokens": 12,
  "generated_tokens": 64
}
```

**Error (no model loaded):**

```json
{
  "success": false,
  "error": "no model loaded. Use load action first."
}
```

---

## 3. Error Handling Patterns

| Situation | HTTP/JSON-RPC | Response body |
|-----------|---------------|---------------|
| Unknown action | 400 | `{"error": "unknown action: X (use 'status', 'models', ...)"}` |
| Missing required param | 400 | `{"error": "prompt parameter is required for generate action"}` |
| Model path not found | 404 | `{"error": "model file not found: /path/to/model.gguf"}` |
| Load failed (invalid GGUF) | 500 | `{"error": "failed to load model: ...", "details": "..."}` |
| Generate failed | 500 | `{"error": "llamacpp generate failed: ..."}` |
| Not built (nocgo) | 200 | `{"success": true, "available": false, "message": "llamacpp not available"}` |

**Go convention:** Return `fmt.Errorf("llamacpp %s: %w", action, err)` for tool-level wrapping.

---

## 4. Ollama Blob Storage Integration

**Ollama storage layout (OLLAMA_MODELS or ~/.ollama/models):**

```
$OLLAMA_MODELS/
├── blobs/           # sha256-<hash> files (raw GGUF or layers)
└── manifests/       # JSON manifests per model tag
```

**Integration approach:**

1. **Discovery:** Parse `manifests/registry.json` (or iterate manifests) to map tag → blob digests.
2. **Path resolution:** Resolve digest to blob path `blobs/sha256-<hash>`.
3. **Reuse:** Expose these paths in `models` action as `source: "ollama"`.
4. **No duplication:** Use same GGUF files Ollama uses; no copy.

**Env:** `OLLAMA_MODELS` — custom model root. Default: `~/.ollama/models` (platform-specific).

**Follow-up:** Separate task for full manifest parsing; Phase 1 can use explicit `model` path only.

---

## 5. Build Requirements

| Requirement | Value |
|-------------|-------|
| **CGO** | `CGO_ENABLED=1` |
| **Build tag** | `llamacpp` (opt-in) or `darwin && arm64 && cgo` (same as Apple FM for initial scope) |
| **Library** | go-skynet/go-llama.cpp (submodule or vendored) |
| **Bindings** | `libbinding.a` via `BUILD_TYPE=metal make libbinding.a` (darwin/arm64) |
| **CGO flags** | `CGO_LDFLAGS="-framework Foundation -framework Metal ..."` for Metal |

**Makefile targets (suggested):**

- `make build-llamacpp` — Build with `-tags llamacpp`, `LIBRARY_PATH`/`C_INCLUDE_PATH` set
- `make build-no-cgo` — Unchanged; exarp-go without llamacpp/Apple FM

**Stub when not built:** `DefaultLlamaCppProvider()` returns nil; `Supported() = false`; tool handler returns `available: false`.

---

## 6. text_generate Provider Integration

**Provider:** `provider=llamacpp`

**Schema change (registry.go):**

```go
"provider": map[string]interface{}{
    "type":    "string",
    "enum":    []string{"fm", "ollama", "insight", "mlx", "localai", "llamacpp", "auto"},
    "default": "fm",
    "description": "Backend: ... llamacpp (local GGUF via go-llama.cpp), or auto",
},
```

**TextGenerator implementation:**

```go
// llamacpp_provider.go (or llamacpp_provider_nocgo.go for stub)
type llamacppTextGenerator struct {
    // holds loaded model state if process-scoped
}

func (l *llamacppTextGenerator) Supported() bool {
    return LlamaCppAvailable() // checks build + model loaded
}

func (l *llamacppTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
    // Call loaded model Predict; map params to llama.SetTokens, llama.SetTemperature, etc.
}
```

**Default model:** Use `LLAMACPP_MODEL` env or last-loaded model. If none, `Supported()` = false.

---

## 7. FM Chain Placement

**Current:** Apple FM → Ollama → stub

**New:** Apple FM → **LlamaCpp** → Ollama → stub

**Rationale:** LlamaCpp is local, no server, and uses GGUF directly. Place after Apple FM (prefer on-device) and before Ollama (server-based).

**Code change (fm_chain.go):**

```go
backends := []TextGenerator{}
if g := appleFMIfAvailable(); g != nil {
    backends = append(backends, g)
}
if g := llamaCppIfAvailable(); g != nil {
    backends = append(backends, g)
}
backends = append(backends, ollamaTG, stub)
```

---

## 8. File Layout

| File | Purpose |
|------|---------|
| `internal/tools/llamacpp_native.go` | Handler + status/models/load/unload/generate (build tag: `llamacpp`) |
| `internal/tools/llamacpp_native_nocgo.go` | Stub when not built |
| `internal/tools/llamacpp_provider.go` | TextGenerator impl; `DefaultLlamaCppProvider()` |
| `internal/tools/llamacpp_ollama.go` | Ollama blob discovery helpers (optional Phase 2) |

---

## 9. Example Requests

**status:**

```json
{"action": "status"}
```

**models:**

```json
{"action": "models", "include_ollama_blobs": true}
```

**load:**

```json
{"action": "load", "model": "/path/to/llama-3.2-3b-Q4_K_M.gguf", "context_size": 4096}
```

**generate:**

```json
{
  "action": "generate",
  "prompt": "Explain quicksort in one sentence.",
  "max_tokens": 128,
  "temperature": 0.7
}
```

**unload:**

```json
{"action": "unload"}
```

---

## 10. References

- [go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp)
- [LLAMACPP_EVALUATION.md](./LLAMACPP_EVALUATION.md)
- [LLM_NATIVE_ABSTRACTION_PATTERNS.md](./LLM_NATIVE_ABSTRACTION_PATTERNS.md)
- [internal/tools/fm_chain.go](../internal/tools/fm_chain.go)
- [internal/tools/text_generate.go](../internal/tools/text_generate.go)
