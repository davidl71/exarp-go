// Package framework provides the MCP server interface abstraction and type re-exports.
package framework

import (
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

// JsonRawMessage is an alias for json.RawMessage to avoid import conflicts.
type JsonRawMessage = json.RawMessage
