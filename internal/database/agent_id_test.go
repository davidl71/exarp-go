package database

import "testing"

func TestAgentIdentityWithoutPID(t *testing.T) {
	t.Parallel()

	if got := AgentIdentityWithoutPID("general-myhost-12345"); got != "general-myhost" {
		t.Fatalf("AgentIdentityWithoutPID = %q, want general-myhost", got)
	}

	if got := AgentIdentityWithoutPID("no-pid-suffix"); got != "no-pid-suffix" {
		t.Fatalf("unchanged = %q", got)
	}
}
