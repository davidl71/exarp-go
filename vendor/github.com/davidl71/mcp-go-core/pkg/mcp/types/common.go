package types

// TextContent represents MCP text content
// This is the standard format for tool responses in the MCP protocol
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolSchema represents tool input schema definition
// Used to define the structure and validation rules for tool parameters
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// ToolInfo represents tool metadata
// Contains all information about a registered tool
type ToolInfo struct {
	Name        string
	Description string
	Schema      ToolSchema
}
