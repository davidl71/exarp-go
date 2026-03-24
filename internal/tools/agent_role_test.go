package tools

import "testing"

func TestAgentRoleFromTask(t *testing.T) {
	task := &Todo2Task{
		Metadata: map[string]interface{}{
			"agent_role": " Planner ",
		},
	}

	if got := AgentRoleFromTask(task); got != AgentRolePlanner {
		t.Fatalf("AgentRoleFromTask() = %q, want %q", got, AgentRolePlanner)
	}
}

func TestAgentRoleFromTask_FallsBackToTags(t *testing.T) {
	task := &Todo2Task{
		Tags: []string{"code", "review"},
	}

	if got := AgentRoleFromTask(task); got != AgentRoleWorker {
		t.Fatalf("AgentRoleFromTask() fallback = %q, want %q", got, AgentRoleWorker)
	}
}

func TestSetAgentRoleAndDominantRole(t *testing.T) {
	tasks := []Todo2Task{
		{ID: "T-1"},
		{ID: "T-2"},
		{ID: "T-3"},
	}

	if !SetAgentRole(&tasks[0], AgentRolePlanner) {
		t.Fatal("SetAgentRole() returned false for valid role")
	}
	if !SetAgentRole(&tasks[1], AgentRolePlanner) {
		t.Fatal("SetAgentRole() returned false for valid role")
	}
	if !SetAgentRole(&tasks[2], AgentRoleReviewer) {
		t.Fatal("SetAgentRole() returned false for valid role")
	}

	if got := dominantAgentRole(tasks); got != AgentRolePlanner {
		t.Fatalf("dominantAgentRole() = %q, want %q", got, AgentRolePlanner)
	}
}
