// Package framework provides framework-agnostic abstractions for MCP servers.
// This file adds MCP Sampling support: server can request LLM generation from the client.

package framework

import "context"

// Sampler represents a client that can perform LLM sampling.
// When set in context (e.g. by a client that supports sampling), tools can request
// the client to generate text using its LLM. SamplerFromContext returns nil if no
// sampler is configured, allowing tools to degrade gracefully.
type Sampler interface {
	// CreateMessage requests the client to generate a message using its LLM.
	// Returns the generated content and any error.
	CreateMessage(ctx context.Context, params CreateMessageParams) (CreateMessageResult, error)
}

// CreateMessageParams holds parameters for sampling request.
// Maps to MCP sampling/createMessage params.
type CreateMessageParams struct {
	Messages        []SamplingMessage `json:"messages"`
	ModelPreference string            `json:"modelPreference,omitempty"`
	SystemPrompt    string            `json:"systemPrompt,omitempty"`
	IncludeContext  string            `json:"includeContext,omitempty"`
	Temperature     float64           `json:"temperature,omitempty"`
	MaxTokens       int               `json:"maxTokens,omitempty"`
	StopSequences   []string          `json:"stopSequences,omitempty"`
}

// SamplingMessage represents a message in the sampling context.
type SamplingMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// CreateMessageResult holds the result from sampling.
type CreateMessageResult struct {
	Content    string `json:"content"`
	Model      string `json:"model,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
}

type samplerKey struct{}

// SamplerFromContext returns the Sampler from context, or nil if not set.
func SamplerFromContext(ctx context.Context) Sampler {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(samplerKey{}); v != nil {
		if s, ok := v.(Sampler); ok {
			return s
		}
	}
	return nil
}

// ContextWithSampler returns a new context with the Sampler set.
func ContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return context.WithValue(ctx, samplerKey{}, s)
}
