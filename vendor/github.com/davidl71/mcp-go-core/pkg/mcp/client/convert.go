// Package client provides type conversion utilities between external client libraries
// and mcp-go-core types.

package client

import (
	"encoding/json"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// ConvertExternalToolToToolInfo converts a tool from an external client library
// to mcp-go-core types.ToolInfo.
//
// This function handles conversion from external library tool types
// (e.g., from github.com/metoro-io/mcp-golang) to mcp-go-core types.
//
// The externalTool parameter should be a tool struct from the external library.
// Since we're using interface{} to avoid direct dependencies, we use JSON
// marshaling/unmarshaling for conversion.
func ConvertExternalToolToToolInfo(externalTool interface{}) (types.ToolInfo, error) {
	// Marshal the external tool to JSON
	jsonData, err := json.Marshal(externalTool)
	if err != nil {
		return types.ToolInfo{}, fmt.Errorf("failed to marshal external tool: %w", err)
	}

	// Unmarshal into a generic map to extract fields
	var toolMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &toolMap); err != nil {
		return types.ToolInfo{}, fmt.Errorf("failed to unmarshal tool: %w", err)
	}

	// Extract name
	name, ok := toolMap["name"].(string)
	if !ok {
		return types.ToolInfo{}, fmt.Errorf("tool missing or invalid name field")
	}

	// Extract description (may be pointer in external library)
	var description string
	if desc, ok := toolMap["description"]; ok {
		if descPtr, ok := desc.(*string); ok {
			if descPtr != nil {
				description = *descPtr
			}
		} else if descStr, ok := desc.(string); ok {
			description = descStr
		}
	}

	// Extract inputSchema
	var schema types.ToolSchema
	if inputSchema, ok := toolMap["inputSchema"]; ok {
		schemaData, err := json.Marshal(inputSchema)
		if err != nil {
			return types.ToolInfo{}, fmt.Errorf("failed to marshal input schema: %w", err)
		}

		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return types.ToolInfo{}, fmt.Errorf("failed to unmarshal input schema: %w", err)
		}
	} else {
		// Default schema if not provided
		schema = types.ToolSchema{
			Type:       "object",
			Properties: make(map[string]interface{}),
		}
	}

	return types.ToolInfo{
		Name:        name,
		Description: description,
		Schema:      schema,
	}, nil
}

// ConvertExternalTextContent converts text content from an external client library
// to mcp-go-core types.TextContent.
func ConvertExternalTextContent(externalContent interface{}) (types.TextContent, error) {
	jsonData, err := json.Marshal(externalContent)
	if err != nil {
		return types.TextContent{}, fmt.Errorf("failed to marshal external content: %w", err)
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &contentMap); err != nil {
		return types.TextContent{}, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	contentType := "text"
	if typ, ok := contentMap["type"].(string); ok {
		contentType = typ
	}

	var text string
	// Handle different possible field names
	if textField, ok := contentMap["text"].(string); ok {
		text = textField
	} else if textContent, ok := contentMap["textContent"]; ok {
		if textContentMap, ok := textContent.(map[string]interface{}); ok {
			if textStr, ok := textContentMap["text"].(string); ok {
				text = textStr
			}
		}
	}

	return types.TextContent{
		Type: contentType,
		Text: text,
	}, nil
}

// ConvertExternalTextContentSlice converts a slice of text content from an external
// client library to a slice of mcp-go-core types.TextContent.
func ConvertExternalTextContentSlice(externalContents []interface{}) ([]types.TextContent, error) {
	result := make([]types.TextContent, 0, len(externalContents))
	for _, content := range externalContents {
		converted, err := ConvertExternalTextContent(content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert text content: %w", err)
		}
		result = append(result, converted)
	}
	return result, nil
}

// ConvertClientInfoToExternal converts mcp-go-core protocol.ClientInfo to a format
// suitable for external client libraries.
func ConvertClientInfoToExternal(clientInfo protocol.ClientInfo) map[string]interface{} {
	return map[string]interface{}{
		"name":    clientInfo.Name,
		"version": clientInfo.Version,
	}
}

// ConvertInitializeParamsToExternal converts mcp-go-core protocol.InitializeParams
// to a format suitable for external client libraries.
func ConvertInitializeParamsToExternal(params protocol.InitializeParams) map[string]interface{} {
	result := map[string]interface{}{
		"protocolVersion": params.ProtocolVersion,
		"clientInfo":      ConvertClientInfoToExternal(params.ClientInfo),
	}

	if params.Capabilities.Experimental != nil && len(params.Capabilities.Experimental) > 0 {
		result["capabilities"] = map[string]interface{}{
			"experimental": params.Capabilities.Experimental,
		}
	}

	return result
}
