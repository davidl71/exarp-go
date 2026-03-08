// Package tools: agent_runner — General agent abstraction for running agents in other contexts.
// Provides a pluggable interface for different agent runtimes: Cursor CLI, Cursor Cloud Agent, MCP agents, etc.
package tools

import (
	"context"
	"fmt"
	"time"
)

type AgentType string

const (
	AgentTypeCursorCLI   AgentType = "cursor_cli"
	AgentTypeCursorCloud AgentType = "cursor_cloud"
	AgentTypeMCP         AgentType = "mcp"
	AgentTypeSubAgent    AgentType = "subagent"
)

type AgentRunnerOptions struct {
	ProjectRoot string
	Prompt      string
	Model       string
	Mode        string
	Timeout     time.Duration
}

type AgentResult struct {
	Output   string
	ExitCode int
	Error    error
}

type AgentRunner interface {
	Type() AgentType
	Run(ctx context.Context, opts AgentRunnerOptions) AgentResult
	Available() bool
}

type agentRunnerRegistry struct {
	runners map[AgentType]AgentRunner
}

var registry = &agentRunnerRegistry{
	runners: make(map[AgentType]AgentRunner),
}

func RegisterAgentRunner(r AgentRunner) {
	registry.runners[r.Type()] = r
}

func GetAgentRunner(t AgentType) AgentRunner {
	return registry.runners[t]
}

func RunAgent(ctx context.Context, agentType AgentType, opts AgentRunnerOptions) AgentResult {
	runner := GetAgentRunner(agentType)
	if runner == nil {
		return AgentResult{Error: fmt.Errorf("unknown agent type: %s", agentType)}
	}
	if !runner.Available() {
		return AgentResult{Error: fmt.Errorf("agent type %s not available", agentType)}
	}
	return runner.Run(ctx, opts)
}

func GetAvailableRunners() []AgentType {
	var types []AgentType
	for t := range registry.runners {
		types = append(types, t)
	}
	return types
}
