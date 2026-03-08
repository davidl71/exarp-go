// Package framework provides the MCP server interface abstraction and type re-exports.
package framework

import (
	"encoding/json"

	core "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// Re-export types and interfaces from mcp-go-core for backward compatibility.
type (
	MCPServer       = core.MCPServer
	ToolHandler     = core.ToolHandler
	PromptHandler   = core.PromptHandler
	ResourceHandler = core.ResourceHandler
	Transport       = core.Transport
	TextContent     = types.TextContent
	ToolSchema      = types.ToolSchema
	ToolInfo        = types.ToolInfo
)

// ToolError re-export for backward compatibility.
type ToolError = core.ToolError

// Re-export error types from mcp-go-core for backward compatibility.
type (
	ParseError         = core.ParseError
	ActionError        = core.ActionError
	UnknownActionError = core.UnknownActionError
	ValidationError    = core.ValidationError
	FormatErrors       = core.FormatErrors
)

// Re-export ToolError helper function from mcp-go-core.
var WrapToolError = core.WrapToolError

// Eliciter re-export from mcp-go-core for backward compatibility.
type Eliciter = core.Eliciter

// Re-export Eliciter context helpers from mcp-go-core.
var (
	EliciterFromContext = core.EliciterFromContext
	ContextWithEliciter = core.ContextWithEliciter
)

// JsonRawMessage is an alias for json.RawMessage to avoid import conflicts.
type JsonRawMessage = json.RawMessage

// Re-export tool hooks and filter from mcp-go-core for backward compatibility.
type (
	ToolHookFunc   = core.ToolHookFunc
	Hooks          = core.Hooks
	ToolFilterFunc = core.ToolFilterFunc
)

// Re-export FilteredServer and NewFilteredServer from mcp-go-core.
type FilteredServer = core.FilteredServer

var NewFilteredServer = core.NewFilteredServer
