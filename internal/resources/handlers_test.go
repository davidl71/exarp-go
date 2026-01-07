package resources

import (
	"testing"

	"github.com/davidl/mcp-stdio-tools/tests/fixtures"
)

func TestRegisterAllResources(t *testing.T) {
	server := fixtures.NewMockServer("test-server")

	err := RegisterAllResources(server)
	if err != nil {
		t.Fatalf("RegisterAllResources() error = %v", err)
	}

	// Verify all 6 resources are registered
	if server.ResourceCount() != 6 {
		t.Errorf("server.ResourceCount() = %v, want 6", server.ResourceCount())
	}

	// Verify specific resources are registered
	expectedResources := []string{
		"stdio://scorecard",
		"stdio://memories",
		"stdio://memories/category/{category}",
		"stdio://memories/task/{task_id}",
		"stdio://memories/recent",
		"stdio://memories/session/{date}",
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
