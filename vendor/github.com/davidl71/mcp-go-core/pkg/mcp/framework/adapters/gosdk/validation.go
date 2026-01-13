package gosdk

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateRegistration validates common registration parameters
// for tools, prompts, and resources.
func ValidateRegistration(name, description string, handler interface{}) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	return nil
}

// ValidateResourceRegistration validates resource-specific registration parameters
func ValidateResourceRegistration(uri, name, description string, handler interface{}) error {
	if err := ValidateRegistration(name, description, handler); err != nil {
		return err
	}
	if uri == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}
	return nil
}

// ValidateCallToolRequest validates a CallToolRequest
func ValidateCallToolRequest(req *mcp.CallToolRequest) error {
	if req == nil {
		return fmt.Errorf("call tool request cannot be nil")
	}
	if req.Params == nil {
		return fmt.Errorf("call tool request params cannot be nil")
	}
	return nil
}

// ValidateGetPromptRequest validates a GetPromptRequest
func ValidateGetPromptRequest(req *mcp.GetPromptRequest) error {
	if req == nil {
		return fmt.Errorf("get prompt request cannot be nil")
	}
	if req.Params == nil {
		return fmt.Errorf("get prompt request params cannot be nil")
	}
	return nil
}

// ValidateReadResourceRequest validates a ReadResourceRequest
func ValidateReadResourceRequest(req *mcp.ReadResourceRequest) error {
	if req == nil {
		return fmt.Errorf("read resource request cannot be nil")
	}
	if req.Params == nil {
		return fmt.Errorf("read resource request params cannot be nil")
	}
	if req.Params.URI == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}
	return nil
}
