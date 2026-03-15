package cli

import (
	"testing"

	"github.com/racingmars/go3270"
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

func TestSplitIntoLines(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLines  int
		maxLen    int
		wantLen   int
		wantFirst string // expected first line content (empty means skip check)
	}{
		{
			name:     "empty string pads to maxLines",
			text:     "",
			maxLines: 3,
			maxLen:   20,
			wantLen:  3,
		},
		{
			name:      "short text fits on one line",
			text:      "hello",
			maxLines:  3,
			maxLen:    20,
			wantLen:   3,
			wantFirst: "hello",
		},
		{
			name:      "exact length text fits on one line",
			text:      "1234567890",
			maxLines:  2,
			maxLen:    10,
			wantLen:   2,
			wantFirst: "1234567890",
		},
		{
			name:     "text longer than maxLen splits into multiple lines",
			text:     "word1 word2 word3 word4 word5",
			maxLines: 4,
			maxLen:   12,
			wantLen:  4,
		},
		{
			name:     "maxLines=1 limits to one line",
			text:     "this is a long string that exceeds one line of twenty chars for sure",
			maxLines: 1,
			maxLen:   20,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoLines(tt.text, tt.maxLines, tt.maxLen)
			if len(result) != tt.wantLen {
				t.Errorf("splitIntoLines(%q, %d, %d) len = %d, want %d", tt.text, tt.maxLines, tt.maxLen, len(result), tt.wantLen)
			}
			if tt.wantFirst != "" && len(result) > 0 && result[0] != tt.wantFirst {
				t.Errorf("splitIntoLines(%q, %d, %d)[0] = %q, want %q", tt.text, tt.maxLines, tt.maxLen, result[0], tt.wantFirst)
			}
			// Verify no line exceeds maxLen
			for i, line := range result {
				if len(line) > tt.maxLen {
					t.Errorf("splitIntoLines(%q, %d, %d)[%d] len=%d exceeds maxLen=%d", tt.text, tt.maxLines, tt.maxLen, i, len(line), tt.maxLen)
				}
			}
		})
	}
}

// colorTestCase groups a status/priority string with whether it should produce
// a non-default color when NO_COLOR is not set.
type colorTestCase struct {
	name        string
	input       string
	wantDefault bool // true if the function should always return DefaultColor
}

func TestStatusColor(t *testing.T) {
	tests := []colorTestCase{
		{"Todo", "Todo", false},
		{"In Progress", "In Progress", false},
		{"Done", "Done", false},
		{"Review", "Review", false},
		{"unknown status", "Pending", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusColor(tt.input)
			if noColor3270 {
				// When NO_COLOR is set every call must return DefaultColor.
				if got != go3270.DefaultColor {
					t.Errorf("statusColor(%q) with NO_COLOR = %v, want DefaultColor", tt.input, got)
				}
				return
			}
			if tt.wantDefault {
				if got != go3270.DefaultColor {
					t.Errorf("statusColor(%q) = %v, want DefaultColor", tt.input, got)
				}
			} else {
				if got == go3270.DefaultColor {
					t.Errorf("statusColor(%q) = DefaultColor, want a specific color", tt.input)
				}
			}
		})
	}
}

func TestPriorityColor(t *testing.T) {
	tests := []colorTestCase{
		{"high", "high", false},
		{"High mixed case", "High", false},
		{"HIGH uppercase", "HIGH", false},
		{"medium", "medium", false},
		{"low", "low", false},
		{"empty string", "", true},
		{"unknown priority", "critical", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityColor(tt.input)
			if noColor3270 {
				if got != go3270.DefaultColor {
					t.Errorf("priorityColor(%q) with NO_COLOR = %v, want DefaultColor", tt.input, got)
				}
				return
			}
			if tt.wantDefault {
				if got != go3270.DefaultColor {
					t.Errorf("priorityColor(%q) = %v, want DefaultColor", tt.input, got)
				}
			} else {
				if got == go3270.DefaultColor {
					t.Errorf("priorityColor(%q) = DefaultColor, want a specific color", tt.input)
				}
			}
		})
	}
}

// scorecardColorCase pairs a scorecard line with its expected color when NO_COLOR is not set.
type scorecardColorCase struct {
	name          string
	line          string
	wantColorNoNC go3270.Color // expected when NO_COLOR is not set
}

func TestScorecardLineColor(t *testing.T) {
	tests := []scorecardColorCase{
		{"section header ===", "=== Build ===", go3270.Green},
		{"PASS keyword", "Build: PASS", go3270.Green},
		{"FAIL keyword", "Tests: FAIL", go3270.Red},
		{"checkmark pass", "lint \u2713", go3270.Green},
		{"cross fail", "vet \u2717", go3270.Red},
		{"green checkmark emoji", "lint \u2705", go3270.Green},
		{"red cross emoji", "lint \u274c", go3270.Red},
		{"score 90/100 green", "Score: 90/100", go3270.Green},
		{"score 60/100 yellow", "Score: 60/100", go3270.Yellow},
		{"score 30/100 red", "Score: 30/100", go3270.Red},
		{"85 percent green", "Coverage: 85%", go3270.Green},
		{"55 percent yellow", "Coverage: 55%", go3270.Yellow},
		{"30 percent red", "Coverage: 30%", go3270.Red},
		{"plain line default", "just some text", go3270.DefaultColor},
		{"empty string default", "", go3270.DefaultColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorecardLineColor(tt.line)
			if noColor3270 {
				// With NO_COLOR all calls return DefaultColor.
				if got != go3270.DefaultColor {
					t.Errorf("scorecardLineColor(%q) with NO_COLOR = %v, want DefaultColor", tt.line, got)
				}
				return
			}
			if got != tt.wantColorNoNC {
				t.Errorf("scorecardLineColor(%q) = %v, want %v", tt.line, got, tt.wantColorNoNC)
			}
		})
	}
}

func TestNextStatusFilter(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{"empty cycles to Todo", "", "Todo"},
		{"Todo cycles to In Progress", "Todo", "In Progress"},
		{"In Progress cycles to Review", "In Progress", "Review"},
		{"Review cycles to Done", "Review", "Done"},
		{"Done cycles to empty", "Done", ""},
		{"unknown value returns first in list", "Unknown", "Todo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextStatusFilter(tt.current)
			if got != tt.want {
				t.Errorf("nextStatusFilter(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestPushPopSession(t *testing.T) {
	t.Run("pop on empty stack returns nil", func(t *testing.T) {
		state := &tui3270State{}
		got := state.popSession()
		if got != nil {
			t.Errorf("popSession on empty state = %v, want nil", got)
		}
	})

	t.Run("push then pop returns same session name", func(t *testing.T) {
		state := &tui3270State{}
		state.pushSession("TestSession", nil)
		if len(state.sessionStack) != 1 {
			t.Fatalf("expected 1 session on stack, got %d", len(state.sessionStack))
		}
		popped := state.popSession()
		if popped == nil {
			t.Fatal("expected non-nil popped session")
		}
		if popped.name != "TestSession" {
			t.Errorf("popped.name = %q, want %q", popped.name, "TestSession")
		}
		if len(state.sessionStack) != 0 {
			t.Errorf("expected 0 sessions after pop, got %d", len(state.sessionStack))
		}
	})

	t.Run("stack is LIFO order", func(t *testing.T) {
		state := &tui3270State{}
		state.pushSession("First", nil)
		state.pushSession("Second", nil)
		state.pushSession("Third", nil)

		third := state.popSession()
		if third == nil || third.name != "Third" {
			t.Errorf("expected Third, got %v", third)
		}
		second := state.popSession()
		if second == nil || second.name != "Second" {
			t.Errorf("expected Second, got %v", second)
		}
		first := state.popSession()
		if first == nil || first.name != "First" {
			t.Errorf("expected First, got %v", first)
		}
		empty := state.popSession()
		if empty != nil {
			t.Errorf("expected nil after all pops, got %v", empty)
		}
	})

	t.Run("stack is capped at 8 entries", func(t *testing.T) {
		state := &tui3270State{}
		for i := 0; i < 10; i++ {
			state.pushSession("entry", nil)
		}
		if len(state.sessionStack) > 8 {
			t.Errorf("expected stack <= 8, got %d", len(state.sessionStack))
		}
	})
}

func TestStateInitialization(t *testing.T) {
	t.Run("default state fields are set correctly", func(t *testing.T) {
		state := &tui3270State{
			status:  "Todo",
			mode:    "tasks",
			cursor:  0,
			command: "",
			filter:  "",
		}
		if state.status != "Todo" {
			t.Errorf("status = %q, want %q", state.status, "Todo")
		}
		if state.mode != "tasks" {
			t.Errorf("mode = %q, want %q", state.mode, "tasks")
		}
		if state.cursor != 0 {
			t.Errorf("cursor = %d, want 0", state.cursor)
		}
		if state.command != "" {
			t.Errorf("command = %q, want empty", state.command)
		}
		if state.filter != "" {
			t.Errorf("filter = %q, want empty", state.filter)
		}
		if len(state.sessionStack) != 0 {
			t.Errorf("sessionStack len = %d, want 0", len(state.sessionStack))
		}
	})

	t.Run("scorecardRecs starts nil", func(t *testing.T) {
		state := &tui3270State{}
		if state.scorecardRecs != nil {
			t.Errorf("scorecardRecs = %v, want nil", state.scorecardRecs)
		}
	})

	t.Run("scorecardFullModeNext defaults false", func(t *testing.T) {
		state := &tui3270State{}
		if state.scorecardFullModeNext {
			t.Error("scorecardFullModeNext should default to false")
		}
	})
}

func TestT3270LayoutConstants(t *testing.T) {
	// Column positions must be positive (1-indexed on mainframe).
	cols := map[string]int{
		"t3270ColS":        t3270ColS,
		"t3270ColID":       t3270ColID,
		"t3270ColStatus":   t3270ColStatus,
		"t3270ColPriority": t3270ColPriority,
		"t3270ColContent":  t3270ColContent,
	}
	for name, val := range cols {
		if val <= 0 {
			t.Errorf("column constant %s = %d, want > 0", name, val)
		}
	}

	// Width constants must be positive.
	widths := map[string]int{
		"t3270WidS":        t3270WidS,
		"t3270WidID":       t3270WidID,
		"t3270WidStatus":   t3270WidStatus,
		"t3270WidPriority": t3270WidPriority,
		"t3270WidContent":  t3270WidContent,
	}
	for name, val := range widths {
		if val <= 0 {
			t.Errorf("width constant %s = %d, want > 0", name, val)
		}
	}

	// Row constants must be positive.
	rows := map[string]int{
		"t3270HeaderRow":    t3270HeaderRow,
		"t3270StatusBarRow": t3270StatusBarRow,
		"t3270PFKeyRow":     t3270PFKeyRow,
	}
	for name, val := range rows {
		if val <= 0 {
			t.Errorf("row constant %s = %d, want > 0", name, val)
		}
	}
}
