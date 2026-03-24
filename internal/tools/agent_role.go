// agent_role.go — helper functions for agent_role metadata used in task routing and inference.
package tools

import (
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

const metadataAgentRoleKey = "agent_role"

const (
	AgentRolePlanner    = "planner"
	AgentRoleWorker     = "worker"
	AgentRoleReviewer   = "reviewer"
	AgentRoleResearcher = "researcher"
)

var agentRoleSet = map[string]struct{}{
	AgentRolePlanner:    {},
	AgentRoleWorker:     {},
	AgentRoleReviewer:   {},
	AgentRoleResearcher: {},
}

func normalizeAgentRole(input string) (string, bool) {
	role := strings.TrimSpace(strings.ToLower(input))
	if _, ok := agentRoleSet[role]; ok {
		return role, true
	}

	return "", false
}

func AgentRoleFromTask(task *models.Todo2Task) string {
	if task == nil {
		return ""
	}

	if task.Metadata != nil {
		if raw, ok := task.Metadata[metadataAgentRoleKey]; ok {
			if role, ok := raw.(string); ok {
				if normalized, valid := normalizeAgentRole(role); valid {
					return normalized
				}
			}
		}
	}

	for _, tag := range task.Tags {
		if role := agentRoleFromTag(tag); role != "" {
			return role
		}
	}

	return ""
}

func agentRoleFromTag(tag string) string {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "planning", "planner", "strategy":
		return AgentRolePlanner
	case "code", "implementation", "dev", "engineer":
		return AgentRoleWorker
	case "review", "audit", "qa", "investigate":
		return AgentRoleReviewer
	case "research", "analysis", "exploration":
		return AgentRoleResearcher
	default:
		return ""
	}
}

func SetAgentRole(task *models.Todo2Task, role string) bool {
	if task == nil {
		return false
	}

	if normalized, ok := normalizeAgentRole(role); ok {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata[metadataAgentRoleKey] = normalized
		return true
	}

	return false
}

func dominantAgentRole(tasks []models.Todo2Task) string {
	if len(tasks) == 0 {
		return ""
	}

	counts := make(map[string]int)
	for _, task := range tasks {
		if role := AgentRoleFromTask(&task); role != "" {
			counts[role]++
		}
	}

	var winner string
	max := 0
	for role, c := range counts {
		if c > max {
			max = c
			winner = role
		}
	}

	return winner
}
