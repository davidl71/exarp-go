package resources

import (
	"context"
	"fmt"

	"github.com/davidl/mcp-stdio-tools/internal/bridge"
	"github.com/davidl/mcp-stdio-tools/internal/framework"
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

	return nil
}

// Resource handlers

func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleMemories(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleMemoriesByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleMemoriesByTask(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleRecentMemories(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}

func handleSessionMemories(ctx context.Context, uri string) ([]byte, string, error) {
	return bridge.ExecutePythonResource(ctx, uri)
}
