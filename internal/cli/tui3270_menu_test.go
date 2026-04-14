package cli

import "testing"

func TestExtractMenuOption(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"1", "1"},
		{"7", "7"},
		{"  3  ", "3"},
		{"Option 5", "5"},
		{"foo2bar", "2"},
		{"8", ""},
		{"0", ""},
	}
	for _, tc := range cases {
		got := extractMenuOption(tc.in)
		if got != tc.want {
			t.Errorf("extractMenuOption(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestT3270VerbDispatchKeys(t *testing.T) {
	t.Parallel()
	required := []string{
		"1", "2", "3", "4", "5", "7",
		"SC", "SCORECARD", "HANDOFFS", "HO",
		"MENU", "M", "MAIN", "TASKS", "T", "CONFIG",
		"HELP", "H", "HEALTH", "SDSF", "GIT", "GITLOG",
		"SPRINT", "BOARD", "SWAP",
	}
	for _, k := range required {
		if _, ok := t3270VerbDispatch[k]; !ok {
			t.Errorf("t3270VerbDispatch missing key %q", k)
		}
	}
}
