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
