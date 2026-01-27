// Package cli provides CLI utilities for MCP servers.
//
// This package provides utilities for detecting execution mode (CLI vs MCP server),
// parsing command-line arguments, and supporting CLI/MCP dual mode operation.
//
// Example usage:
//
//	// Detect if running in CLI mode
//	if cli.IsTTY() {
//		// Run CLI interface
//		cli.RunCLI()
//		return
//	}
//	// Run as MCP server
//	server.Run(ctx, transport)
package cli

import (
	"os"

	"golang.org/x/term"
)

// IsTTY checks if the program is running in a terminal (TTY).
// Returns true if stdin is connected to a terminal, false otherwise.
//
// This is useful for detecting if the program should run in CLI mode
// (when TTY is detected) or MCP server mode (when running via stdio).
//
// Example:
//
//	if cli.IsTTY() {
//		// CLI mode - user is interacting via terminal
//		runCLI()
//	} else {
//		// MCP server mode - running via stdio (e.g., from Cursor)
//		runMCPServer()
//	}
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// IsTTYFile checks if a specific file descriptor is connected to a terminal.
// This allows checking stdout or stderr instead of stdin.
//
// Example:
//
//	if cli.IsTTYFile(os.Stdout) {
//		// Can use colored output
//		fmt.Println("\033[32mSuccess\033[0m")
//	}
func IsTTYFile(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
}

// ExecutionMode represents the detected execution mode
type ExecutionMode string

const (
	// ModeCLI indicates the program is running in CLI mode (TTY detected)
	ModeCLI ExecutionMode = "cli"
	// ModeMCP indicates the program is running as an MCP server (no TTY)
	ModeMCP ExecutionMode = "mcp"
)

// DetectMode detects the execution mode based on TTY detection.
// Returns ModeCLI if running in a terminal, ModeMCP otherwise.
func DetectMode() ExecutionMode {
	if IsTTY() {
		return ModeCLI
	}
	return ModeMCP
}

// Args represents parsed command-line arguments
type Args struct {
	// Command is the main command (e.g., "tool", "prompt", "resource")
	Command string
	// Subcommand is the subcommand (e.g., "list", "call", "get")
	Subcommand string
	// Flags contains parsed flags
	Flags map[string]string
	// Positional contains positional arguments
	Positional []string
}

// ParseArgs parses command-line arguments into a structured Args object.
// This is a simple parser for basic CLI operations.
//
// Example:
//
//	args := cli.ParseArgs(os.Args[1:])
//	if args.Command == "tool" && args.Subcommand == "list" {
//		// List all tools
//	}
func ParseArgs(argv []string) *Args {
	args := &Args{
		Flags:      make(map[string]string),
		Positional: make([]string, 0),
	}

	for i, arg := range argv {
		if arg == "" {
			continue
		}

		// Handle flags (--flag or --flag=value)
		if len(arg) > 2 && arg[0:2] == "--" {
			flag := arg[2:]
			if equals := indexByte(flag, '='); equals >= 0 {
				// --flag=value format
				key := flag[:equals]
				value := flag[equals+1:]
				args.Flags[key] = value
			} else {
				// --flag format (check if next arg is value)
				if i+1 < len(argv) && argv[i+1][0] != '-' {
					args.Flags[flag] = argv[i+1]
					i++ // Skip next arg
				} else {
					args.Flags[flag] = "true" // Boolean flag
				}
			}
			continue
		}

		// Handle short flags (-f or -f value)
		if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
			flag := arg[1:]
			if i+1 < len(argv) && argv[i+1][0] != '-' {
				args.Flags[flag] = argv[i+1]
				i++ // Skip next arg
			} else {
				args.Flags[flag] = "true" // Boolean flag
			}
			continue
		}

		// Positional argument
		if args.Command == "" {
			args.Command = arg
		} else if args.Subcommand == "" {
			args.Subcommand = arg
		} else {
			args.Positional = append(args.Positional, arg)
		}
	}

	return args
}

// indexByte returns the index of the first occurrence of byte c in s,
// or -1 if c is not present in s.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// GetFlag returns the value of a flag, or the default value if not set.
func (a *Args) GetFlag(name, defaultValue string) string {
	if value, ok := a.Flags[name]; ok {
		return value
	}
	return defaultValue
}

// HasFlag checks if a flag is set.
func (a *Args) HasFlag(name string) bool {
	_, ok := a.Flags[name]
	return ok
}

// GetBoolFlag returns true if a flag is set and not "false", otherwise returns defaultValue.
func (a *Args) GetBoolFlag(name string, defaultValue bool) bool {
	if !a.HasFlag(name) {
		return defaultValue
	}
	value := a.GetFlag(name, "true")
	return value != "false" && value != "0" && value != ""
}
