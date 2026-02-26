// registry.go — Registers all MCP prompts with the server framework.
//
// Package prompts provides MCP prompt registration and template definitions.
package prompts

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// allPrompts is the single source of truth for all registered MCP prompts.
// Both RegisterAllPrompts and ListAllPromptNames derive from this list.
var allPrompts = []struct {
	name        string
	description string
}{
	// Original prompts (9)
	{"align", "Analyze Todo2 task alignment with project goals."},
	{"discover", "Discover tasks from TODO comments, markdown, and orphaned tasks."},
	{"config", "Generate IDE configuration files."},
	{"scan", "Scan project dependencies for security vulnerabilities. Supports all languages via tool parameter."},
	{"scorecard", "Generate comprehensive project health scorecard with all metrics."},
	{"overview", "Generate one-page project overview for stakeholders."},
	{"plan", "Generate Cursor-style .plan.md (Purpose, Technical Foundation, Milestones, Open Questions)."},
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
	// Task management prompts
	{"task_update", "Update Todo2 task status using proper MCP tools - never edit JSON directly."},
	// Migrated from Python (16): docs, automation, workflow, advisor, personas
	{"docs", "Analyze documentation health and optionally create Todo2 tasks for issues."},
	{"automation_discover", "Discover new automation opportunities in the codebase."},
	{"weekly_maintenance", "Weekly maintenance workflow: docs, duplicates, security, sync."},
	{"task_review", "Comprehensive task review workflow for backlog hygiene."},
	{"project_health", "Comprehensive project health assessment (server, docs, testing, security, cicd, alignment)."},
	{"automation_setup", "One-time automation setup: setup_hooks(git), setup_hooks(patterns), cron."},
	{"advisor_consult", "Consult a trusted advisor by metric, stage, or tool."},
	{"advisor_briefing", "Get a morning briefing from trusted advisors based on project health."},
	{"persona_developer", "Developer daily workflow for writing quality code."},
	{"persona_project_manager", "Project Manager workflow for delivery tracking."},
	{"persona_code_reviewer", "Code Reviewer workflow for quality gates."},
	{"persona_executive", "Executive/Stakeholder workflow for strategic view."},
	{"persona_security", "Security Engineer workflow for risk management."},
	{"persona_architect", "Architect workflow for system design."},
	{"persona_qa", "QA Engineer workflow for quality assurance."},
	{"persona_tech_writer", "Technical Writer workflow for documentation."},
	{"tractatus_decompose", "Decompose a concept using Tractatus Thinking MCP; instructs AI to call tractatus_thinking (start, add, export) and produce structured output."},
}

// ListAllPromptNames returns the names of all registered prompts in registration order.
// Use this instead of maintaining a separate hardcoded list elsewhere.
func ListAllPromptNames() []string {
	names := make([]string, len(allPrompts))
	for i, p := range allPrompts {
		names[i] = p.name
	}
	return names
}

// RegisterAllPrompts registers all prompts with the server.
func RegisterAllPrompts(server framework.MCPServer) error {
	for _, p := range allPrompts {
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

// createPromptHandler creates a prompt handler that uses Go templates.
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
