package gosdk

import (
	"context"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerSessionRootsHandler wraps an MCP ServerSession to implement framework.RootsHandler
type ServerSessionRootsHandler struct {
	session *mcp.ServerSession
}

// NewServerSessionRootsHandler creates a RootsHandler that wraps a ServerSession
func NewServerSessionRootsHandler(session *mcp.ServerSession) framework.RootsHandler {
	return &ServerSessionRootsHandler{session: session}
}

// GetRoots implements framework.RootsHandler by calling ServerSession.ListRoots
func (h *ServerSessionRootsHandler) GetRoots() []framework.Root {
	ctx := context.Background()
	result, err := h.session.ListRoots(ctx, nil)
	if err != nil {
		return nil
	}

	roots := make([]framework.Root, len(result.Roots))
	for i, r := range result.Roots {
		roots[i] = framework.Root{
			URI:  r.URI,
			Name: r.Name,
		}
	}
	return roots
}
