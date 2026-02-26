# Provider/Transform Architecture Design for exarp-go v2

> Research document — no code changes required.

## Background

FastMCP 3.0 introduced a **Provider + Transform** composability model that separates data sources from data transformations. This document explores how exarp-go could adopt similar patterns for its v2 architecture.

## Current Architecture (Flat Registry)

```
cmd/server/main.go
  → factory.NewServer()
  → tools.RegisterAllTools(server)
    → registerCoreTools(server)     # task_workflow, session, report, health
    → registerAITools(server)       # memory, estimation, LLM backends
    → registerInfraTools(server)    # automation, git, lint, testing, security
    → registerMiscTools(server)     # alignment, attribution, catalog, context
  → resources.RegisterAllResources(server)
  → prompts.RegisterAllPrompts(server)
```

**Strengths**: Simple, direct, easy to understand.
**Weaknesses**: All tools must be registered upfront; no runtime composition; no per-session tool sets; monolithic handler dispatch.

## FastMCP 3.0 Provider/Transform Model

### Provider Interface

A **Provider** is a source of MCP components (tools, resources, prompts):

```python
class Provider(ABC):
    def list_tools() -> list[Tool]
    def list_resources() -> list[Resource]
    def call_tool(name, args) -> Result
    def read_resource(uri) -> Content
```

Built-in providers:
- **OpenAPIProvider** — Auto-generates tools from OpenAPI specs
- **FileSystemProvider** — Exposes files as resources
- **ProxyProvider** — Forwards to another MCP server
- **StaticProvider** — In-memory tool/resource definitions

### Transform Interface

A **Transform** wraps a Provider and modifies its output:

```python
class Transform(Provider):
    def __init__(self, provider: Provider)
    # Overrides list_tools/call_tool/etc. to filter, augment, or transform
```

Built-in transforms:
- **ResourcesAsTools** — Converts resources into tool definitions
- **PromptsAsTools** — Converts prompts into tool definitions
- **VisibilityFilter** — Per-session tool/resource visibility
- **AuthMiddleware** — Per-component authorization

### Composition

```python
server = FastMCP()
server.mount("/api", OpenAPIProvider("api.yaml"))
server.mount("/files", FileSystemProvider("./data"))
server.add_transform(ResourcesAsTools())
server.add_transform(VisibilityFilter(filter_fn))
```

## Proposed exarp-go v2 Architecture

### Provider Interface (Go)

```go
type Provider interface {
    ListTools(ctx context.Context) []ToolInfo
    ListResources(ctx context.Context) []ResourceInfo
    CallTool(ctx context.Context, name string, args json.RawMessage) ([]TextContent, error)
    ReadResource(ctx context.Context, uri string) ([]byte, string, error)
}
```

### Concrete Providers

| Provider | Purpose | Migration from |
|---|---|---|
| `CoreProvider` | task_workflow, session, report, health | `registerCoreTools` |
| `AIProvider` | memory, estimation, LLM backends | `registerAITools` |
| `InfraProvider` | automation, git, lint, testing | `registerInfraTools` |
| `MiscProvider` | alignment, attribution, catalog | `registerMiscTools` |
| `ResourceProvider` | All stdio:// resources | `RegisterAllResources` |

### Transform Interface (Go)

```go
type Transform interface {
    Wrap(provider Provider) Provider
}
```

### Built-in Transforms

| Transform | Purpose | Current equivalent |
|---|---|---|
| `RecoveryTransform` | Panic recovery | `toolRecoveryMiddleware` |
| `LoggingTransform` | Tool call logging | `toolLoggingMiddleware` |
| `CacheTransform` | Request-scoped caching | `toolContextCacheMiddleware` |
| `HooksTransform` | Before/after callbacks | `toolHooksMiddleware` |
| `FilterTransform` | Per-session tool filtering | `filteredServer` |
| `ResourcesAsToolsTransform` | Resource → tool conversion | `resources_as_tools.go` |
| `SingleflightTransform` | Dedup concurrent calls | `scorecardFlight` |

### Composition Example

```go
server := mcp.NewServer("exarp-go", "1.0.0")

core := NewCoreProvider()
ai := NewAIProvider()
infra := NewInfraProvider()

pipeline := compose.New(
    compose.Recovery(),
    compose.Logging(logger),
    compose.Cache(),
    compose.Hooks(hooks),
    compose.Filter(filterFn),
)

server.Mount("/core", pipeline.Wrap(core))
server.Mount("/ai", pipeline.Wrap(ai))
server.Mount("/infra", pipeline.Wrap(infra))
```

## Migration Path

### Phase 1: Current (v1) — Implemented

- Flat registry with `RegisterAllTools`
- Middleware chain in `factory/server.go`
- `resources_as_tools.go` as standalone feature
- `filteredServer` wrapper for tool filtering
- Hooks as middleware

### Phase 2: Extract Providers

- Group existing handlers into Provider implementations
- Each provider owns its handler map
- Registry delegates to providers instead of flat switch

### Phase 3: Formalize Transforms

- Convert existing middleware to Transform interface
- Add composability (wrap providers with transforms)
- Enable runtime provider mounting

### Phase 4: Dynamic Composition

- Per-session provider selection
- Hot-reload providers
- Plugin system for external providers

## Key Design Decisions

### 1. Provider Granularity

**Option A**: One provider per registry group (4 providers) — simple, matches current code.
**Option B**: One provider per tool — maximum flexibility but high overhead.
**Recommendation**: Option A for v2, with support for sub-providers within groups.

### 2. Transform Ordering

Transforms must compose in a defined order. The current middleware chain already handles this:
1. Recovery (outermost — catches panics from all inner layers)
2. Cache (per-request memoization)
3. Logging (audit trail)
4. Hooks (user callbacks)
5. Handler (innermost)

### 3. Backward Compatibility

The `framework.MCPServer` interface must remain stable. Providers and transforms are internal implementation details. The existing `RegisterAllTools` function becomes a compatibility shim that creates providers internally.

## References

- [FastMCP 3.0 launch](https://www.jlowin.dev/blog/fastmcp-3-launch)
- [mcp-go (mark3labs)](https://github.com/mark3labs/mcp-go) — Most advanced Go MCP SDK
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) — Official Go SDK
- `docs/research/FASTMCP3_GO_LANDSCAPE.md` — Go SDK landscape analysis
- `internal/factory/server.go` — Current middleware chain
- `internal/tools/resources_as_tools.go` — ResourcesAsTools implementation
