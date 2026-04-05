// agent_id.go — Agent ID generation for task lock ownership.
package database

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/utils"
)

// GetAgentID generates a unique agent identifier for task locking
// Format: {agent-type}-{hostname}-{pid}
// Example: "backend-agent-Davids-Mac-mini-12345"
//
// Detection order:
// 1. EXARP_AGENT environment variable
// 2. Cursor agent type (from session detection)
// 3. Default: "general"
//
// The hostname and PID ensure uniqueness across processes/machines.
func GetAgentID() (string, error) {
	// 1. Get agent type from environment
	agentType := os.Getenv("EXARP_AGENT")
	if agentType == "" {
		// Try to detect from Cursor context
		// For now, default to "general" - can be enhanced with session detection
		agentType = "general"
	}

	// 2. Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		// Fallback to environment variable or generic
		hostname = os.Getenv("HOSTNAME")
		if hostname == "" {
			hostname = "unknown"
		}
	}

	// 3. Get process ID for uniqueness
	pid := os.Getpid()

	// Format: agent-type-hostname-pid
	agentID := fmt.Sprintf("%s-%s-%d", agentType, hostname, pid)

	return agentID, nil
}

// AgentIdentityWithoutPID returns the agent_id with the trailing "-<pid>" removed when the ID
// matches GetAgentID format (…-<numeric_pid>). Same workstation session across different
// processes (e.g. separate CLI invocations) then compares equal for the same agent type+host.
func AgentIdentityWithoutPID(agentID string) string {
	pid, ok := ParsePIDFromAgentID(agentID)
	if !ok {
		return agentID
	}

	suffix := fmt.Sprintf("-%d", pid)

	return strings.TrimSuffix(agentID, suffix)
}

// ParsePIDFromAgentID extracts the process ID from an agent ID string.
// Agent ID format: {agent-type}-{hostname}-{pid}, e.g. "general-Davids-Mac-mini-12345".
// Returns (pid, true) if the last segment is numeric, otherwise (0, false).
func ParsePIDFromAgentID(agentID string) (pid int, ok bool) {
	if agentID == "" {
		return 0, false
	}

	parts := strings.Split(agentID, "-")
	if len(parts) < 2 {
		return 0, false
	}

	last := parts[len(parts)-1]

	p, err := strconv.Atoi(last)
	if err != nil || p <= 0 {
		return 0, false
	}

	return p, true
}

// AgentProcessExists returns true if the agent identified by agentID has a running process.
// Uses the PID embedded in the agent ID (see GetAgentID). If agentID has no PID
// (e.g. from GetAgentIDSimple), returns false (cannot verify).
func AgentProcessExists(agentID string) bool {
	pid, ok := ParsePIDFromAgentID(agentID)
	if !ok {
		return false
	}

	return utils.ProcessExists(pid)
}
