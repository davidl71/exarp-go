package resources

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleCursorSkillByName(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	// Ensure project root resolution is stable under test runs even if the
	// environment has PROJECT_ROOT set to some other workspace.
	// internal/resources/ → repo root is ../..
	t.Setenv("PROJECT_ROOT", filepath.Clean(filepath.Join(wd, "..", "..")))

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

func TestHandleAgentSkillByNameAlias(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Setenv("PROJECT_ROOT", filepath.Clean(filepath.Join(wd, "..", "..")))

	data, mimeType, err := handleCursorSkillByName(context.Background(), "stdio://agent/skills/use-exarp-tools")
	if err != nil {
		t.Fatalf("handleCursorSkillByName(agent alias): %v", err)
	}
	if mimeType != "text/markdown" {
		t.Fatalf("mimeType = %q, want text/markdown", mimeType)
	}
	if !strings.Contains(string(data), "# Using exarp-go MCP Tools") {
		t.Fatalf("unexpected skill body: %q", string(data))
	}
}
