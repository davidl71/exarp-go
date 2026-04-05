---
name: text-generate
description: Use text_generate for quick local LLM text generation. Use when you need fast, on-device text generation without full conversational context, or when other AI backends are unavailable. Good for summaries, classifications, code snippets, and simple completions.
---

# text_generate Skill

Apply this skill when you need quick text generation using exarp-go backends: Apple Foundation Models (FM chain), Ollama, report insight provider, LocalAI, or an OpenAI-compatible gateway — via the **`text_generate`** MCP tool.

## When to Use

| Scenario | Use text_generate |
|----------|-------------------|
| **Quick generation** | Need text without full LLM conversational overhead |
| **On-device / local** | Privacy-sensitive tasks (no cloud APIs), or self-hosted gateways |
| **No context needed** | Simple completions, not multi-turn conversations |
| **Fallback** | Other AI tools unavailable or overkill |
| **Classification** | Categorize text (sentiment, topic, etc.) |
| **Summarization** | Condense text to brief/detailed/key_metrics/actionable |
| **Code snippets** | Generate small code blocks or examples |

## Providers

Supported `provider` values (see `internal/tools/text_generate.go`):

| Provider | Backend | When to Use |
|----------|---------|-------------|
| `fm` (default) | `DefaultFMProvider()` / FM chain | Stock chain often probes Ollama; Apple FM when built with darwin/arm64/cgo |
| `ollama` | Ollama HTTP | Local server (`ollama serve`) |
| `insight` | `DefaultReportInsight()` | Same insight path as report/scorecard AI blurbs |
| `localai` | LocalAI-compatible server | `LOCALAI_BASE_URL` set |
| `gateway` | OpenAI-compatible router | `OPENAI_GATEWAY_BASE_URL` set |
| `auto` | Model router | Task hints or automatic backend selection |

Legacy **`mlx`** is not a valid provider; stored metadata values map to auto/fm.

## Usage

### text_generate (MCP — preferred)

```json
{
  "prompt": "Summarize this: <text>",
  "provider": "fm",
  "max_tokens": 200
}
```

```json
{
  "prompt": "Summarize this: <text>",
  "provider": "ollama",
  "max_tokens": 200
}
```

### CLI (fallback)

```bash
./bin/exarp-go -tool text_generate -args '{"prompt": "Say hello", "provider": "ollama"}'
```

## Examples

**Simple generation (fm — default):**
```json
{"prompt": "Write a Python function to reverse a string", "provider": "fm"}
```

**Classification:**
```json
{"prompt": "Classify this review as positive/negative/neutral: 'Great product, fast shipping!'", "provider": "fm"}
```

**Auto — let system pick:**
```json
{"prompt": "Explain what a monad is", "provider": "auto"}
```

## Check Availability First

```
Resource: stdio://models
Look for: data.backends.fm_available, ollama_reachable, localai_available, gateway_available
```

## Decision Flow

1. **Need quick text gen?** → Use `text_generate`
2. **Ollama running?** → `provider=ollama` or rely on FM chain / `auto`
3. **Self-hosted OpenAI-compatible?** → `provider=localai` or `gateway` when env is set
4. **Need to pick model automatically?** → `provider=auto`
5. **Need conversation/memory?** → Use a full LLM client, not `text_generate`
