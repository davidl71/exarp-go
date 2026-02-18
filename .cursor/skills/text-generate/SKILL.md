---
name: text-generate
description: Use text_generate for quick local LLM text generation. Use when you need fast, on-device text generation without full conversational context, or when other AI backends are unavailable. Good for summaries, classifications, code snippets, and simple completions.
---

# text_generate Skill

Apply this skill when you need quick text generation using local AI backends (Apple Foundation Models, MLX, Ollama) via the exarp-go MCP server.

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

## Providers

| Provider | Backend | When to Use |
|----------|---------|-------------|
| `fm` (default) | Apple Foundation Models | On Apple Silicon, macOS 26+, fast & private |
| `mlx` | MLX | Apple Silicon, bridge-based |
| `insight` | Report insight provider | Report/analysis tasks |
| `auto` | Best available | Let system choose |

## Usage

### MCP Tool (preferred)

```json
{
  "prompt": "Summarize this: <text>",
  "provider": "fm",
  "max_tokens": 200
}
```

### CLI (fallback)

```bash
./bin/exarp-go -tool text_generate -args '{"prompt": "Say hello", "provider": "fm"}'
```

## Examples

**Simple generation:**
```json
{"prompt": "Write a Python function to reverse a string", "provider": "fm"}
```

**Classification:**
```json
{"prompt": "Classify this review as positive/negative/neutral: 'Great product, fast shipping!'", "provider": "fm"}
```

**Summarization:**
```json
{"prompt": "Summarize in 2 sentences: <long text here>", "provider": "fm"}
```

## Check Availability First

Before using, check if backends are available:

```
Resource: stdio://models
Look for: data.backends.fm_available = true
```

If `fm_available` is false, fall back to `ollama` or `mlx` providers.

## Decision Flow

1. **Need quick text gen?** → Use `text_generate`
2. **On Apple Silicon?** → Use `provider=fm` (default)
3. **FM unavailable?** → Try `provider=ollama` or `provider=mlx`
4. **Need conversation/memory?** → Use full LLM, not text_generate
