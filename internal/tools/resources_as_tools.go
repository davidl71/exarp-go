// resources_as_tools.go — Synthetic tools that expose MCP resources to tool-only clients.
// Inspired by FastMCP 3.0 ResourcesAsTools: auto-generates read_resource and list_resources tools
// so clients that only support tools (e.g. OpenCode plugin tools) can access resource data.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/davidl71/exarp-go/internal/framework"
)

type resourceEntry struct {
	URI         string                   `json:"uri"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	MimeType    string                   `json:"mime_type"`
	handler     framework.ResourceHandler `json:"-"`
}

var (
	resourceRegistry   []resourceEntry
	resourceRegistryMu sync.RWMutex
)

// TrackResource records a resource for the read_resource/list_resources tools.
// Call this alongside server.RegisterResource to maintain the registry.
func TrackResource(uri, name, description, mimeType string, handler framework.ResourceHandler) {
	resourceRegistryMu.Lock()
	defer resourceRegistryMu.Unlock()
	resourceRegistry = append(resourceRegistry, resourceEntry{
		URI: uri, Name: name, Description: description, MimeType: mimeType, handler: handler,
	})
}

// RegisterResourcesAsTools registers read_resource and list_resources tools.
func RegisterResourcesAsTools(server framework.MCPServer) error {
	if err := server.RegisterTool(
		"read_resource",
		"[HINT: Read a registered MCP resource by URI. Use list_resources to discover available URIs.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"uri": map[string]interface{}{
					"type":        "string",
					"description": "Resource URI (e.g. stdio://scorecard)",
				},
			},
			Required: []string{"uri"},
		},
		handleReadResource,
	); err != nil {
		return fmt.Errorf("failed to register read_resource: %w", err)
	}

	if err := server.RegisterTool(
		"list_resources",
		"[HINT: List all registered MCP resources with URIs, names, and descriptions.]",
		framework.ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
		handleListResources,
	); err != nil {
		return fmt.Errorf("failed to register list_resources: %w", err)
	}

	return nil
}

func handleReadResource(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}
	if params.URI == "" {
		return nil, fmt.Errorf("uri is required")
	}

	resourceRegistryMu.RLock()
	var handler framework.ResourceHandler
	for _, entry := range resourceRegistry {
		if entry.URI == params.URI {
			handler = entry.handler
			break
		}
	}
	resourceRegistryMu.RUnlock()

	if handler == nil {
		return nil, fmt.Errorf("resource not found: %s", params.URI)
	}

	data, mimeType, err := handler(ctx, params.URI)
	if err != nil {
		return nil, fmt.Errorf("resource %s: %w", params.URI, err)
	}

	result := map[string]interface{}{
		"uri":       params.URI,
		"mime_type": mimeType,
		"content":   string(data),
	}
	out, _ := json.MarshalIndent(result, "", "  ")

	return []framework.TextContent{{Text: string(out)}}, nil
}

type resourceListEntry struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mime_type"`
}

func handleListResources(_ context.Context, _ json.RawMessage) ([]framework.TextContent, error) {
	resourceRegistryMu.RLock()
	defer resourceRegistryMu.RUnlock()

	entries := make([]resourceListEntry, len(resourceRegistry))
	for i, r := range resourceRegistry {
		entries[i] = resourceListEntry{URI: r.URI, Name: r.Name, Description: r.Description, MimeType: r.MimeType}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].URI < entries[j].URI })

	data, err := json.MarshalIndent(map[string]interface{}{
		"resources": entries,
		"count":     len(entries),
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources: %w", err)
	}

	return []framework.TextContent{{Text: string(data)}}, nil
}
