# exarp-go ACP (Agent Client Protocol) Mode

exarp-go can run as an **ACP agent** for editors that support the [Agent Client Protocol](https://agentclientprotocol.com/), such as [Zed](https://zed.dev), [JetBrains](https://www.jetbrains.com/), and other ACP-compatible clients.

## Running in ACP Mode

```bash
# Run exarp-go as an ACP agent on stdio
./bin/exarp-go -acp
```

When started with `-acp`, exarp-go speaks the ACP protocol on stdin/stdout instead of MCP. ACP clients (e.g. Zed) spawn exarp-go as a subprocess and communicate via JSON-RPC 2.0.

## Editor Configuration

### Zed

Add exarp-go to Zed's agent configuration (e.g. in `~/.config/zed/settings.json` or project settings):

```json
{
  "agent_servers": {
    "exarp-go": {
      "command": "/absolute/path/to/exarp-go/bin/exarp-go",
      "args": ["-acp"]
    }
  }
}
```

### JetBrains

Configure the ACP agent in your IDE's agent settings to use the exarp-go binary with `-acp`.

## Capabilities

- **initialize** – Protocol negotiation
- **session/new** – Create new conversation session
- **session/prompt** – Process user prompts using exarp-go's `text_generate` tool (Ollama, Apple FM, MLX, etc.)
- **session/cancel** – Cancel in-flight prompts

Responses are streamed to the client as agent message chunks.

## Requirements

- LLM backend: Ollama, Apple Foundation Models, or MLX (for `text_generate`)
- Same environment as normal exarp-go (Todo2, config, etc.)

## Relationship to MCP

- **MCP mode** (default): exarp-go is an MCP *server* – clients (e.g. Cursor, OpenCode) call exarp-go tools.
- **ACP mode** (`-acp`): exarp-go is an ACP *agent* – clients send prompts and exarp-go responds using its LLM and tools.

ACP mode uses exarp-go's MCP tool infrastructure internally (e.g. `text_generate`) but exposes the ACP protocol to the editor.
