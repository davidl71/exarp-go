// Package framework provides framework-agnostic abstractions for MCP servers.
//
// This file defines FilteredServer, which wraps an MCPServer and applies
// a ToolFilterFunc to ListTools() so callers see a filtered tool list per context.

package framework

import (
	"context"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// FilteredServer wraps an MCPServer and applies a filter to ListTools().
// All other methods delegate to the inner server.
type FilteredServer struct {
	MCPServer
	Filter ToolFilterFunc
}

// ListTools returns the inner server's tools filtered by Filter.
// If Filter is nil, returns the inner server's full list.
func (f *FilteredServer) ListTools() []types.ToolInfo {
	list := f.MCPServer.ListTools()
	if f.Filter == nil {
		return list
	}
	return f.Filter(context.Background(), list)
}

// Ensure FilteredServer implements MCPServer (delegate methods are on embedded MCPServer).
var _ MCPServer = (*FilteredServer)(nil)

// NewFilteredServer wraps server with a tool filter. If filter is nil, ListTools
// returns the full list (same as server.ListTools()).
func NewFilteredServer(server MCPServer, filter ToolFilterFunc) *FilteredServer {
	return &FilteredServer{MCPServer: server, Filter: filter}
}
