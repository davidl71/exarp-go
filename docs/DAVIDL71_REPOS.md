# davidl71 GitHub Repositories

Reference documentation for code authored by davidl71 (David Lowes).

## Authored Repositories (Primary)

| Repository | Description | Language |
|------------|-------------|----------|
| [exarp-go](https://github.com/davidl71/exarp-go) | MCP server for project management automation - documentation health, task alignment, duplicate detection, security scanning | Go |
| [devwisdom-go](https://github.com/davidl71/devwisdom-go) | Wisdom Module Extraction - Standalone Go MCP server providing wisdom quotes, trusted advisors, and inspirational guidance for developers | Go |
| [mcp-go-core](https://github.com/davidl71/mcp-go-core) | Core MCP framework in Go - reusable components for building MCP servers | Go |
| [cursor-mcp-config](https://github.com/davidl71/cursor-mcp-config) | Cursor MCP configuration | Shell |
| [project-management-automation](https://github.com/davidl71/project-management-automation) | MCP server for project management automation tools - documentation health, task alignment, duplicate detection, security scanning | Python |

## Dependencies

The exarp-go project depends on:

```go
github.com/davidl71/devwisdom-go v0.1.2
github.com/davidl71/mcp-go-core v0.3.1
```

These are stored as local dependencies in sibling directories:
- `../devwisdom-go`
- `../mcp-go-core`

## Local Development Setup

To clone all dependencies for local development:

```bash
cd ..
git clone https://github.com/davidl71/mcp-go-core ../mcp-go-core
git clone https://github.com/davidl71/devwisdom-go ../devwisdom-go
```

## MCP Configuration

### OpenCode
```json
{
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/path/to/exarp-go/bin/exarp-go"],
      "enabled": true
    }
  }
}
```

### Cursor
```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "/path/to/exarp-go/bin/exarp-go",
      "args": [],
      "env": { "PROJECT_ROOT": "/path/to/exarp-go" }
    }
  }
}
```

## Ansible Development Dependencies

See [ansible/README.md](../ansible/README.md) for setting up development environment with:
- Go 1.24.0
- Python 3.10+ with pip and uv
- Node.js & npm
- Linters: golangci-lint, shellcheck, gomarklint, etc.
- Optional: Ollama, Redis
