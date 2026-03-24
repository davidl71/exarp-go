package resources

import (
	"context"
	"strings"
	"testing"
)

func TestHandleCursorSkillByName(t *testing.T) {
	data, mimeType, err := handleCursorSkillByName(context.Background(), "stdio://cursor/skills/use-exarp-tools")
	if err != nil {
		t.Fatalf("handleCursorSkillByName: %v", err)
	}
	if mimeType != "text/markdown" {
		t.Fatalf("mimeType = %q, want text/markdown", mimeType)
	}
	if !strings.Contains(string(data), "# Using exarp-go MCP Tools") {
		t.Fatalf("unexpected skill body: %q", string(data))
	}
}
