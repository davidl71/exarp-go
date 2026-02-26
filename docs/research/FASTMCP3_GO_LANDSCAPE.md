# FastMCP 3.0 Patterns in Go — Landscape Analysis

**Date:** 2026-02-26
**Purpose:** Assess whether any Go MCP library implements FastMCP 3.0's Provider/Transform architecture, and identify actionable patterns for exarp-go.

---

## Answer: No Go Library Implements FastMCP 3.0

No Go MCP SDK implements FastMCP 3.0's core architectural innovation — the **Provider + Transform** composability model. The Go ecosystem is 6–12 months behind Python's FastMCP in framework-level features.

However, **mark3labs/mcp-go** (8.2k stars) implements several individual FastMCP 3.0 features natively. The official Go SDK (3.9k stars) is spec-focused with minimal framework features.

---

## Feature Matrix: FastMCP 3.0 vs Go SDKs vs exarp-go

### Architecture

| FastMCP 3.0 Feature | mcp-go | Official Go SDK | exarp-go | Gap |
|---------------------|--------|----------------|----------|-----|
| **Provider abstraction** (where components come from) | ❌ | ❌ | ❌ | No Go implementation exists |
| **Transform pipeline** (middleware for components) | ❌ | ❌ | ❌ | No Go implementation exists |
| **Composable mounting** (server-in-server) | ❌ | ❌ | ❌ | No Go implementation exists |
| **OpenAPI Provider** (REST → MCP tools) | ❌ | ❌ | ❌ | No Go implementation exists |
| **FileSystem Provider** (auto-discover from dir) | ❌ | ❌ | ❌ | No Go implementation exists |
| **Proxy Provider** (remote MCP → local) | ❌ | ❌ | ❌ | No Go implementation exists |

### Session & Visibility

| FastMCP 3.0 Feature | mcp-go | Official Go SDK | exarp-go | Gap |
|---------------------|--------|----------------|----------|-----|
| **Per-session tools** | ✅ `AddSessionTool()` | ❌ | ❌ | mcp-go has it |
| **Tool filtering** (per-request) | ✅ `WithToolFilter()` | ❌ | Partial (`workflow_mode`) | mcp-go has it |
| **Visibility system** (enable/disable by tag) | ❌ | ❌ | ❌ | No Go implementation |
| **Session state** (`set_state`/`get_state`) | ❌ | ❌ | Partial (session prime) | No Go equivalent |
| **Component versioning** (`version="2.0"`) | ❌ | ❌ | ❌ | No Go implementation |
| **ResourcesAsTools / PromptsAsTools** | ❌ | ❌ | ❌ | No Go implementation |

### Middleware & Auth

| FastMCP 3.0 Feature | mcp-go | Official Go SDK | exarp-go | Gap |
|---------------------|--------|----------------|----------|-----|
| **Tool handler middleware** | ✅ `WithToolHandlerMiddleware()` | ❌ | ✅ `WrapHandler` | Both have it |
| **Request hooks** (before/after) | ✅ `WithHooks()` | ❌ | ❌ | mcp-go has it |
| **Recovery middleware** | ✅ `WithRecovery()` | ❌ | ❌ | mcp-go has it |
| **Per-component auth** | ❌ | ✅ `auth` package | ✅ access control config | Partial |
| **Auth middleware** (server-wide) | ❌ | ✅ OAuth | ✅ rate limiting | Different approaches |

### Production & Operations

| FastMCP 3.0 Feature | mcp-go | Official Go SDK | exarp-go | Gap |
|---------------------|--------|----------------|----------|-----|
| **Background tasks** (async execution) | ✅ `AddTaskTool()` | ❌ | ✅ Redis+Asynq queue | Both have it |
| **Tool timeouts** | ❌ | ❌ | ✅ config timeouts | exarp-go has it |
| **Pagination** | ✅ `WithPaginationLimit()` | ❌ | ❌ | mcp-go has it |
| **OpenTelemetry tracing** | ❌ | ❌ | ❌ | No Go implementation |
| **Hot reload** | ❌ | ❌ | ✅ `make dev-watch` | exarp-go has it |
| **Elicitation** | ✅ | ❌ | ✅ | Both have it |

### Transports

| Transport | mcp-go | Official Go SDK | exarp-go |
|-----------|--------|----------------|----------|
| **Stdio** | ✅ | ✅ | ✅ |
| **Streamable HTTP** | ✅ | ❌ | ❌ |
| **SSE** | ✅ | ❌ | ❌ |
| **In-process** | ✅ | ❌ | ❌ |
| **Command** | ❌ | ✅ | ❌ |

### CLI & Developer Experience

| FastMCP 3.0 Feature | mcp-go | Official Go SDK | exarp-go | Gap |
|---------------------|--------|----------------|----------|-----|
| **CLI tool invocation** (`fastmcp call`) | ❌ | ❌ | ✅ `exarp-go -tool X` | exarp-go has it |
| **CLI discovery** (`fastmcp discover`) | ❌ | ❌ | ❌ | No Go implementation |
| **CLI generation** (`generate-cli`) | ❌ | ❌ | ❌ | No Go implementation |
| **Typed input structs** (auto schema) | Options pattern | ✅ `jsonschema` tags | JSON schema in registry | Official SDK best |

---

## Top Patterns to Adopt (Prioritized)

### Priority 1: Low-Effort, High-Impact (from mcp-go)

#### 1.1 Per-Session Tool Visibility

mcp-go's `WithToolFilter` + `SessionWithTools` pattern is directly applicable. exarp-go's `workflow_mode` already filters by mode — extending to per-session filtering would enable progressive capability reveal.

```go
// mcp-go pattern
server.WithToolFilter(func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
    session := server.ClientSessionFromContext(ctx)
    // Filter tools based on session context
})
```

**exarp-go implementation:** Add a `ToolFilter` to `internal/framework/server.go` that checks session context for enabled/disabled tool sets.

#### 1.2 Singleflight for Cache Deduplication

From `golang/groupcache` research. Prevents duplicate scorecard/report computations when multiple tool calls trigger simultaneously.

```go
import "golang.org/x/sync/singleflight"
var scorecardGroup singleflight.Group
result, err, _ := scorecardGroup.Do(cacheKey, func() (interface{}, error) {
    return GenerateGoScorecard(ctx, projectRoot, opts)
})
```

#### 1.3 Request Hooks (from mcp-go)

mcp-go's `WithHooks` provides before/after callbacks for all request types. Would enable OpenTelemetry-style tracing and the JSONL audit trail without modifying individual handlers.

### Priority 2: Medium-Effort, High-Impact

#### 2.1 ResourcesAsTools Transform

FastMCP 3.0's `ResourcesAsTools` auto-generates `list_resources` and `read_resource` tools. exarp-go has 24 resources that tool-only clients can't access. A Go implementation would be ~100 lines.

#### 2.2 Component Versioning

FastMCP 3.0 lets you register `@tool(version="2.0")` alongside v1. In Go, this could be metadata on the tool registration. Useful as exarp-go evolves tool schemas.

#### 2.3 Recovery Middleware

mcp-go's `WithRecovery()` catches panics in tool handlers. exarp-go handlers could panic on unexpected input — a recovery wrapper would return a clean MCP error instead of crashing the server.

### Priority 3: High-Effort, Architectural

#### 3.1 Provider/Transform Architecture

The architectural heart of FastMCP 3.0. No Go library implements this. exarp-go could be the first Go MCP framework to implement composable Providers and Transforms.

**Providers:** Where tools come from (local registry, remote MCP, OpenAPI, filesystem).
**Transforms:** How tools are shaped (namespace, filter, version, auth).

This would replace exarp-go's current flat registry with a composable pipeline. Major refactoring — best as a v2 initiative.

#### 3.2 Streamable HTTP Transport

mcp-go supports this; exarp-go doesn't. Required for remote/shared server deployments. Would enable multi-agent scenarios without Redis.

---

## Go SDK Recommendation for exarp-go

exarp-go should **not** migrate to any external SDK. The custom `internal/framework` abstraction gives full control over the handler pipeline. Instead:

1. **Cherry-pick patterns** from mcp-go (tool filtering, hooks, recovery)
2. **Adopt singleflight** from groupcache for cache dedup
3. **Build ResourcesAsTools** as an exarp-go native feature
4. **Design Provider/Transform** as a future architecture (v2)
5. **Keep monitoring** the official Go SDK for spec-level features

---

## Caching Research Summary

| Library | Stars | Best Use | exarp-go Fit |
|---------|-------|----------|-------------|
| `golang/groupcache` | 13.3k | `singleflight` dedup | Adopt `singleflight` only |
| `hashicorp/golang-lru` | ~4k | Size-bounded LRU | Best upgrade for TTLCache |
| `eko/gocache` | 2.8k | Chain/loadable/metrics | Future distributed caching |
| `patrickmn/go-cache` | 8.8k | Simple TTL map | Similar to current, unmaintained |

---

## References

- FastMCP 3.0 GA: https://www.jlowin.dev/blog/fastmcp-3-launch
- FastMCP 3.0 Features: https://www.jlowin.dev/blog/fastmcp-3-whats-new
- mcp-go: https://github.com/mark3labs/mcp-go (8.2k stars)
- Official Go SDK: https://github.com/modelcontextprotocol/go-sdk (3.9k stars)
- mcp-golang: https://github.com/metoro-io/mcp-golang (1.2k stars)
- SetiabudiResearch FastMCP Go: https://github.com/SetiabudiResearch/mcp-go-sdk (7 stars, thin wrapper)
- groupcache singleflight: https://github.com/golang/groupcache
- Go caching article: https://dev.to/leapcell/caching-in-go-doing-it-right-25i5
- MCP server in Go tutorial: https://en.bioerrorlog.work/entry/hello-mcp-golang
- MCP server Go guide (2026): https://fast.io/resources/mcp-server-golang/
- eko/gocache: https://github.com/eko/gocache
- patrickmn/go-cache: https://github.com/patrickmn/go-cache
