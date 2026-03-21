// Package framework provides the MCP server interface abstraction and local compatibility shims.
package framework

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

type samplerContextKeyType struct{}

var samplerContextKey = &samplerContextKeyType{}

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

// ResourceTemplateRegistrar is an optional extension used by servers that support
// registering parameterized resource templates for client discovery.
type ResourceTemplateRegistrar interface {
	RegisterResourceTemplate(uriTemplate, name, description, mimeType string, handler ResourceHandler) error
}

// ServerExtensionReporter is implemented by servers that can report the
// advertised MCP capability extensions configured at startup.
type ServerExtensionReporter interface {
	ServerExtensions() map[string]any
}

// ResourceUpdateNotifier allows servers to trigger resource update notifications.
type ResourceUpdateNotifier interface {
	NotifyResourceUpdated(context.Context, *mcp.ResourceUpdatedNotificationParams) error
}

// Eliciter re-export from mcp-go-core for backward compatibility.
type Eliciter = framework.Eliciter

// Re-export Eliciter context helpers from mcp-go-core.
var (
	EliciterFromContext = framework.EliciterFromContext
	ContextWithEliciter = framework.ContextWithEliciter
)

// Sampler allows the server to request LLM generation from the client.
// Stub implementation - actual sampling requires client support.
type Sampler interface {
	CreateMessage(ctx context.Context, params CreateMessageParams) (CreateMessageResult, error)
}

// CreateMessageParams contains parameters for sampling request.
type CreateMessageParams struct {
	Messages     []SamplingMessage
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
	Preferences  MessagePreferences
}

// NewCreateMessageParams creates CreateMessageParams with required messages.
func NewCreateMessageParams(messages []SamplingMessage) CreateMessageParams {
	return CreateMessageParams{Messages: messages}
}

// AddMessage appends a message to the params.
func (p CreateMessageParams) AddMessage(msg SamplingMessage) CreateMessageParams {
	p.Messages = append(p.Messages, msg)
	return p
}

// AddUserMessage appends a user message.
func (p CreateMessageParams) AddUserMessage(content string) CreateMessageParams {
	return p.AddMessage(NewSamplingMessage("user", content))
}

// WithSystemPrompt sets the system prompt.
func (p CreateMessageParams) WithSystemPrompt(systemPrompt string) CreateMessageParams {
	p.SystemPrompt = systemPrompt
	return p
}

// WithTemperature sets the temperature.
func (p CreateMessageParams) WithTemperature(temperature float64) CreateMessageParams {
	p.Temperature = temperature
	return p
}

// WithMaxTokens sets the max tokens.
func (p CreateMessageParams) WithMaxTokens(maxTokens int) CreateMessageParams {
	p.MaxTokens = maxTokens
	return p
}

// MessagePreferences contains sampling preferences.
type MessagePreferences struct {
	Priority string
}

// SamplingMessage represents a message in a sampling conversation.
type SamplingMessage struct {
	Role    string
	Content string
}

// NewSamplingMessage creates a SamplingMessage with role and content.
func NewSamplingMessage(role, content string) SamplingMessage {
	return SamplingMessage{Role: role, Content: content}
}

// CreateMessageResult contains the result of a sampling request.
type CreateMessageResult struct {
	Content string
	TokenUsage
}

// TokenUsage contains token usage information.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// SamplerFromContext extracts a Sampler from context.
// Returns nil if not available.
func SamplerFromContext(ctx context.Context) Sampler {
	if v := ctx.Value(samplerContextKey); v != nil {
		if s, ok := v.(Sampler); ok {
			return s
		}
	}
	return nil
}

// ContextWithSampler adds a Sampler to context.
func ContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return context.WithValue(ctx, samplerContextKey, s)
}

// Root represents a client workspace boundary.
type Root struct {
	URI  string
	Name string
}

// RootsFromContext extracts Roots from context.
// Returns empty slice if not available.
func RootsFromContext(ctx context.Context) []Root {
	return nil
}

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
