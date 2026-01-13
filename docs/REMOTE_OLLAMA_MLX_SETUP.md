# Remote Ollama and MLX MCP Setup

## Overview

This guide explains how to use Ollama and MLX tools via the exarp-go MCP server on a remote macOS machine.

## Prerequisites

### Ollama Installation
- ✅ **Installed:** `/opt/homebrew/bin/ollama`
- ✅ **Version:** 0.13.5
- ✅ **Status:** Available on remote machine

### MLX Installation
- ✅ **Installed:** Python MLX package
- ✅ **Platform:** Apple Silicon M4
- ✅ **Status:** Available and working

## Configuration

### MCP Server Setup

The `exarp-go` MCP server includes built-in `ollama` and `mlx` tools. These are accessed through the exarp-go server, not as separate MCP servers.

**Current Configuration** (in `.cursor/mcp.json`):
```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "PATH": "/opt/homebrew/bin:...",
        "OLLAMA_HOST": "http://localhost:11434"
      },
      "description": "Go-based MCP server - 24 tools including Ollama and MLX integration"
    }
  }
}
```

### Key Environment Variables

- **PATH**: Includes `/opt/homebrew/bin` to find `ollama` command
- **OLLAMA_HOST**: Set to `http://localhost:11434` (default Ollama API endpoint)

## Using Ollama Tools

### Available Actions

The `ollama` tool supports these actions:
- `status` - Check Ollama server status
- `models` - List available Ollama models
- `generate` - Generate text using a model
- `pull` - Download a model
- `hardware` - Get hardware information
- `docs` - Generate documentation
- `quality` - Quality analysis
- `summary` - Summarize content

### Example Usage in Cursor

**Check Ollama Status:**
```
@exarp-go ollama status
```

**List Available Models:**
```
@exarp-go ollama models
```

**Generate Text:**
```
@exarp-go ollama generate "Review this code: [code]" model="codellama"
```

**Check CodeML/CodeLlama Models:**
```
@exarp-go ollama models
```
Then look for models like:
- `codellama`
- `codellama:7b`
- `codellama:13b`
- `codellama:34b`

## Using MLX Tools

### Available Actions

The `mlx` tool supports these actions:
- `status` - Check MLX availability and hardware
- `hardware` - Get detailed hardware information
- `models` - List available MLX models
- `generate` - Generate/analyze code using MLX

### Example Usage in Cursor

**Check MLX Status:**
```
@exarp-go mlx status
```

**Check Hardware:**
```
@exarp-go mlx hardware
```

**List MLX Models:**
```
@exarp-go mlx models
```

**Generate/Analyze Code:**
```
@exarp-go mlx generate "Review this Go code: [code]" model="mlx-community/Phi-3.5-mini-instruct-4bit"
```

### Available MLX Models

Common MLX models for code:
- `mlx-community/Phi-3.5-mini-instruct-4bit` - Fast, good for general tasks
- `mlx-community/CodeLlama-7b-mlx` - Code-focused, 7B parameters
- `mlx-community/CodeLlama-13b-mlx` - Code-focused, 13B parameters
- `mlx-community/CodeLlama-34b-mlx` - Code-focused, 34B parameters (larger, more capable)
- `mlx-community/Mistral-7B-Instruct-v0.2` - General purpose
- `mlx-community/Qwen2.5-7B-Instruct` - Good reasoning

## CodeML (CodeLlama) Setup

### Via Ollama

If CodeLlama is available via Ollama:

1. **Pull CodeLlama model:**
   ```
   @exarp-go ollama pull codellama
   ```
   Or via terminal:
   ```bash
   /opt/homebrew/bin/ollama pull codellama
   ```

2. **Use CodeLlama:**
   ```
   @exarp-go ollama generate "Review this code: [code]" model="codellama"
   ```

### Via MLX

If CodeLlama is available via MLX:

1. **Check available models:**
   ```
   @exarp-go mlx models
   ```

2. **Use CodeLlama MLX:**
   ```
   @exarp-go mlx generate "Review this code: [code]" model="mlx-community/CodeLlama-7b-mlx"
   ```

## Troubleshooting

### Ollama Not Found

**Issue:** `ollama: command not found`

**Solution:**
1. Verify Ollama is installed:
   ```bash
   /opt/homebrew/bin/ollama --version
   ```

2. Ensure PATH includes `/opt/homebrew/bin` in MCP config:
   ```json
   "env": {
     "PATH": "/opt/homebrew/bin:..."
   }
   ```

3. Restart Cursor after updating config

### Ollama Server Not Running

**Issue:** Cannot connect to Ollama

**Solution:**
1. Start Ollama server:
   ```bash
   /opt/homebrew/bin/ollama serve
   ```

2. Or run in background:
   ```bash
   brew services start ollama
   ```

3. Verify it's running:
   ```bash
   curl http://localhost:11434/api/tags
   ```

### MLX Import Error

**Issue:** `ModuleNotFoundError: No module named 'mlx'`

**Solution:**
```bash
pip3 install mlx mlx-lm
```

### No CodeLlama Models Found

**Pull CodeLlama via Ollama:**
```bash
/opt/homebrew/bin/ollama pull codellama
```

**Or use MLX CodeLlama:**
The MLX models are automatically downloaded when first used. No manual pull needed.

## Best Practices

1. **Ollama for General Purpose:**
   - Use Ollama for general text generation
   - Easier model management
   - Good for interactive use

2. **MLX for Code Analysis:**
   - Use MLX for code review and analysis
   - Faster on Apple Silicon
   - Better Metal GPU utilization

3. **Model Selection:**
   - **Small tasks:** Phi-3.5-mini or CodeLlama-7b
   - **Medium tasks:** CodeLlama-13b
   - **Large/complex tasks:** CodeLlama-34b

4. **Resource Management:**
   - M4 with 16GB RAM: Can run 7B-13B models comfortably
   - 34B models may require more RAM or slower performance

## Quick Reference

### Check Status
```bash
# Ollama
@exarp-go ollama status

# MLX  
@exarp-go mlx status
```

### List Models
```bash
# Ollama models
@exarp-go ollama models

# MLX models
@exarp-go mlx models
```

### Code Review
```bash
# Via Ollama CodeLlama
@exarp-go ollama generate "Review this Go code: [code]" model="codellama"

# Via MLX CodeLlama
@exarp-go mlx generate "Review this Go code: [code]" model="mlx-community/CodeLlama-7b-mlx"
```

## Additional Resources

- [Ollama Documentation](https://ollama.ai/docs)
- [MLX Documentation](https://ml-explore.github.io/mlx/)
- [CodeLlama Models](https://huggingface.co/codellama)

