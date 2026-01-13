package gosdk

import (
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TextContentToMCP converts framework TextContent to MCP Content
func TextContentToMCP(contents []types.TextContent) []mcp.Content {
	mcpContents := make([]mcp.Content, len(contents))
	for i, content := range contents {
		mcpContents[i] = &mcp.TextContent{
			Text: content.Text,
		}
	}
	return mcpContents
}

// ToolSchemaToMCP converts framework ToolSchema to MCP input schema
func ToolSchemaToMCP(schema types.ToolSchema) map[string]interface{} {
	inputSchema := map[string]interface{}{
		"type":       schema.Type,
		"properties": schema.Properties,
	}
	if len(schema.Required) > 0 {
		inputSchema["required"] = schema.Required
	}
	return inputSchema
}
