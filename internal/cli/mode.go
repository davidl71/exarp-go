package cli

import (
	"encoding/json"
	"strings"
)

// ReservedSubcommands are first-argument values that are CLI subcommands,
// not tool names. Used by mode detection so "exarp-go task list" stays CLI.
var ReservedSubcommands = []string{"task", "config", "tui", "tui3270"}

// CLIFlags are arguments that indicate CLI mode (e.g. from hooks or scripts).
// If any of these appear in os.Args, we run CLI instead of MCP server.
var CLIFlags = []string{
	"-completion", "--completion",
	"-list", "--list",
	"-tool", "--tool",
	"-test", "--test",
	"-i", "--interactive",
	"-args", "--args",
	"-h", "--help", "help",
}

// isReservedSubcommand returns true if s is a reserved CLI subcommand.
func isReservedSubcommand(s string) bool {
	for _, r := range ReservedSubcommands {
		if s == r {
			return true
		}
	}
	return false
}

// NormalizeToolArgs rewrites "exarp-go tool_name key=value ..." to
// ["exarp-go", "-tool", "tool_name", "-args", `{"key":"value"}`].
// Returns the new args and true if rewritten; otherwise returns args unchanged and false.
// Used so hooks/scripts with non-TTY stdin run CLI mode instead of MCP (which would
// read stdin as JSON-RPC and fail on e.g. "refs/heads/main ...").
func NormalizeToolArgs(args []string) ([]string, bool) {
	if len(args) < 2 || len(args[1]) == 0 || args[1][0] == '-' {
		return args, false
	}
	first := args[1]
	if isReservedSubcommand(first) || strings.Contains(first, "/") {
		return args, false
	}
	argsMap := make(map[string]interface{})
	for _, arg := range args[2:] {
		if idx := strings.IndexByte(arg, '='); idx >= 0 {
			key := strings.TrimSpace(arg[:idx])
			val := strings.TrimSpace(arg[idx+1:])
			if key != "" {
				argsMap[key] = val
			}
		}
	}
	argsJSON, err := json.Marshal(argsMap)
	if err != nil {
		return args, false
	}
	return []string{args[0], "-tool", first, "-args", string(argsJSON)}, true
}

// HasCLIFlags returns true if args indicate CLI mode (subcommand or flag).
func HasCLIFlags(args []string) bool {
	if len(args) < 2 {
		return false
	}
	if isReservedSubcommand(args[1]) {
		return true
	}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		for _, f := range CLIFlags {
			if arg == f {
				return true
			}
		}
		if len(arg) > 0 && arg[0] != '-' {
			break
		}
	}
	return false
}
