package models

import "testing"

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"", "", false},
		{"   ", "", false},
		{"#", "", false},
		{"##", "", false},
		{"cli", "#cli", true},
		{"#CLI", "#cli", true},
		{"  #cache  ", "#cache", true},
		{"perf/bench", "#perf-bench", true},
		{"a..b", "#a-b", true},
		{"--a--", "#a", true},
		{"__a__", "#a", true},
		{"🚫", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalizeTag(tt.in)
		if ok != tt.ok || got != tt.want {
			t.Fatalf("NormalizeTag(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNormalizeTags_SplitsWhitespaceAndDedupes(t *testing.T) {
	got := NormalizeTags([]string{"a b", "#b", "  c  ", "A", "a"})
	want := []string{"#a", "#b", "#c"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}
