package cli

import (
	"encoding/json"
	"testing"
)

func TestNormalizeToolArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantPrefix  []string // first elements must match
		wantJSONKey string   // if set, 5th arg must be JSON with this key (and value checked below)
		wantJSONVal string   // expected value for wantJSONKey
		wantOK      bool
	}{
		{
			name:        "tool with key=value",
			args:        []string{"exarp-go", "analyze_alignment", "action=todo2"},
			wantPrefix:  []string{"exarp-go", "-tool", "analyze_alignment", "-args"},
			wantJSONKey: "action",
			wantJSONVal: "todo2",
			wantOK:      true,
		},
		{
			name:        "tool with multiple key=value",
			args:        []string{"exarp-go", "security", "action=scan", "repo=."},
			wantPrefix:  []string{"exarp-go", "-tool", "security", "-args"},
			wantJSONKey: "action",
			wantJSONVal: "scan",
			wantOK:      true,
		},
		{
			name:       "task subcommand not normalized",
			args:       []string{"exarp-go", "task", "list"},
			wantPrefix: []string{"exarp-go", "task", "list"},
			wantOK:     false,
		},
		{
			name:       "config subcommand not normalized",
			args:       []string{"exarp-go", "config", "show"},
			wantPrefix: []string{"exarp-go", "config", "show"},
			wantOK:     false,
		},
		{
			name:       "tui subcommand not normalized",
			args:       []string{"exarp-go", "tui"},
			wantPrefix: []string{"exarp-go", "tui"},
			wantOK:     false,
		},
		{
			name:       "tui3270 subcommand not normalized",
			args:       []string{"exarp-go", "tui3270"},
			wantPrefix: []string{"exarp-go", "tui3270"},
			wantOK:     false,
		},
		{
			name:       "first arg with slash not normalized",
			args:       []string{"exarp-go", "some/path"},
			wantPrefix: []string{"exarp-go", "some/path"},
			wantOK:     false,
		},
		{
			name:       "first arg is flag not normalized",
			args:       []string{"exarp-go", "-tool", "lint"},
			wantPrefix: []string{"exarp-go", "-tool", "lint"},
			wantOK:     false,
		},
		{
			name:       "empty args",
			args:       []string{"exarp-go"},
			wantPrefix: []string{"exarp-go"},
			wantOK:     false,
		},
		{
			name:        "tool with no key=value",
			args:        []string{"exarp-go", "health"},
			wantPrefix:  []string{"exarp-go", "-tool", "health", "-args"},
			wantJSONKey: "",
			wantOK:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, gotOK := NormalizeToolArgs(tt.args)
			if gotOK != tt.wantOK {
				t.Errorf("NormalizeToolArgs() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotOK {
				for i := 0; i < len(tt.wantPrefix) && i < len(gotArgs); i++ {
					if gotArgs[i] != tt.wantPrefix[i] {
						t.Errorf("NormalizeToolArgs() [%d] = %q, want %q", i, gotArgs[i], tt.wantPrefix[i])
					}
				}
				if tt.wantJSONKey != "" && len(gotArgs) >= 5 {
					var m map[string]interface{}
					if err := json.Unmarshal([]byte(gotArgs[4]), &m); err != nil {
						t.Errorf("NormalizeToolArgs() args[4] not valid JSON: %v", err)
					} else if v, ok := m[tt.wantJSONKey].(string); !ok || v != tt.wantJSONVal {
						t.Errorf("NormalizeToolArgs() args[4][%q] = %v, want %q", tt.wantJSONKey, m[tt.wantJSONKey], tt.wantJSONVal)
					}
				}
				if tt.name == "tool with no key=value" && gotArgs[4] != "{}" {
					t.Errorf("NormalizeToolArgs() args[4] = %q, want {}", gotArgs[4])
				}
			} else {
				for i := 0; i < len(tt.wantPrefix) && i < len(gotArgs); i++ {
					if gotArgs[i] != tt.wantPrefix[i] {
						t.Errorf("NormalizeToolArgs() [%d] = %q, want %q", i, gotArgs[i], tt.wantPrefix[i])
					}
				}
			}
		})
	}
}

func TestHasCLIFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"task subcommand", []string{"exarp-go", "task", "list"}, true},
		{"config subcommand", []string{"exarp-go", "config"}, true},
		{"tui subcommand", []string{"exarp-go", "tui"}, true},
		{"tui3270 subcommand", []string{"exarp-go", "tui3270"}, true},
		{"-tool flag", []string{"exarp-go", "-tool", "lint"}, true},
		{"-args flag", []string{"exarp-go", "-args", "{}"}, true},
		{"-list flag", []string{"exarp-go", "-list"}, true},
		{"-h flag", []string{"exarp-go", "-h"}, true},
		{"help", []string{"exarp-go", "help"}, true},
		{"no flags", []string{"exarp-go"}, false},
		{"tool name only (no flags)", []string{"exarp-go", "analyze_alignment"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasCLIFlags(tt.args); got != tt.want {
				t.Errorf("HasCLIFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReservedSubcommands(t *testing.T) {
	if len(ReservedSubcommands) != 4 {
		t.Errorf("ReservedSubcommands len = %d, want 4", len(ReservedSubcommands))
	}
	want := map[string]bool{"task": true, "config": true, "tui": true, "tui3270": true}
	for _, s := range ReservedSubcommands {
		if !want[s] {
			t.Errorf("ReservedSubcommands contains unexpected %q", s)
		}
	}
}

func TestCLIFlags(t *testing.T) {
	if len(CLIFlags) == 0 {
		t.Error("CLIFlags should not be empty")
	}
	// -tool and -args must be present for normalized invocations
	hasTool, hasArgs := false, false
	for _, f := range CLIFlags {
		if f == "-tool" || f == "--tool" {
			hasTool = true
		}
		if f == "-args" || f == "--args" {
			hasArgs = true
		}
	}
	if !hasTool || !hasArgs {
		t.Errorf("CLIFlags missing -tool or -args: hasTool=%v hasArgs=%v", hasTool, hasArgs)
	}
}
