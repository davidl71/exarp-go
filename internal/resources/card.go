// card.go — agent://card resource: A2A/ACP agent discovery card.
//
// Returns a JSON document describing this agent's identity, transports,
// capabilities, and skills. Follows the A2A Agent Card schema so clients
// like Cursor, Claude Code, OpenCode, and Zed can auto-discover what the
// agent can do without reading the source.
package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/tools"
)

// agentCard is the JSON payload returned at agent://card.
type agentCard struct {
	SchemaVersion  string            `json:"schema_version"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Description    string            `json:"description"`
	Transports     []string          `json:"transports"`
	Capabilities   agentCapabilities `json:"capabilities"`
	Authentication agentAuth         `json:"authentication"`
	DefaultInputs  []string          `json:"default_input_modes"`
	DefaultOutputs []string          `json:"default_output_modes"`
	Skills         []agentSkill      `json:"skills"`
	ToolCount      int               `json:"tool_count"`
	Backends       map[string]any    `json:"backends"`
}

type agentCapabilities struct {
	Streaming              bool           `json:"streaming"`
	PushNotifications      bool           `json:"push_notifications"`
	StateTransitionHistory bool           `json:"state_transition_history"`
	Extensions             map[string]any `json:"extensions"`
}

type agentAuth struct {
	Schemes []string `json:"schemes"`
}

type agentSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
}

// handleAgentCard returns the agent://card discovery document.
func handleAgentCard(ctx context.Context, uri string) ([]byte, string, error) {
	cfg, _ := config.Load() // best-effort; defaults are fine if config missing
	version := "1.0.0"
	if cfg != nil && cfg.Version != "" {
		version = cfg.Version
	}

	catalog := tools.GetToolCatalog()

	// Group tools into skills by category.
	byCategory := make(map[string][]string)
	for id, entry := range catalog {
		byCategory[entry.Category] = append(byCategory[entry.Category], id)
	}
	categories := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	skills := make([]agentSkill, 0, len(categories))
	for _, cat := range categories {
		toolIDs := byCategory[cat]
		sort.Strings(toolIDs)
		skills = append(skills, agentSkill{
			ID:          slugify(cat),
			Name:        cat,
			Description: fmt.Sprintf("%s tools (%d)", cat, len(toolIDs)),
			Tools:       toolIDs,
		})
	}

	card := agentCard{
		SchemaVersion: "1.0",
		Name:          "exarp-go",
		Version:       version,
		Description:   "exarp-go MCP server — task management, local AI (FM/Ollama), project health, semantic memory, and agent automation.",
		Transports:    []string{"stdio-mcp", "acp", "mcp-http"},
		Capabilities: agentCapabilities{
			Streaming:              true,
			PushNotifications:      false,
			StateTransitionHistory: false,
			Extensions: map[string]any{
				"projectRootContext":    true,
				"resourceTemplates":     true,
				"toolFiltering":         true,
				"resourceSubscriptions": true,
				"agentRunner":           true,
				"fmPlanExecute":         true,
			},
		},
		Authentication: agentAuth{Schemes: []string{"none"}},
		DefaultInputs:  []string{"text/plain"},
		DefaultOutputs: []string{"text/plain", "application/json"},
		Skills:         skills,
		ToolCount:      len(catalog),
		Backends:       tools.LLMBackendStatus(),
	}

	data, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("marshal agent card: %w", err)
	}
	return data, "application/json", nil
}

// slugify converts a category name like "Project Health" to "project_health".
func slugify(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32)
		case c == ' ' || c == '-':
			out = append(out, '_')
		default:
			out = append(out, c)
		}
	}
	return string(out)
}
