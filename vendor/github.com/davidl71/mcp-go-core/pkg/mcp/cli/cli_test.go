package cli

import (
	"os"
	"testing"
)

func TestIsTTY(t *testing.T) {
	// This test will pass or fail depending on the environment
	// We just verify the function doesn't panic
	_ = IsTTY()
}

func TestIsTTYFile(t *testing.T) {
	// Test with stdin
	_ = IsTTYFile(os.Stdin)

	// Test with stdout
	_ = IsTTYFile(os.Stdout)

	// Test with stderr
	_ = IsTTYFile(os.Stderr)
}

func TestDetectMode(t *testing.T) {
	mode := DetectMode()
	if mode != ModeCLI && mode != ModeMCP {
		t.Errorf("DetectMode() = %q, want %q or %q", mode, ModeCLI, ModeMCP)
	}
}

func TestParseArgs_SimpleCommand(t *testing.T) {
	args := ParseArgs([]string{"tool", "list"})

	if args.Command != "tool" {
		t.Errorf("args.Command = %q, want %q", args.Command, "tool")
	}
	if args.Subcommand != "list" {
		t.Errorf("args.Subcommand = %q, want %q", args.Subcommand, "list")
	}
	if len(args.Positional) != 0 {
		t.Errorf("args.Positional = %v, want []", args.Positional)
	}
}

func TestParseArgs_WithFlags(t *testing.T) {
	args := ParseArgs([]string{"tool", "call", "--name", "my_tool", "--arg", "value"})

	if args.Command != "tool" {
		t.Errorf("args.Command = %q, want %q", args.Command, "tool")
	}
	if args.Subcommand != "call" {
		t.Errorf("args.Subcommand = %q, want %q", args.Subcommand, "call")
	}
	if args.GetFlag("name", "") != "my_tool" {
		t.Errorf("args.GetFlag(\"name\") = %q, want %q", args.GetFlag("name", ""), "my_tool")
	}
	if args.GetFlag("arg", "") != "value" {
		t.Errorf("args.GetFlag(\"arg\") = %q, want %q", args.GetFlag("arg", ""), "value")
	}
}

func TestParseArgs_FlagEqualsValue(t *testing.T) {
	args := ParseArgs([]string{"tool", "call", "--name=my_tool", "--arg=value"})

	if args.GetFlag("name", "") != "my_tool" {
		t.Errorf("args.GetFlag(\"name\") = %q, want %q", args.GetFlag("name", ""), "my_tool")
	}
	if args.GetFlag("arg", "") != "value" {
		t.Errorf("args.GetFlag(\"arg\") = %q, want %q", args.GetFlag("arg", ""), "value")
	}
}

func TestParseArgs_ShortFlags(t *testing.T) {
	args := ParseArgs([]string{"tool", "list", "-v", "-f", "file.txt"})

	if !args.HasFlag("v") {
		t.Error("args.HasFlag(\"v\") = false, want true")
	}
	if args.GetFlag("f", "") != "file.txt" {
		t.Errorf("args.GetFlag(\"f\") = %q, want %q", args.GetFlag("f", ""), "file.txt")
	}
}

func TestParseArgs_BooleanFlags(t *testing.T) {
	args := ParseArgs([]string{"tool", "list", "--verbose", "--debug"})

	if !args.GetBoolFlag("verbose", false) {
		t.Error("args.GetBoolFlag(\"verbose\", false) = false, want true")
	}
	if !args.GetBoolFlag("debug", false) {
		t.Error("args.GetBoolFlag(\"debug\", false) = false, want true")
	}
	if args.GetBoolFlag("nonexistent", false) {
		t.Error("args.GetBoolFlag(\"nonexistent\", false) = true, want false")
	}
	if !args.GetBoolFlag("nonexistent", true) {
		t.Error("args.GetBoolFlag(\"nonexistent\", true) = false, want true")
	}
}

func TestParseArgs_PositionalArgs(t *testing.T) {
	args := ParseArgs([]string{"tool", "call", "arg1", "arg2", "arg3"})

	if args.Command != "tool" {
		t.Errorf("args.Command = %q, want %q", args.Command, "tool")
	}
	if args.Subcommand != "call" {
		t.Errorf("args.Subcommand = %q, want %q", args.Subcommand, "call")
	}
	if len(args.Positional) != 3 {
		t.Errorf("len(args.Positional) = %d, want 3", len(args.Positional))
	}
	expected := []string{"arg1", "arg2", "arg3"}
	for i, want := range expected {
		if i >= len(args.Positional) || args.Positional[i] != want {
			t.Errorf("args.Positional[%d] = %q, want %q", i, args.Positional[i], want)
		}
	}
}

func TestParseArgs_Empty(t *testing.T) {
	args := ParseArgs([]string{})

	if args.Command != "" {
		t.Errorf("args.Command = %q, want %q", args.Command, "")
	}
	if args.Subcommand != "" {
		t.Errorf("args.Subcommand = %q, want %q", args.Subcommand, "")
	}
	if len(args.Positional) != 0 {
		t.Errorf("len(args.Positional) = %d, want 0", len(args.Positional))
	}
}

func TestArgs_GetFlag(t *testing.T) {
	args := ParseArgs([]string{"--flag", "value"})

	if args.GetFlag("flag", "default") != "value" {
		t.Errorf("args.GetFlag(\"flag\", \"default\") = %q, want %q", args.GetFlag("flag", "default"), "value")
	}
	if args.GetFlag("nonexistent", "default") != "default" {
		t.Errorf("args.GetFlag(\"nonexistent\", \"default\") = %q, want %q", args.GetFlag("nonexistent", "default"), "default")
	}
}

func TestArgs_HasFlag(t *testing.T) {
	args := ParseArgs([]string{"--flag", "value"})

	if !args.HasFlag("flag") {
		t.Error("args.HasFlag(\"flag\") = false, want true")
	}
	if args.HasFlag("nonexistent") {
		t.Error("args.HasFlag(\"nonexistent\") = true, want false")
	}
}

func TestArgs_GetBoolFlag(t *testing.T) {
	args := ParseArgs([]string{"--enabled", "--disabled=false"})

	if !args.GetBoolFlag("enabled", false) {
		t.Error("args.GetBoolFlag(\"enabled\", false) = false, want true")
	}
	if args.GetBoolFlag("disabled", true) {
		t.Error("args.GetBoolFlag(\"disabled\", true) = true, want false")
	}
}
