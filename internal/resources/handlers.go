package resources

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// RegisterAllResources registers all resources with the server
func RegisterAllResources(server framework.MCPServer) error {
	// Register 6 resources (T-45)

	// stdio://scorecard
	if err := server.RegisterResource(
		"stdio://scorecard",
		"Project Scorecard",
		"Get current project scorecard with all health metrics.",
		"application/json",
		handleScorecard,
	); err != nil {
		return fmt.Errorf("failed to register scorecard resource: %w", err)
	}

	// stdio://memories
	if err := server.RegisterResource(
		"stdio://memories",
		"All Memories",
		"Get all AI session memories - browsable context for session continuity.",
		"application/json",
		handleMemories,
	); err != nil {
		return fmt.Errorf("failed to register memories resource: %w", err)
	}

	// stdio://memories/category/{category}
	if err := server.RegisterResource(
		"stdio://memories/category/{category}",
		"Memories by Category",
		"Get memories filtered by category (debug, research, architecture, preference, insight).",
		"application/json",
		handleMemoriesByCategory,
	); err != nil {
		return fmt.Errorf("failed to register memories by category resource: %w", err)
	}

	// stdio://memories/task/{task_id}
	if err := server.RegisterResource(
		"stdio://memories/task/{task_id}",
		"Memories for Task",
		"Get memories linked to a specific task.",
		"application/json",
		handleMemoriesByTask,
	); err != nil {
		return fmt.Errorf("failed to register memories by task resource: %w", err)
	}

	// stdio://memories/recent
	if err := server.RegisterResource(
		"stdio://memories/recent",
		"Recent Memories",
		"Get memories from the last 24 hours.",
		"application/json",
		handleRecentMemories,
	); err != nil {
		return fmt.Errorf("failed to register recent memories resource: %w", err)
	}

	// stdio://memories/session/{date}
	if err := server.RegisterResource(
		"stdio://memories/session/{date}",
		"Session Memories",
		"Get memories from a specific session date (YYYY-MM-DD format).",
		"application/json",
		handleSessionMemories,
	); err != nil {
		return fmt.Errorf("failed to register session memories resource: %w", err)
	}

	// stdio://prompts
	if err := server.RegisterResource(
		"stdio://prompts",
		"All Prompts",
		"Get all prompts in compact format for discovery.",
		"application/json",
		handleAllPrompts,
	); err != nil {
		return fmt.Errorf("failed to register all prompts resource: %w", err)
	}

	// stdio://prompts/mode/{mode}
	if err := server.RegisterResource(
		"stdio://prompts/mode/{mode}",
		"Prompts by Mode",
		"Get prompts for a workflow mode (daily_checkin, security_review, task_management, etc.).",
		"application/json",
		handlePromptsByMode,
	); err != nil {
		return fmt.Errorf("failed to register prompts by mode resource: %w", err)
	}

	// stdio://prompts/persona/{persona}
	if err := server.RegisterResource(
		"stdio://prompts/persona/{persona}",
		"Prompts by Persona",
		"Get prompts for a persona (developer, pm, qa, reviewer, etc.).",
		"application/json",
		handlePromptsByPersona,
	); err != nil {
		return fmt.Errorf("failed to register prompts by persona resource: %w", err)
	}

	// stdio://prompts/category/{category}
	if err := server.RegisterResource(
		"stdio://prompts/category/{category}",
		"Prompts by Category",
		"Get prompts in a category (workflow, persona, analysis, etc.).",
		"application/json",
		handlePromptsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register prompts by category resource: %w", err)
	}

	// stdio://session/mode
	if err := server.RegisterResource(
		"stdio://session/mode",
		"Session Mode",
		"Get current inferred session mode (AGENT/ASK/MANUAL) with confidence.",
		"application/json",
		handleSessionMode,
	); err != nil {
		return fmt.Errorf("failed to register session mode resource: %w", err)
	}

	return nil
}

// Resource handlers
// Note: handleScorecard and memory handlers are implemented in scorecard.go and memories.go
// They are in the same package, so they're automatically available here

func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handlePromptsByMode(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handlePromptsByPersona(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handlePromptsByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleSessionMode(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}
