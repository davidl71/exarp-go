// Package framework provides framework-agnostic abstractions for MCP servers.
//
// This file defines tool hooks and tool filter types for cross-cutting concerns
// (logging, metrics, audit) and per-request tool visibility.

package framework

import (
	"context"
	"encoding/json"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

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
// The context can be used to pass session or mode information.
type ToolFilterFunc func(ctx context.Context, tools []types.ToolInfo) []types.ToolInfo
