# AI/LLM Integration Guide

**Last Updated**: 2026-03-08  
**Version**: v0.3.5  
**Scope**: Complete guide to exarp-go's AI and LLM capabilities

This guide covers all AI/LLM integration in exarp-go, including backend selection, model recommendations, MLX support, and OpenCode integration.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Supported Backends](#supported-backends)
3. [Backend Selection Strategy](#backend-selection-strategy)
4. [MLX Integration](#mlx-integration)
5. [OpenCode Integration](#opencode-integration)
6. [Tool Reference](#tool-reference)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Check Available Backends

```bash
# Via CLI
exarp-go -tool ollama -args '{"action":"status"}'
exarp-go -tool mlx -args '{"action":"status"}'
exarp-go -tool apple_foundation_models -args '{"action":"status"}'

# Via MCP resource (fastest)
# Read stdio://models for full backend availability
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

exarp-go supports **8 LLM backends** with unified interface:

| Backend | Status | Best For | Platform | Tool Name |
|---------|--------|----------|----------|-----------|
| **Ollama** | ✅ Production | Local inference, variety | All | `ollama` |
| **Apple FM** | ✅ Production | Lowest latency | macOS only | `apple_foundation_models` |
| **MLX** | ⚠️ Experimental | GPU acceleration | Apple Silicon | `mlx` |
| **llamacpp** | ⏸️ Deferred | GGUF models | All | `llamacpp` |
| **Insight** | ✅ Available | Cloud API | All | `text_generate` (provider=insight) |
| **LocalAI** | ✅ Available | OpenAI-compatible | All | `text_generate` (provider=localai) |
| **Gateway** | ✅ Available | Load balancing | All | `text_generate` (provider=gateway) |
| **Auto** | ✅ Recommended | Best available | All | `text_generate` (provider=auto) |

### Backend Priorities (Auto Mode)

When using `provider=auto`, exarp-go selects backends in this order:

1. **Apple Foundation Models** (if available and on macOS)
2. **Ollama** (if running and models available)
3. **MLX** (if available and on Apple Silicon)
4. **LocalAI** (if configured)
5. **Insight** (fallback cloud)

---

## Backend Selection Strategy

### Decision Tree

```
Are you on macOS?
├─ Yes → Do you need lowest latency?
│   ├─ Yes → Use Apple Foundation Models
│   └─ No → Do you want local models?
│       ├─ Yes → Use Ollama (recommended) or MLX
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
| **Offline work** | Ollama or MLX | No internet required |
| **Production** | Ollama + fallback to Insight | Reliability + backup |

---

## MLX Integration

### Overview

**MLX** is Apple's machine learning framework optimized for Apple Silicon (M1/M2/M3). exarp-go includes experimental MLX support for local GPU-accelerated inference.

### Current Status

- **Implementation**: Go stub + Python bridge
- **Status**: Experimental (not enabled in default builds)
- **Recommendation**: Use Ollama instead for production
- **Build**: Native Go implementation returns "not available" message

### Why Use MLX?

✅ **Pros:**
- GPU acceleration on Apple Silicon
- Lower memory usage than Ollama for some models
- Direct Metal API access

❌ **Cons:**
- More complex setup than Ollama
- Requires Python bridge
- Less mature than Ollama
- Smaller model ecosystem

### Installation (If Needed)

```bash
# Install MLX (requires Python 3.9+)
pip install mlx mlx-lm

# Test MLX
python3 -c "import mlx; print(mlx.__version__)"

# Run exarp-go with MLX
exarp-go -tool mlx -args '{"action":"status"}'
```

### Using MLX

```bash
# Check hardware support
exarp-go -tool mlx -args '{"action":"hardware"}'

# List available models
exarp-go -tool mlx -args '{"action":"models"}'

# Generate text
exarp-go -tool mlx -args '{
  "action": "generate",
  "model": "mlx-community/Llama-3.2-3B-Instruct-4bit",
  "prompt": "Explain goroutines",
  "max_tokens": 500
}'
```

### MLX vs Ollama

| Feature | MLX | Ollama |
|---------|-----|--------|
| **Setup** | Complex (Python + pip) | Simple (brew install) |
| **Performance** | GPU-accelerated | CPU/GPU hybrid |
| **Models** | MLX-converted only | Broad ecosystem |
| **Stability** | Experimental | Production-ready |
| **Documentation** | Limited | Excellent |
| **Recommendation** | Experimental use | Production use |

**Verdict**: **Use Ollama** unless you have specific needs for MLX's Metal API integration or are experimenting with MLX-specific features.

---

## OpenCode Integration

### Overview

**OpenCode** is an AI-powered code assistant that works via MCP (Model Context Protocol). exarp-go integrates seamlessly with OpenCode via the MCP server.

### Configuration

The `opencode.json` file configures exarp-go as an MCP server for OpenCode:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
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

### Key Configuration Parameters

| Parameter | Purpose | Example |
|-----------|---------|---------|
| `command` | Path to exarp-go wrapper script | `run_exarp_go.sh` |
| `PROJECT_ROOT` | Workspace root directory | `/path/to/project` |
| `EXARP_MIGRATIONS_DIR` | Todo2 migrations location | `/path/to/migrations` |
| `EXARP_WATCH` | Auto-reload on changes (0=off) | `0` or `1` |

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

### OpenCode + MLX Workflow

```bash
# 1. Configure OpenCode with exarp-go
opencode --config opencode.json

# 2. In OpenCode, use exarp-go's LLM tools
#    OpenCode sends: {"tool": "text_generate", "args": {"provider": "auto", ...}}
#    
# 3. exarp-go auto-selects best backend:
#    - Tries Apple FM (fastest on macOS)
#    - Falls back to Ollama (most reliable)
#    - Falls back to MLX (if available)
#    - Final fallback: cloud (Insight)
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
- `fm` - Apple Foundation Models
- `ollama` - Ollama
- `mlx` - MLX (experimental)
- `llamacpp` - llama.cpp (deferred)
- `insight` - Cloud API
- `localai` - LocalAI
- `gateway` - Load balancer

### Backend-Specific Tools

#### Apple Foundation Models

```bash
# Status
exarp-go -tool apple_foundation_models -args '{"action":"status"}'

# List models
exarp-go -tool apple_foundation_models -args '{"action":"models"}'

# Generate
exarp-go -tool apple_foundation_models -args '{
  "action": "generate",
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

#### MLX

```bash
# Status
exarp-go -tool mlx -args '{"action":"status"}'

# Hardware capabilities
exarp-go -tool mlx -args '{"action":"hardware"}'

# List models
exarp-go -tool mlx -args '{"action":"models"}'

# Generate
exarp-go -tool mlx -args '{
  "action": "generate",
  "model": "mlx-community/Llama-3.2-3B-Instruct-4bit",
  "prompt": "Explain channels in Go"
}'
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
- Use experimental backends (MLX) in production
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

### MLX Errors

```bash
# Check Python environment
python3 --version  # Requires 3.9+

# Check MLX installation
pip list | grep mlx

# Reinstall if needed
pip install --upgrade mlx mlx-lm
```

### Memory Issues

- **Large models**: Use quantized versions (4-bit, 8-bit)
- **MLX**: Adjust max_tokens parameter
- **Ollama**: Use smaller models (3B instead of 13B)

### Performance Issues

- **Slow generation**: Check backend status (may be downloading models)
- **High latency**: Switch to smaller model or faster backend
- **Timeout**: Increase timeout or reduce max_tokens

### Backend Selection Issues

```bash
# Debug backend availability
exarp-go -tool list_resources -args '{}'

# Check which backends are available
# Read stdio://models response for fm_available, ollama_available, mlx_available
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
- **`docs/LLAMACPP_FUTURE.md`** - llamacpp integration notes
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
| **Use Apple FM** | `apple_foundation_models` | `action=generate` |
| **Use MLX** | `mlx` | `action=generate` |
| **Get recommendation** | `recommend` | `action=model` |
| **Execute task with AI** | `task_execute` | `task_id=T-123` |

### Key Takeaways

1. **Use Ollama** for production local inference
2. **Use provider=auto** for automatic backend selection
3. **Avoid MLX** unless you need experimental features
4. **Check availability** before using specific backends
5. **OpenCode integration** works seamlessly via MCP
6. **Apple FM** is fastest for quick queries on macOS
7. **llamacpp** is deferred - use Ollama instead

---

**For questions or issues**, see:
- GitHub Issues: (your repo)
- Troubleshooting section above
- `docs/EXARP_ABILITIES_AUDIT.md` for tool details
