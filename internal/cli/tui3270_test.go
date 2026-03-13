package cli

import (
	"testing"
)

func TestExtractMenuOption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single digit 1", "1", "1"},
		{"single digit 7", "7", "7"},
		{"out of range digit", "8", ""},
		{"zero", "0", ""},
		{"option prefix", "Option: 3", "3"},
		{"first digit in string", "Select option 5 here", "5"},
		{"last digit wins", "Option 1 and 2", "2"},
		{"leading spaces", "  4", "4"},
		{"trailing spaces", "6  ", "6"},
		{"non-digit string", "abc", ""},
		{"mixed content", "abc5def", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMenuOption(tt.input)
			if result != tt.expected {
				t.Errorf("extractMenuOption(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRecommendationToCommand(t *testing.T) {
	tests := []struct {
		name     string
		rec      string
		wantName string
		wantArgs []string
		wantOk   bool
	}{
		{"go mod tidy", "run go mod tidy", "make", []string{"go-mod-tidy"}, true},
		{"go fmt", "try go fmt", "make", []string{"go-fmt"}, true},
		{"go vet", "execute go vet", "make", []string{"go-vet"}, true},
		{"fix build", "Fix Go build", "make", []string{"build"}, true},
		{"golangci-lint", "fix golangci-lint issues", "make", []string{"golangci-lint-fix"}, true},
		{"go test", "run failing Go tests", "make", []string{"test"}, true},
		{"go test alt", "run go test", "make", []string{"test"}, true},
		{"govulncheck", "run govulncheck", "make", []string{"govulncheck"}, true},
		{"test coverage", "check test coverage", "make", []string{"test-coverage"}, true},
		{"unknown recommendation", "do something else", "", nil, false},
		{"empty string", "", "", nil, false},
		{"whitespace only", "   ", "", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, ok := recommendationToCommand(tt.rec)
			if ok != tt.wantOk {
				t.Errorf("recommendationToCommand(%q) ok = %v, want %v", tt.rec, ok, tt.wantOk)
				return
			}
			if name != tt.wantName {
				t.Errorf("recommendationToCommand(%q) name = %q, want %q", tt.rec, name, tt.wantName)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("recommendationToCommand(%q) args = %v, want %v", tt.rec, args, tt.wantArgs)
				return
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("recommendationToCommand(%q) args[%d] = %q, want %q", tt.rec, i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "help", "help"},
		{"two words", "help foo", "help"},
		{"many words", "task list --status Todo", "task"},
		{"leading spaces", "  task", "task"},
		{"trailing spaces", "task  ", "task"},
		{"multiple spaces between", "task    list", "task"},
		{"tabs", "task\tlist", "task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommand(tt.input)
			if result != tt.expected {
				t.Errorf("parseCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
