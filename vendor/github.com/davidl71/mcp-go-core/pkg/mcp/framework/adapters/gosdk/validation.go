package gosdk

import (
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateToolRegistration validates tool registration parameters and returns typed errors.
func ValidateToolRegistration(name, description string, handler interface{}) error {
	if name == "" {
		return &framework.ErrInvalidTool{ToolName: "", Reason: "name cannot be empty"}
	}
	if description == "" {
		return &framework.ErrInvalidTool{ToolName: name, Reason: "description cannot be empty"}
	}
	if handler == nil {
		return &framework.ErrInvalidTool{ToolName: name, Reason: "handler cannot be nil"}
	}
	return nil
}

// ValidatePromptRegistration validates prompt registration parameters and returns typed errors.
func ValidatePromptRegistration(name, description string, handler interface{}) error {
	if name == "" {
		return &framework.ErrInvalidPrompt{PromptName: "", Reason: "name cannot be empty"}
	}
	if description == "" {
		return &framework.ErrInvalidPrompt{PromptName: name, Reason: "description cannot be empty"}
	}
	if handler == nil {
		return &framework.ErrInvalidPrompt{PromptName: name, Reason: "handler cannot be nil"}
	}
	return nil
}

// ValidateResourceRegistration validates resource-specific registration parameters and returns typed errors.
func ValidateResourceRegistration(uri, name, description string, handler interface{}) error {
	if name == "" {
		return &framework.ErrInvalidResource{URI: uri, Reason: "name cannot be empty"}
	}
	if description == "" {
		return &framework.ErrInvalidResource{URI: uri, Reason: "description cannot be empty"}
	}
	if handler == nil {
		return &framework.ErrInvalidResource{URI: uri, Reason: "handler cannot be nil"}
	}
	if uri == "" {
		return &framework.ErrInvalidResource{URI: "", Reason: "resource URI cannot be empty"}
	}
	return nil
}

// ValidateCallToolRequest validates a CallToolRequest and returns typed errors.
func ValidateCallToolRequest(req *mcp.CallToolRequest) error {
	if req == nil {
		return &framework.ErrInvalidRequest{RequestType: "CallTool", Reason: "request cannot be nil"}
	}
	if req.Params == nil {
		return &framework.ErrInvalidRequest{RequestType: "CallTool", Reason: "params cannot be nil"}
	}
	return nil
}

// ValidateGetPromptRequest validates a GetPromptRequest and returns typed errors.
func ValidateGetPromptRequest(req *mcp.GetPromptRequest) error {
	if req == nil {
		return &framework.ErrInvalidRequest{RequestType: "GetPrompt", Reason: "request cannot be nil"}
	}
	if req.Params == nil {
		return &framework.ErrInvalidRequest{RequestType: "GetPrompt", Reason: "params cannot be nil"}
	}
	return nil
}

// ValidateReadResourceRequest validates a ReadResourceRequest and returns typed errors.
func ValidateReadResourceRequest(req *mcp.ReadResourceRequest) error {
	if req == nil {
		return &framework.ErrInvalidRequest{RequestType: "ReadResource", Reason: "request cannot be nil"}
	}
	if req.Params == nil {
		return &framework.ErrInvalidRequest{RequestType: "ReadResource", Reason: "params cannot be nil"}
	}
	if req.Params.URI == "" {
		return &framework.ErrInvalidResource{URI: "", Reason: "resource URI cannot be empty"}
	}
	return nil
}
