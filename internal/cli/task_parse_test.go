package cli

import (
	"reflect"
	"testing"

	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

func TestParseArgs_TaskCreate_DoesNotTreatFlagValuesAsPositional(t *testing.T) {
	parsed := mcpcli.ParseArgs([]string{
		"task", "create", "Fix", "bug",
		"--description", "Fix the bug",
		"--priority", "high",
		"--tags", "cli,bug",
	})

	if parsed.Command != "task" || parsed.Subcommand != "create" {
		t.Fatalf("unexpected parse result: command=%q subcommand=%q", parsed.Command, parsed.Subcommand)
	}

	wantPositional := []string{"Fix", "bug"}
	if !reflect.DeepEqual(parsed.Positional, wantPositional) {
		t.Fatalf("Positional = %#v, want %#v", parsed.Positional, wantPositional)
	}

	if got := parsed.GetFlag("description", ""); got != "Fix the bug" {
		t.Fatalf("description = %q, want %q", got, "Fix the bug")
	}
	if got := parsed.GetFlag("priority", ""); got != "high" {
		t.Fatalf("priority = %q, want %q", got, "high")
	}
	if got := parsed.GetFlag("tags", ""); got != "cli,bug" {
		t.Fatalf("tags = %q, want %q", got, "cli,bug")
	}
}

func TestParseArgs_TaskEstimate_DoesNotTreatFlagValuesAsPositional(t *testing.T) {
	parsed := mcpcli.ParseArgs([]string{
		"task", "estimate", "Add", "tests",
		"--details", "Cover edge cases",
		"--priority", "medium",
	})

	wantPositional := []string{"Add", "tests"}
	if !reflect.DeepEqual(parsed.Positional, wantPositional) {
		t.Fatalf("Positional = %#v, want %#v", parsed.Positional, wantPositional)
	}

	if got := parsed.GetFlag("details", ""); got != "Cover edge cases" {
		t.Fatalf("details = %q, want %q", got, "Cover edge cases")
	}
	if got := parsed.GetFlag("priority", ""); got != "medium" {
		t.Fatalf("priority = %q, want %q", got, "medium")
	}
}
