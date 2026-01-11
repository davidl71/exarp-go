package resources

import (
	"testing"

	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestRegisterAllResources(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllResources(server)
	if err != nil {
		t.Fatalf("RegisterAllResources() error = %v", err)
	}

	// Verify all 21 resources are registered (11 base + 6 task resources + 1 server resource + 1 models resource + 2 tools resources)
	if server.ResourceCount() != 21 {
		t.Errorf("server.ResourceCount() = %v, want 21", server.ResourceCount())
	}

	// Verify specific resources are registered
	expectedResources := []string{
		"stdio://scorecard",
		"stdio://memories",
		"stdio://memories/category/{category}",
		"stdio://memories/task/{task_id}",
		"stdio://memories/recent",
		"stdio://memories/session/{date}",
		"stdio://prompts",
		"stdio://prompts/mode/{mode}",
		"stdio://prompts/persona/{persona}",
		"stdio://prompts/category/{category}",
		"stdio://session/mode",
		"stdio://server/status",
		"stdio://models",
		"stdio://tools",
		"stdio://tools/{category}",
		"stdio://tasks",
		"stdio://tasks/{task_id}",
		"stdio://tasks/status/{status}",
		"stdio://tasks/priority/{priority}",
		"stdio://tasks/tag/{tag}",
		"stdio://tasks/summary",
	}

	for _, uri := range expectedResources {
		resource, exists := server.GetResource(uri)
		if !exists {
			t.Errorf("resource %q not registered", uri)
			continue
		}
		if resource.URI != uri {
			t.Errorf("resource.URI = %v, want %v", resource.URI, uri)
		}
		if resource.Name == "" {
			t.Errorf("resource %q name is empty", uri)
		}
		if resource.MimeType == "" {
			t.Errorf("resource %q mimeType is empty", uri)
		}
	}
}

func TestRegisterAllResources_URIParsing(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllResources(server)
	if err != nil {
		t.Fatalf("RegisterAllResources() error = %v", err)
	}

	// Verify URI patterns are registered correctly
	tests := []struct {
		name string
		uri  string
		want bool
	}{
		{
			name: "scorecard",
			uri:  "stdio://scorecard",
			want: true,
		},
		{
			name: "memories",
			uri:  "stdio://memories",
			want: true,
		},
		{
			name: "memories category pattern",
			uri:  "stdio://memories/category/{category}",
			want: true,
		},
		{
			name: "memories task pattern",
			uri:  "stdio://memories/task/{task_id}",
			want: true,
		},
		{
			name: "memories recent",
			uri:  "stdio://memories/recent",
			want: true,
		},
		{
			name: "memories session pattern",
			uri:  "stdio://memories/session/{date}",
			want: true,
		},
		{
			name: "all prompts",
			uri:  "stdio://prompts",
			want: true,
		},
		{
			name: "prompts by mode pattern",
			uri:  "stdio://prompts/mode/{mode}",
			want: true,
		},
		{
			name: "prompts by persona pattern",
			uri:  "stdio://prompts/persona/{persona}",
			want: true,
		},
		{
			name: "prompts by category pattern",
			uri:  "stdio://prompts/category/{category}",
			want: true,
		},
		{
			name: "session mode",
			uri:  "stdio://session/mode",
			want: true,
		},
		{
			name: "server status",
			uri:  "stdio://server/status",
			want: true,
		},
		{
			name: "models",
			uri:  "stdio://models",
			want: true,
		},
		{
			name: "all tools",
			uri:  "stdio://tools",
			want: true,
		},
		{
			name: "tools by category pattern",
			uri:  "stdio://tools/{category}",
			want: true,
		},
		{
			name: "all tasks",
			uri:  "stdio://tasks",
			want: true,
		},
		{
			name: "task by ID pattern",
			uri:  "stdio://tasks/{task_id}",
			want: true,
		},
		{
			name: "tasks by status pattern",
			uri:  "stdio://tasks/status/{status}",
			want: true,
		},
		{
			name: "tasks by priority pattern",
			uri:  "stdio://tasks/priority/{priority}",
			want: true,
		},
		{
			name: "tasks by tag pattern",
			uri:  "stdio://tasks/tag/{tag}",
			want: true,
		},
		{
			name: "tasks summary",
			uri:  "stdio://tasks/summary",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := server.GetResource(tt.uri)
			if exists != tt.want {
				t.Errorf("resource %q exists = %v, want %v", tt.uri, exists, tt.want)
			}
		})
	}
}

func TestRegisterAllResources_HandlerRegistration(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllResources(server)
	if err != nil {
		t.Fatalf("RegisterAllResources() error = %v", err)
	}

	// Verify handlers are registered for each resource
	expectedResources := []string{
		"stdio://scorecard",
		"stdio://memories",
		"stdio://memories/category/{category}",
		"stdio://memories/task/{task_id}",
		"stdio://memories/recent",
		"stdio://memories/session/{date}",
		"stdio://prompts",
		"stdio://prompts/mode/{mode}",
		"stdio://prompts/persona/{persona}",
		"stdio://prompts/category/{category}",
		"stdio://session/mode",
		"stdio://server/status",
		"stdio://models",
		"stdio://tools",
		"stdio://tools/{category}",
		"stdio://tasks",
		"stdio://tasks/{task_id}",
		"stdio://tasks/status/{status}",
		"stdio://tasks/priority/{priority}",
		"stdio://tasks/tag/{tag}",
		"stdio://tasks/summary",
	}

	for _, uri := range expectedResources {
		resource, exists := server.GetResource(uri)
		if !exists {
			continue
		}
		if resource.Handler == nil {
			t.Errorf("resource %q handler is nil", uri)
		}
	}
}
