package database

import (
	"fmt"
	"testing"
)

func TestGetAgentID(t *testing.T) {
	agentID, err := GetAgentID()
	if err != nil {
		t.Fatalf("GetAgentID() error = %v", err)
	}

	fmt.Printf("Full Agent ID: %s\n", agentID)

	// Should not be empty
	if agentID == "" {
		t.Error("Expected non-empty agent ID, got empty string")
	}

	// Should contain hostname (at minimum)
	// Format should be: {agent-type}-{hostname}-{pid}
	if len(agentID) < 5 {
		t.Errorf("Agent ID too short: %s", agentID)
	}
}

func TestGetAgentIDSimple(t *testing.T) {
	agentID, err := GetAgentIDSimple()
	if err != nil {
		t.Fatalf("GetAgentIDSimple() error = %v", err)
	}

	fmt.Printf("Simple Agent ID: %s\n", agentID)

	// Should not be empty
	if agentID == "" {
		t.Error("Expected non-empty agent ID, got empty string")
	}
}

func TestParsePIDFromAgentID(t *testing.T) {
	tests := []struct {
		agentID string
		wantPID int
		wantOK  bool
	}{
		{"general-Davids-Mac-mini-12345", 12345, true},
		{"backend-agent-host-999", 999, true},
		{"a-b-1", 1, true},
		{"no-pid", 0, false},
		{"single", 0, false},
		{"", 0, false},
		{"two-parts-42", 42, true},
	}
	for _, tt := range tests {
		pid, ok := ParsePIDFromAgentID(tt.agentID)
		if ok != tt.wantOK || pid != tt.wantPID {
			t.Errorf("ParsePIDFromAgentID(%q) = (%d, %v), want (%d, %v)", tt.agentID, pid, ok, tt.wantPID, tt.wantOK)
		}
	}
}

func TestAgentProcessExists(t *testing.T) {
	// Current process as agent ID should exist
	agentID, err := GetAgentID()
	if err != nil {
		t.Fatalf("GetAgentID() error = %v", err)
	}

	if !AgentProcessExists(agentID) {
		t.Errorf("AgentProcessExists(current agent ID) = false, want true")
	}
	// Simple ID has no PID, so cannot verify
	simpleID, _ := GetAgentIDSimple()
	if AgentProcessExists(simpleID) {
		t.Errorf("AgentProcessExists(simple ID without PID) = true, want false")
	}
	// Bogus ID
	if AgentProcessExists("no-pid") {
		t.Error("AgentProcessExists(no-pid) = true, want false")
	}
}
