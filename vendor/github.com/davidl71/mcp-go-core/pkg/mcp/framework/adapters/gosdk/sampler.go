package gosdk

import (
	"context"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerSessionSampler wraps an MCP ServerSession to implement framework.Sampler
type ServerSessionSampler struct {
	session *mcp.ServerSession
}

// NewServerSessionSampler creates a Sampler that wraps a ServerSession
func NewServerSessionSampler(session *mcp.ServerSession) framework.Sampler {
	return &ServerSessionSampler{session: session}
}

// CreateMessage implements framework.Sampler by calling ServerSession.CreateMessageWithTools
func (s *ServerSessionSampler) CreateMessage(ctx context.Context, params framework.CreateMessageParams) (framework.CreateMessageResult, error) {
	// Convert framework params to MCP params
	messages := make([]*mcp.SamplingMessageV2, len(params.Messages))
	for i, msg := range params.Messages {
		messages[i] = &mcp.SamplingMessageV2{
			Role:    msg.Role,
			Content: []mcp.Content{&mcp.TextContent{Text: msg.Content}},
		}
	}

	mcpParams := &mcp.CreateMessageWithToolsParams{
		Messages: messages,
	}

	if params.Temperature > 0 {
		// Temperature is passed via options in MCP
		_ = mcpParams // TODO: add to params when SDK supports it
	}
	if params.MaxTokens > 0 {
		_ = mcpParams // TODO: add to params when SDK supports it
	}

	result, err := s.session.CreateMessageWithTools(ctx, mcpParams)
	if err != nil {
		return framework.CreateMessageResult{}, err
	}

	// Extract text content from result
	content := ""
	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(*mcp.TextContent); ok {
			content = tc.Text
		}
	}

	return framework.CreateMessageResult{
		Content:    content,
		Model:      result.Model,
		StopReason: result.StopReason,
	}, nil
}
