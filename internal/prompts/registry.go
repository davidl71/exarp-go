package prompts

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// RegisterAllPrompts registers all prompts with the server
func RegisterAllPrompts(server framework.MCPServer) error {
	// Register 17 prompts (8 original + 7 high-value workflow prompts + 2 mcp-generic-tools prompts)
	prompts := []struct {
		name        string
		description string
	}{
		// Original prompts (8)
		{"align", "Analyze Todo2 task alignment with project goals."},
		{"discover", "Discover tasks from TODO comments, markdown, and orphaned tasks."},
		{"config", "Generate IDE configuration files."},
		{"scan", "Scan project dependencies for security vulnerabilities. Supports all languages via tool parameter."},
		{"scorecard", "Generate comprehensive project health scorecard with all metrics."},
		{"overview", "Generate one-page project overview for stakeholders."},
		{"dashboard", "Display comprehensive project dashboard with key metrics and status overview."},
		{"remember", "Use AI session memory to persist insights."},
		// High-value workflow prompts (7)
		{"daily_checkin", "Daily check-in workflow: server status, blockers, git health."},
		{"sprint_start", "Sprint start workflow: clean backlog, align tasks, queue work."},
		{"sprint_end", "Sprint end workflow: test coverage, docs, security check."},
		{"pre_sprint", "Pre-sprint cleanup workflow: duplicates, alignment, documentation."},
		{"post_impl", "Post-implementation review workflow: docs, security, automation."},
		{"sync", "Synchronize tasks between shared TODO table and Todo2."},
		{"dups", "Find and consolidate duplicate Todo2 tasks."},
		// mcp-generic-tools prompts
		{"context", "Manage LLM context with summarization and budget tools."},
		{"mode", "Suggest optimal Cursor IDE mode (Agent vs Ask) for a task."},
	}

	for _, p := range prompts {
		if err := server.RegisterPrompt(
			p.name,
			p.description,
			createPromptHandler(p.name),
		); err != nil {
			return fmt.Errorf("failed to register prompt %s: %w", p.name, err)
		}
	}

	return nil
}

// createPromptHandler creates a prompt handler that uses Go templates
func createPromptHandler(promptName string) framework.PromptHandler {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Retrieve prompt template from Go map
		promptText, err := GetPromptTemplate(promptName)
		if err != nil {
			return "", fmt.Errorf("failed to get prompt %s: %w", promptName, err)
		}

		// Apply template substitution if args provided
		if len(args) > 0 {
			promptText = substituteTemplate(promptText, args)
		}

		return promptText, nil
	}
}
