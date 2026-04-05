# AI/LLM Integration Guide

**Last Updated**: 2026-03-08  
**Version**: v0.3.5  
**Scope**: Complete guide to exarp-go's AI and LLM capabilities

This guide covers AI/LLM integration in exarp-go: backend selection, `text_generate`, Ollama, FM chain, LocalAI/gateway, and OpenCode MCP usage. The **`mlx` MCP tool was removed** (2026).

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Supported Backends](#supported-backends)
3. [Backend Selection Strategy](#backend-selection-strategy)
4. [OpenCode Integration](#opencode-integration)
5. [Tool Reference](#tool-reference)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Check Available Backends

```bash
# Via CLI
exarp-go -tool ollama -args '{"action":"status"}'

# Via MCP resource (fastest)
# Read stdio://models for full backend availability (fm_available, ollama_reachable, localai, gateway)
```

### Generate Text with Auto-Backend Selection

```bash
exarp-go -tool text_generate -args '{
  "provider": "auto",
  "prompt": "Explain dependency injection in Go",
  "temperature": 0.7
}'
```

### Recommended Setup (macOS)

```bash
# Install Ollama (recommended for local inference)
brew install ollama
ollama pull llama3.2

# Verify
ollama list
exarp-go -tool ollama -args '{"action":"status"}'
```

---

## Supported Backends

exarp-go exposes **local and gateway-backed** generation primarily via **`text_generate`** and the **`ollama`** tool:

| Backend | Status | Best For | Entry |
|---------|--------|----------|--------|
| **Ollama** | ✅ Production | Local inference, variety | `ollama` tool; `text_generate` `provider=ollama` |
| **FM chain** | ✅ Production | On-device / default chain | `text_generate` `provider=fm` |
| **Insight** | ✅ Available | Report-style blurbs | `text_generate` `provider=insight` |
| **LocalAI** | ✅ Optional | OpenAI-compatible self-host | `text_generate` `provider=localai` |
| **Gateway** | ✅ Optional | OpenAI-compatible router | `text_generate` `provider=gateway` |
| **Auto** | ✅ Recommended | Router selection | `text_generate` `provider=auto` |

### Backend Priorities (Auto Mode)

When using `provider=auto`, exarp-go selects backends in this order:

1. **FM chain** (when `FMAvailable()` / build allows)
2. **Ollama** (when reachable)
3. **LocalAI** / **gateway** (when configured)
4. **Insight** or other fallbacks per router rules

---

## Backend Selection Strategy

### Decision Tree

```
Are you on macOS?
├─ Yes → Do you need lowest latency?
│   ├─ Yes → Use Apple Foundation Models
│   └─ No → Do you want local models?
│       ├─ Yes → Use Ollama (recommended) or `text_generate` `provider=fm`
│       └─ No → Use Insight/cloud
└─ No (Linux/Windows)
    ├─ Do you want local models?
    │   ├─ Yes → Use Ollama
    │   └─ No → Use Insight/cloud
    └─ Do you need OpenAI compatibility?
        └─ Yes → Use LocalAI
```

### Use Cases

| Use Case | Recommended Backend | Reason |
|----------|---------------------|--------|
| **Task execution** | Ollama (llama3.2) | Good balance of quality and speed |
| **Code generation** | Apple FM or Ollama (codellama) | Low latency, good code understanding |
| **Analysis/research** | Ollama (larger models) | Better reasoning, acceptable latency |
| **Quick queries** | Apple FM | Lowest latency |
| **Offline work** | Ollama (or FM on-device) | No cloud required |
| **Production** | Ollama + fallback to Insight | Reliability + backup |

---

### MLX (removed from exarp-go)

exarp-go **does not** ship an `mlx` MCP tool or MLX Python bridge. For Apple Silicon MLX models inside **OpenCode**, configure OpenCode’s MLX provider separately ([OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md) §5). For exarp-go, use **Ollama** or **`text_generate`** (`provider=fm`, `ollama`, `auto`).

---

## OpenCode Integration

### Overview

**OpenCode** is an AI-powered code assistant that works via MCP (Model Context Protocol). exarp-go integrates seamlessly with OpenCode via the MCP server.

### Configuration

The `opencode.json` file configures exarp-go as an MCP server for OpenCode:

**Option 1: Direct Binary (Recommended for exarp-go repo)**
```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/Projects/mcp/exarp-go/bin/exarp-go"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/mcp/exarp-go",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Option 2: Wrapper Script (Recommended for other projects)**
```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/path/to/your/project",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Why two options?**
- **Direct binary**: Fastest, best for exarp-go development
- **Wrapper script**: Auto-resolves binary location, best for other projects

### Key Configuration Parameters

| Parameter | Purpose | Example |
|-----------|---------|---------|
| `command` | Path to exarp-go binary or wrapper | Direct: `bin/exarp-go`<br>Wrapper: `run_exarp_go.sh` |
| `PROJECT_ROOT` | Workspace root directory | `/path/to/your/project` |
| `EXARP_MIGRATIONS_DIR` | Todo2 migrations location | `/path/to/exarp-go/migrations` |
| `EXARP_WATCH` | Auto-reload on changes (0=off) | `0` (stable) or `1` (dev) |

**Setup Instructions**:
1. For **exarp-go repo**: Use direct binary (`bin/exarp-go`)
2. For **other projects**: Use wrapper script (`run_exarp_go.sh`)
3. Always set `PROJECT_ROOT` to your current project
4. Use `EXARP_WATCH=0` for production stability
5. See `opencode.json.template` for quick setup

### OpenCode-Specific Features

exarp-go includes optimizations for OpenCode:

1. **CLI Flags** for machine-readable output:
   - `--quiet`: Suppress verbose output
   - `--json`: Structured JSON responses
   - `--concise`: Strip emojis and decorative elements

2. **Tool Descriptions**: Optimized hints for MCP clients
   - Each tool includes `[HINT: ...]` for quick understanding
   - Action-based tool design (vs. multiple specialized tools)

3. **Resource Access**: Fast read access via `stdio://`
   - `stdio://tools` - Full tool catalog
   - `stdio://tasks` - Task list
   - `stdio://models` - Backend availability
   - No process spawns required

### Usage in OpenCode

```bash
# Start OpenCode with exarp-go
opencode --config opencode.json

# OpenCode can now access exarp-go tools:
# - Task management (list, create, update)
# - Project health checks
# - LLM generation (via any backend)
# - Testing and linting
# - Security scanning
```

### OpenCode + exarp-go LLM workflow

```bash
# 1. Configure OpenCode with exarp-go
opencode --config opencode.json

# 2. In OpenCode, use exarp-go's LLM tools
#    OpenCode sends: {"tool": "text_generate", "args": {"provider": "auto", ...}}
#    
# 3. exarp-go auto-selects best backend (provider=auto):
#    - FM chain / Ollama / gateway per model router and env
```

### Best Practices for OpenCode

1. **Use `provider=auto`** in `text_generate` for best backend selection
2. **Set PROJECT_ROOT** correctly for workspace-aware operations
3. **Disable watch mode** (`EXARP_WATCH=0`) for stability
4. **Use resources** (`stdio://`) instead of tool calls when possible
5. **Enable structured output** with `--json` flag for parsing

---

## Tool Reference

### Universal Interface: `text_generate`

The **recommended** way to use LLMs in exarp-go:

```bash
exarp-go -tool text_generate -args '{
  "provider": "auto",
  "prompt": "Your prompt here",
  "temperature": 0.7,
  "max_tokens": 1000
}'
```

**Providers:**
- `auto` - Auto-select best backend (recommended)
- `fm` - FM chain / Apple FM when available
- `ollama` - Ollama
- `insight` - Report insight / FM chain path
- `localai` - LocalAI (`LOCALAI_BASE_URL`)
- `gateway` - OpenAI-compatible gateway (`OPENAI_GATEWAY_BASE_URL`)

### Backend-Specific Tools

#### Apple Foundation Models (via `text_generate`)

There is no separate `apple_foundation_models` MCP tool in the default registry; use:

```bash
exarp-go -tool text_generate -args '{
  "provider": "fm",
  "prompt": "Explain error handling in Go",
  "temperature": 0.7
}'
```

#### Ollama

```bash
# Status
exarp-go -tool ollama -args '{"action":"status"}'

# List models
exarp-go -tool ollama -args '{"action":"models"}'

# Pull model
exarp-go -tool ollama -args '{"action":"pull","model":"llama3.2"}'

# Generate
exarp-go -tool ollama -args '{
  "action": "generate",
  "model": "llama3.2",
  "prompt": "Write a Go function to reverse a string"
}'

# Quality check
exarp-go -tool ollama -args '{"action":"quality","model":"llama3.2"}'

# Hardware info
exarp-go -tool ollama -args '{"action":"hardware"}'
```

### High-Level AI Tools

#### Model Recommendations

```bash
# Get model recommendation for task
exarp-go -tool recommend -args '{
  "action": "model",
  "task_description": "Code generation for REST API",
  "constraints": ["local", "fast"]
}'
```

#### Plan and Execute

```bash
# AI-powered task breakdown and execution
exarp-go -tool fm_plan_and_execute -args '{
  "task": "Implement user authentication",
  "backend": "ollama"
}'
```

#### Task Execution

```bash
# Execute Todo2 task with AI
exarp-go -tool task_execute -args '{
  "task_id": "T-123",
  "backend": "auto"
}'
```

---

## Best Practices

### 1. Backend Selection

✅ **Do:**
- Use `provider=auto` for automatic backend selection
- Check backend availability with `stdio://models` resource
- Fallback to cloud when local backends unavailable
- Use Apple FM for quick queries on macOS
- Use Ollama for complex reasoning tasks

❌ **Don't:**
- Hardcode specific backends (reduces portability)
- Assume backend availability without checking
- Assume a backend is up without checking `stdio://models` or `ollama` status
- Ignore hardware constraints (memory, GPU)

### 2. Prompt Engineering

✅ **Do:**
- Be specific and concise
- Include context (code snippets, error messages)
- Use appropriate temperature (0.2 for code, 0.7 for creative)
- Limit max_tokens to reasonable values (500-1000)

❌ **Don't:**
- Send excessive context (watch token limits)
- Use vague prompts
- Expect consistency across backends

### 3. Error Handling

```bash
# Always check status before using a backend
STATUS=$(exarp-go -tool ollama -args '{"action":"status"}' | jq -r '.success')

if [ "$STATUS" = "true" ]; then
  # Use Ollama
  exarp-go -tool text_generate -args '{"provider":"ollama",...}'
else
  # Fallback to auto
  exarp-go -tool text_generate -args '{"provider":"auto",...}'
fi
```

### 4. Performance Optimization

- **Cache responses** when possible (deterministic prompts)
- **Use streaming** for long responses (if supported)
- **Batch requests** when doing bulk operations
- **Monitor token usage** to avoid excessive costs/latency

### 5. Model Selection

| Task Type | Model Size | Recommended Model |
|-----------|-----------|-------------------|
| Quick queries | Small (3B) | Apple FM or llama3.2 |
| Code generation | Medium (7-13B) | codellama or llama3.2 |
| Analysis | Large (30B+) | Deepseek-coder or mixtral |
| Production | Stable | llama3.2 (Ollama) |

---

## Troubleshooting

### Ollama Not Available

```bash
# Check if Ollama is running
ollama list

# If not running, start Ollama service
brew services start ollama

# Or run manually
ollama serve
```

### Apple FM Not Available

- **macOS only**: Apple Foundation Models require macOS
- **Version**: Requires recent macOS version
- **Permissions**: May require approval in System Settings

### Memory Issues

- **Large models**: Use quantized versions (4-bit, 8-bit)
- **Ollama**: Use smaller models (3B instead of 13B); adjust `max_tokens`

### Performance Issues

- **Slow generation**: Check backend status (may be downloading models)
- **High latency**: Switch to smaller model or faster backend
- **Timeout**: Increase timeout or reduce max_tokens

### Backend Selection Issues

```bash
# Debug backend availability
exarp-go -tool list_resources -args '{}'

# Check which backends are available
# Read stdio://models response for fm_available, ollama_reachable, localai_available, gateway_available
```

---

## Advanced Topics

### Custom Backend Configuration

exarp-go backends can be configured via environment variables:

```bash
# Ollama
export OLLAMA_HOST=http://localhost:11434

# LocalAI
export LOCALAI_HOST=http://localhost:8080

# Gateway
export GATEWAY_URL=http://your-gateway.com
```

### Multi-Backend Strategy

For production systems, use multiple backends with fallback:

1. **Primary**: Ollama (local, fast)
2. **Secondary**: Apple FM (local, faster but limited)
3. **Tertiary**: Cloud (Insight) for high availability

Implementation:

```bash
exarp-go -tool text_generate -args '{"provider":"auto",...}'
```

The `auto` provider implements this strategy automatically.

### Integration with Other Tools

#### With Cursor

- Cursor uses its own LLM backend
- exarp-go tools provide *additional* capabilities
- Use exarp-go for task-specific operations (task_execute, etc.)

#### With GitHub Copilot

- Copilot handles code completion
- exarp-go handles project automation, task management
- Complementary, not competitive

---

## Related Documentation

- **`docs/EXARP_ABILITIES_AUDIT.md`** - Complete tool reference (39 tools)
- **`docs/TASK_TOOLS_GUIDE.md`** - Task management tools
- **`docs/INDEX.md`** - Full documentation index
- **`.cursor/skills/`** - Cursor skill documentation
- **`AGENTS.md`** - Agent usage rules

---

## Summary

### Quick Reference

| Goal | Tool | Command |
|------|------|---------|
| **Generate text (auto)** | `text_generate` | `provider=auto` |
| **Check backends** | Resource | `stdio://models` |
| **Use Ollama** | `ollama` | `action=generate` |
| **Use Apple FM / FM chain** | `text_generate` | `provider=fm` |
| **Get recommendation** | `recommend` | `action=model` |
| **Execute task with AI** | `task_execute` | `task_id=T-123` |

### Key Takeaways

1. **Use Ollama** for production local inference
2. **Use provider=auto** for automatic backend selection
3. **Check availability** (`stdio://models`) before relying on a backend
4. **OpenCode integration** works via MCP alongside OpenCode’s own chat providers
5. **`provider=fm`** is the usual path for the FM chain on supported macOS builds
6. **Use Ollama** when you want a broad local model catalog without CGO

---

**For questions or issues**, see:
- GitHub Issues: (your repo)
- Troubleshooting section above
- `docs/EXARP_ABILITIES_AUDIT.md` for tool details
