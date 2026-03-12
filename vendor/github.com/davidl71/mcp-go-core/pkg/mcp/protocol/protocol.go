// Package protocol provides JSON-RPC 2.0 protocol types and helpers for MCP servers.
//
// This package includes request/response structures, error types, and helper functions
// for creating JSON-RPC messages. It provides a complete implementation of the
// JSON-RPC 2.0 specification as used by the MCP protocol.
//
// Example:
//
//	req := protocol.JSONRPCRequest{
//		JSONRPC: "2.0",
//		ID:      "123",
//		Method:  "tools/list",
//	}
//	resp := protocol.NewSuccessResponse("123", result)
package protocol

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 message structures

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"` // No omitempty - null is valid for parse errors
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON-RPC 2.0 error codes
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
)

// InitializeParams represents the initialize request parameters
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the initialize response
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

// ToolsCapability indicates tools support
type ToolsCapability struct{}

// ResourcesCapability indicates resources support
type ResourcesCapability struct{}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource definition
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ToolCallParams represents parameters for a tool call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ResourceReadParams represents parameters for reading a resource
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// Helper functions for creating responses

// NewErrorResponse creates a JSON-RPC error response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewSuccessResponse creates a JSON-RPC success response
func NewSuccessResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(id interface{}, method string) *JSONRPCResponse {
	return NewErrorResponse(id, ErrCodeMethodNotFound, fmt.Sprintf("Method not found: %s", method), nil)
}

// NewInvalidParamsError creates an invalid params error
func NewInvalidParamsError(id interface{}, message string) *JSONRPCResponse {
	return NewErrorResponse(id, ErrCodeInvalidParams, message, nil)
}

// NewInternalError creates an internal error
func NewInternalError(id interface{}, message string) *JSONRPCResponse {
	return NewErrorResponse(id, ErrCodeInternalError, message, nil)
}

// FormatRequestID converts a JSON-RPC request ID to a string for logging.
// Handles string, number (float64 from JSON), int, int64, and nil.
func FormatRequestID(id interface{}) string {
	if id == nil {
		return "null"
	}
	switch v := id.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", id)
	}
}
