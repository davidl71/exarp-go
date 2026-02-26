// Package framework provides the MCP server interface abstraction and type re-exports.
package framework

import (
	"context"
	"encoding/json"

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

// ToolError re-export for backward compatibility.
type ToolError = framework.ToolError

// Re-export ToolError helper functions from mcp-go-core.
var (
	WrapToolError      = framework.WrapToolError
	ParseError         = framework.ParseError
	ActionError        = framework.ActionError
	UnknownActionError = framework.UnknownActionError
	ValidationError    = framework.ValidationError
	FormatErrors       = framework.FormatErrors
)

// Eliciter re-export from mcp-go-core for backward compatibility.
type Eliciter = framework.Eliciter

// Re-export Eliciter context helpers from mcp-go-core.
var (
	EliciterFromContext = framework.EliciterFromContext
	ContextWithEliciter = framework.ContextWithEliciter
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
