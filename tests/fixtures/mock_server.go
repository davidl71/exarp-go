package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/davidl71/exarp-go/internal/framework"
)

// MockServer is a mock implementation of MCPServer for testing.
type MockServer struct {
	name           string
	tools          map[string]*mockTool
	prompts        map[string]*mockPrompt
	resources      map[string]*mockResource
	mu             sync.RWMutex
	registerErrors []error
}

type mockTool struct {
	Name        string
	Description string
	Schema      framework.ToolSchema
	Handler     framework.ToolHandler
}

type mockPrompt struct {
	Name        string
	Description string
	Handler     framework.PromptHandler
}

type mockResource struct {
	URI         string
	Name        string
	Description string
	MimeType    string
	Handler     framework.ResourceHandler
}

// NewMockServer creates a new mock server.
func NewMockServer(name string) *MockServer {
	return &MockServer{
		name:      name,
		tools:     make(map[string]*mockTool),
		prompts:   make(map[string]*mockPrompt),
		resources: make(map[string]*mockResource),
	}
}

// RegisterTool registers a tool with the mock server.
func (m *MockServer) RegisterTool(name, description string, schema framework.ToolSchema, handler framework.ToolHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}

	if _, exists := m.tools[name]; exists {
		return fmt.Errorf("tool %q already registered", name)
	}

	m.tools[name] = &mockTool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     handler,
	}

	return nil
}

// RegisterPrompt registers a prompt with the mock server.
func (m *MockServer) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return fmt.Errorf("prompt name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("prompt handler cannot be nil")
	}

	if _, exists := m.prompts[name]; exists {
		return fmt.Errorf("prompt %q already registered", name)
	}

	m.prompts[name] = &mockPrompt{
		Name:        name,
		Description: description,
		Handler:     handler,
	}

	return nil
}

// RegisterResource registers a resource with the mock server.
func (m *MockServer) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if uri == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("resource handler cannot be nil")
	}

	if _, exists := m.resources[uri]; exists {
		return fmt.Errorf("resource %q already registered", uri)
	}

	m.resources[uri] = &mockResource{
		URI:         uri,
		Name:        name,
		Description: description,
		MimeType:    mimeType,
		Handler:     handler,
	}

	return nil
}

// Run starts the server with the given transport (no-op for mock).
func (m *MockServer) Run(ctx context.Context, transport framework.Transport) error {
	// Mock server doesn't actually run
	<-ctx.Done()
	return ctx.Err()
}

// GetName returns the server name.
func (m *MockServer) GetName() string {
	return m.name
}

// GetTool returns a registered tool.
func (m *MockServer) GetTool(name string) (*mockTool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tool, exists := m.tools[name]

	return tool, exists
}

// GetPrompt returns a registered prompt.
func (m *MockServer) GetPrompt(name string) (*mockPrompt, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prompt, exists := m.prompts[name]

	return prompt, exists
}

// GetResource returns a registered resource.
func (m *MockServer) GetResource(uri string) (*mockResource, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	resource, exists := m.resources[uri]

	return resource, exists
}

// ToolCount returns the number of registered tools.
func (m *MockServer) ToolCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.tools)
}

// PromptCount returns the number of registered prompts.
func (m *MockServer) PromptCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.prompts)
}

// ResourceCount returns the number of registered resources.
func (m *MockServer) ResourceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.resources)
}

// CallTool calls a tool handler (for testing).
func (m *MockServer) CallTool(ctx context.Context, name string, args json.RawMessage) ([]framework.TextContent, error) {
	tool, exists := m.GetTool(name)
	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}

	return tool.Handler(ctx, args)
}

// GetPromptText calls a prompt handler (for testing).
func (m *MockServer) GetPromptText(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	prompt, exists := m.GetPrompt(name)
	if !exists {
		return "", fmt.Errorf("prompt %q not found", name)
	}

	return prompt.Handler(ctx, args)
}

// ReadResource calls a resource handler (for testing).
func (m *MockServer) ReadResource(ctx context.Context, uri string) ([]byte, string, error) {
	resource, exists := m.GetResource(uri)
	if !exists {
		return nil, "", fmt.Errorf("resource %q not found", uri)
	}

	return resource.Handler(ctx, uri)
}

// ListTools returns all registered tools (for CLI mode).
func (m *MockServer) ListTools() []framework.ToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]framework.ToolInfo, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, framework.ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			Schema:      tool.Schema,
		})
	}

	return tools
}
