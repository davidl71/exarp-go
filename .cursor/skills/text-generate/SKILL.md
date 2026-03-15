---
name: text-generate
description: Use text_generate for quick local LLM text generation. Use when you need fast, on-device text generation without full conversational context, or when other AI backends are unavailable. Good for summaries, classifications, code snippets, and simple completions.
---

# text_generate Skill

Apply this skill when you need quick text generation using local AI backends (Apple Foundation Models, MLX, Ollama, llama.cpp) via the exarp-go MCP server.

## When to Use

| Scenario | Use text_generate |
|----------|-------------------|
| **Quick generation** | Need text without full LLM conversational overhead |
| **On-device only** | Privacy-sensitive tasks (no cloud APIs) |
| **No context needed** | Simple completions, not multi-turn conversations |
| **Fallback** | Other AI tools unavailable or overkill |
| **Classification** | Categorize text (sentiment, topic, etc.) |
| **Summarization** | Condense text to brief/detailed/key_metrics/actionable |
| **Code snippets** | Generate small code blocks or examples |
| **Offline use** | No network or server required (use `llamacpp`) |

## Providers

| Provider | Backend | When to Use |
|----------|---------|-------------|
| `fm` (default) | Apple Foundation Models | On Apple Silicon, macOS 26+, fast & private |
| `mlx` | MLX | Apple Silicon, bridge-based |
| `llamacpp` | llama.cpp (go-llama.cpp) | Direct GGUF inference, no server needed. Requires `llamacpp` build tag |
| `ollama` | Ollama server | Local server (`ollama serve`), broad model support |
| `insight` | Report insight provider | Report/analysis tasks |
| `localai` | LocalAI gateway | Self-hosted OpenAI-compatible API |
| `gateway` | Cloud/custom gateway | Custom model gateway with model override support |
| `auto` | Best available | Let system choose via model router |

## llamacpp Provider — Detail

### What it is

`provider=llamacpp` routes text_generate through the embedded llama.cpp engine (go-llama.cpp bindings). It loads GGUF model files directly — no Ollama server, no HTTP round-trip, no external process required.

### When to prefer llamacpp

- **Offline / air-gapped** — no server or network needed
- **Low latency** — no HTTP overhead, direct in-process inference
- **Direct GGUF** — you have a `.gguf` file and want to use it without Ollama
- **Apple FM unavailable** — macOS version < 26, or non-Apple hardware
- **Server-free fallback** — Ollama server not running (`ollama serve` not active)
- **GPU offload control** — fine-grained control over `gpu_layers` and `context_size`

### Build requirement

llamacpp requires the `llamacpp` build tag and CGO. The standard binary does NOT include it.

```bash
# Build with llamacpp support
make build-apple-fm   # includes llamacpp on darwin/arm64

# Verify availability
llamacpp  action=status
# Look for: "available": true
```

Check via resource before calling:
```
Resource: stdio://models
Look for: data.backends.llamacpp_available = true
```

### Model selection for llamacpp

llamacpp resolves model names using a two-tier alias system:

1. **Built-in aliases** (shorthand → Ollama model name):
   - `llama`, `llama3`, `llama3b`, `llama1b` → llama3.2 family
   - `phi`, `phi3`, `phi4` → phi family
   - `mistral`, `mistral7b` → mistral family
   - `gemma`, `gemma2`, `gemma3` → gemma family
   - `qwen`, `qwen2`, `qwen-coder` → qwen2.5 family
   - `deepseek`, `deepseek-r1` → deepseek-r1
   - `codellama`, `code` → codellama
   - `starcoder` → starcoder2

2. **User aliases** from `.exarp/llamacpp.json` (overrides built-ins):
   ```json
   {
     "aliases": {
       "my-llama": "llama3.2:latest",
       "local-7b": "/path/to/my-model.gguf",
       "deepseek": "deepseek-r1:7b"
     }
   }
   ```

Models are resolved in order: alias lookup → Ollama blob path → direct file path.

### llamacpp-specific parameters

When calling the `llamacpp` tool directly (not via text_generate), additional parameters are available:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `gpu_layers` | -1 | Layers to offload to GPU (-1 = all available) |
| `context_size` | 2048 | Context window size in tokens |
| `temperature` | 0.7 | Sampling temperature |
| `max_tokens` | 512 | Max tokens to generate |

## Usage

### text_generate with llamacpp (MCP tool — preferred)

```json
{
  "prompt": "Summarize this: <text>",
  "provider": "llamacpp",
  "max_tokens": 200
}
```

### text_generate with fm (default — Apple Silicon)

```json
{
  "prompt": "Summarize this: <text>",
  "provider": "fm",
  "max_tokens": 200
}
```

### llamacpp tool — direct generation

```json
{
  "action": "generate",
  "prompt": "Write a Python function to reverse a string",
  "max_tokens": 300,
  "temperature": 0.5
}
```

### llamacpp tool — load a specific model first

```json
{"action": "load", "model": "llama3b"}
```

Then generate:
```json
{"action": "generate", "prompt": "Explain recursion briefly"}
```

### llamacpp tool — list available models

```json
{"action": "models"}
```

Returns GGUF paths discovered in Ollama blob storage plus the full alias map.

### CLI (fallback)

```bash
./bin/exarp-go -tool text_generate -args '{"prompt": "Say hello", "provider": "llamacpp"}'
./bin/exarp-go -tool llamacpp -args '{"action": "status"}'
```

## Examples

**Simple generation (fm — default):**
```json
{"prompt": "Write a Python function to reverse a string", "provider": "fm"}
```

**Simple generation (llamacpp — offline/server-free):**
```json
{"prompt": "Write a Python function to reverse a string", "provider": "llamacpp"}
```

**Classification:**
```json
{"prompt": "Classify this review as positive/negative/neutral: 'Great product, fast shipping!'", "provider": "fm"}
```

**Summarization:**
```json
{"prompt": "Summarize in 2 sentences: <long text here>", "provider": "llamacpp"}
```

**Auto — let system pick best available backend:**
```json
{"prompt": "Explain what a monad is", "provider": "auto"}
```

## Check Availability First

Before using, check if backends are available:

```
Resource: stdio://models
Look for: data.backends.fm_available = true
         data.backends.llamacpp_available = true
```

Or check llamacpp status directly:
```json
{"action": "status"}
```
Returns `available`, `model_loaded`, GPU info, and platform.

## Decision Flow

1. **Need quick text gen?** → Use `text_generate`
2. **On Apple Silicon, macOS 26+?** → Use `provider=fm` (default, fastest)
3. **FM unavailable or non-Apple?** → Try `provider=llamacpp` (no server needed) or `provider=ollama` or `provider=mlx`
4. **Offline / no server / air-gapped?** → Use `provider=llamacpp` (requires `llamacpp` build tag)
5. **Want direct GGUF with GPU offload control?** → Use `provider=llamacpp` or the `llamacpp` tool directly
6. **Need to pick model automatically?** → Use `provider=auto` (model router selects)
7. **Need conversation/memory?** → Use full LLM, not text_generate
