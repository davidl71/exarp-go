package taskworkflowspec

import "testing"

func containsAllFlags(specs []FieldSpec, flags []string) bool {
	set := map[string]bool{}
	for _, spec := range specs {
		if spec.CLIFlag != "" {
			set[spec.CLIFlag] = true
		}
	}
	for _, flag := range flags {
		if !set[flag] {
			return false
		}
	}
	return true
}

func TestCreateFieldSpecsExposeExpectedCLIFlags(t *testing.T) {
	expected := []string{
		"description",
		"priority",
		"tags",
		"dependencies",
		"local-ai-backend",
		"recommended-tools",
		"planning-doc",
		"epic-id",
		"parent-id",
	}
	if !containsAllFlags(CreateFieldSpecs, expected) {
		t.Fatalf("create field spec set is missing one or more expected CLI flags")
	}
}

func TestUpdateFieldSpecsExposeExpectedCLIFlags(t *testing.T) {
	expected := []string{
		"new-status",
		"new-priority",
		"tags",
		"remove-tags",
		"name",
		"description",
		"dependencies",
		"recommended-tools",
		"local-ai-backend",
		"parent-id",
	}
	if !containsAllFlags(UpdateFieldSpecs, expected) {
		t.Fatalf("update field spec set is missing one or more expected CLI flags")
	}
}

func TestFieldSpecsHaveUniqueMCPNamesWithinOperation(t *testing.T) {
	for name, specs := range map[string][]FieldSpec{
		"create": CreateFieldSpecs,
		"update": UpdateFieldSpecs,
	} {
		seen := map[string]bool{}
		for _, spec := range specs {
			if seen[spec.MCPName] {
				t.Fatalf("%s field specs contain duplicate MCP field %q", name, spec.MCPName)
			}
			seen[spec.MCPName] = true
		}
	}
}

func TestAppendTaskFieldSchemaPropertiesIncludesSharedFields(t *testing.T) {
	props := AppendTaskFieldSchemaProperties(map[string]interface{}{
		"action": map[string]interface{}{"type": "string"},
	})

	for _, key := range []string{
		"action",
		"new_status",
		"priority",
		"dependencies",
		"parent_id",
		"planning_doc",
		"epic_id",
		"recommended_tools",
		"local_ai_backend",
	} {
		if _, ok := props[key]; !ok {
			t.Fatalf("expected schema properties to include %q", key)
		}
	}
}

func TestTaskCreateInputToToolArgs(t *testing.T) {
	input := TaskCreateInput{
		Name:            "Task A",
		LongDescription: OptionalString{Set: true, Value: "details"},
		Priority:        OptionalString{Set: true, Value: "high"},
		Tags:            OptionalList{Set: true, Values: []string{"docs", "mcp"}},
		Dependencies:    OptionalList{Set: true, Values: []string{"T-1", "T-2"}},
		PlanningDoc:     OptionalString{Set: true, Value: "docs/plan.md"},
		EpicID:          OptionalString{Set: true, Value: "T-100"},
		ParentID:        OptionalString{Set: true, Value: "T-50"},
	}

	args := input.ToToolArgs()

	if got := args["action"]; got != "create" {
		t.Fatalf("action = %v, want create", got)
	}
	if got := args["dependencies"]; got != "T-1,T-2" {
		t.Fatalf("dependencies = %v, want T-1,T-2", got)
	}
	if got := args["parent_id"]; got != "T-50" {
		t.Fatalf("parent_id = %v, want T-50", got)
	}
	if got := args["epic_id"]; got != "T-100" {
		t.Fatalf("epic_id = %v, want T-100", got)
	}
}

func TestTaskUpdateInputToToolArgs(t *testing.T) {
	input := TaskUpdateInput{
		TaskIDs:         []string{"T-1", "T-2"},
		NewStatus:       OptionalString{Set: true, Value: "Done"},
		Priority:        OptionalString{Set: true, Value: "low"},
		Tags:            OptionalList{Set: true, Values: []string{"docs"}},
		RemoveTags:      OptionalList{Set: true, Values: []string{"old"}},
		Dependencies:    OptionalList{Set: true, Values: []string{"T-9"}},
		LocalAIBackend:  OptionalString{Set: true, Value: "ollama"},
		ParentID:        OptionalString{Set: true, Value: "T-10"},
		Name:            OptionalString{Set: true, Value: "New title"},
		LongDescription: OptionalString{Set: true, Value: "New description"},
	}

	args := input.ToToolArgs()

	if got := args["task_ids"]; got != "T-1,T-2" {
		t.Fatalf("task_ids = %v, want T-1,T-2", got)
	}
	if got := args["new_status"]; got != "Done" {
		t.Fatalf("new_status = %v, want Done", got)
	}
	if got := args["dependencies"]; got != "T-9" {
		t.Fatalf("dependencies = %v, want T-9", got)
	}
	if got := args["parent_id"]; got != "T-10" {
		t.Fatalf("parent_id = %v, want T-10", got)
	}
}
