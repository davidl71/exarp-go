# Go SDK v1.4.0 Compatibility Notes

This project now builds against `github.com/modelcontextprotocol/go-sdk v1.4.0`
and `github.com/davidl71/mcp-go-core v0.4.0`.

## Implemented

### Advertised MCP extension metadata

The server now explicitly advertises an extension capability:

- `davidl71/exarp-go`

Current extension settings:

- `projectRootContext: true`
- `resourceTemplates: true`
- `toolFiltering: true`

These flags describe behavior that is specific to `exarp-go` and useful for
MCP-aware clients:

- tool execution often depends on `PROJECT_ROOT` context
- resource template registration is supported when the backing adapter exposes it
- visible tools may be filtered by workflow mode / session context

## Documented but not wired yet

### Streamable HTTP localhost protection

`go-sdk v1.4.0` enables DNS rebinding protection by default for streamable HTTP
servers via `StreamableHTTPOptions.DisableLocalhostProtection`.

`exarp-go` does **not** currently expose an MCP streamable HTTP server. The
existing `-serve` mode is a REST/PWA API layered on top of tool calls, not an
MCP streamable transport. Because of that, there is no active
`StreamableHTTPOptions` integration to configure yet.

If a streamable HTTP MCP endpoint is added later:

- keep localhost protection enabled by default
- add an explicit configuration knob only for deployments that understand the risk
- prefer a named config field over relying on the temporary `MCPGODEBUG`
  compatibility switch from the SDK

### Sampling with tools

`go-sdk v1.4.0` adds `CreateMessageWithTools` and
`CreateMessageWithToolsHandler`.

`exarp-go` currently acts primarily as an MCP server with tool, prompt, and
resource handlers. It does not yet have a server-side feature that issues
sampling requests back to the client, so there is no current call site to
migrate.

If this is added in the future:

- prefer `CreateMessageWithTools` over the older single-content sampling path
- treat tool-enabled sampling as an explicit feature, not an implicit default
- add contract tests for multi-block content and tool-use responses

### Client-side OAuth

`go-sdk v1.4.0` introduces experimental client OAuth support behind the
`mcp_go_client_oauth` build tag.

`exarp-go` does not currently embed a Go SDK MCP client that needs this flow,
so no integration was added. This should remain deferred until there is a real
client-side transport or remote MCP client feature that requires OAuth.
