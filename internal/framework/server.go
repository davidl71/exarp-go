// Package framework provides the MCP server interface abstraction and local compatibility shims.
package framework

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// Re-export types and interfaces from mcp-go-core for backward compatibility.
type (
	MCPServer       = framework.MCPServer
	ToolHandler     = framework.ToolHandler
	PromptHandler   = framework.PromptHandler
	ResourceHandler = framework.ResourceHandler
	Transport       = framework.Transport
	TextContent     = types.TextContent
	ToolSchema      = types.ToolSchema
	ToolInfo        = types.ToolInfo
)

// ResourceTemplateRegistrar is an optional extension used by servers that support
// registering parameterized resource templates for client discovery.
type ResourceTemplateRegistrar interface {
	RegisterResourceTemplate(uriTemplate, name, description, mimeType string, handler ResourceHandler) error
}

// ServerExtensionReporter is implemented by servers that can report the
// advertised MCP capability extensions configured at startup.
type ServerExtensionReporter interface {
	ServerExtensions() map[string]any
}

// ResourceUpdateNotifier allows servers to trigger resource update notifications.
type ResourceUpdateNotifier interface {
	NotifyResourceUpdated(context.Context, *mcp.ResourceUpdatedNotificationParams) error
}

// Eliciter re-export from mcp-go-core for backward compatibility.
type Eliciter = framework.Eliciter

// Re-export Eliciter context helpers from mcp-go-core.
var (
	EliciterFromContext = framework.EliciterFromContext
	ContextWithEliciter = framework.ContextWithEliciter
)

// Sampler re-export from mcp-go-core for sampling support.
// Sampler allows the server to request LLM generation from the client.
type (
	Sampler             = framework.Sampler
	CreateMessageParams = framework.CreateMessageParams
	SamplingMessage     = framework.SamplingMessage
	CreateMessageResult = framework.CreateMessageResult
)

// Re-export Sampler context helpers from mcp-go-core.
var (
	SamplerFromContext = framework.SamplerFromContext
	ContextWithSampler = framework.ContextWithSampler
)

// Root re-export from mcp-go-core for Roots support.
// Root represents a client workspace boundary.
type Root = framework.Root

// Re-export Roots context helpers from mcp-go-core.
var (
	RootsFromContext = framework.RootsFromContext
	ContextWithRoots = framework.ContextWithRoots
)

// JsonRawMessage is an alias for json.RawMessage to avoid import conflicts.
type JsonRawMessage = json.RawMessage

// ToolHookFunc is called before or after a tool invocation for cross-cutting concerns
// (logging, metrics, audit trail). The name parameter is the tool name.
type ToolHookFunc func(ctx context.Context, name string, args json.RawMessage)

// Hooks provides before/after callbacks for the tool handler pipeline.
type Hooks struct {
	BeforeToolCall ToolHookFunc
	AfterToolCall  ToolHookFunc
}

// ToolFilterFunc filters the set of visible tools per request context.
// Return a subset of tools to restrict visibility for the current session/mode.
type ToolFilterFunc func(ctx context.Context, tools []ToolInfo) []ToolInfo
