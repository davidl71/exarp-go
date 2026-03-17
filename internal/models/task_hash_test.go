// task_hash_test.go — Unit tests for NormalizeForComparison and content hash (tasksync matching).
package models

import (
	"testing"
)

func TestNormalizeForComparison(t *testing.T) {
	tests := []struct {
		content     string
		description string
		want        string
	}{
		// Empty
		{"", "", ""},
		{"  ", "", ""},
		{"", "  ", ""},
		{"  ", "  ", ""},
		// Whitespace: trim
		{"  hello  ", "", "hello"},
		{"  hello  world  ", "", "hello world"},
		// Whitespace: collapse multiple spaces
		{"hello    world", "", "hello world"},
		{"a   b   c", "", "a b c"},
		// Case: lowercase
		{"Hello World", "", "hello world"},
		{"T-123", "", "t-123"},
		{"T-123", "Fix BUG", "t-123 fix bug"},
		// Combined content + description
		{"Task", "Do something", "task do something"},
		{"  Task  ", "  Do something  ", "task do something"},
		// Newlines and tabs collapse to single space
		{"hello\nworld", "", "hello world"},
		{"hello\tworld", "", "hello world"},
		{"hello \n \t world", "", "hello world"},
	}
	for _, tt := range tests {
		got := NormalizeForComparison(tt.content, tt.description)
		if got != tt.want {
			t.Errorf("NormalizeForComparison(%q, %q) = %q, want %q", tt.content, tt.description, got, tt.want)
		}
	}
}

func TestNormalizeForComparisonMatching(t *testing.T) {
	// Pairs that should normalize to the same string (tasksync matching).
	pairs := [][2]struct{ content, description string }{
		{{"T-123 Fix bug", ""}, {"t-123 fix bug", ""}},
		{{"  Add feature  ", ""}, {"add feature", ""}},
		{{"Hello   World", ""}, {"hello world", ""}},
		{{"Task", "Desc"}, {"task", "desc"}},
		{{"TASK", "  DESC  "}, {"task", "desc"}},
	}
	for i, p := range pairs {
		a := NormalizeForComparison(p[0].content, p[0].description)
		b := NormalizeForComparison(p[1].content, p[1].description)
		if a != b {
			t.Errorf("pair %d: NormalizeForComparison(%q, %q) = %q vs NormalizeForComparison(%q, %q) = %q; should match",
				i, p[0].content, p[0].description, a, p[1].content, p[1].description, b)
		}
	}
}

func TestContentHashFromString(t *testing.T) {
	// Deterministic: same input => same hash
	h1 := ContentHashFromString("hello world")
	h2 := ContentHashFromString("hello world")
	if h1 != h2 {
		t.Errorf("ContentHashFromString same input gave different hashes: %q vs %q", h1, h2)
	}
	// Different input => different hash
	h3 := ContentHashFromString("hello world!")
	if h1 == h3 {
		t.Errorf("ContentHashFromString different input should give different hashes")
	}
	// Hex length (SHA-256 => 64 hex chars)
	if len(h1) != 64 {
		t.Errorf("ContentHashFromString hex length = %d, want 64", len(h1))
	}
}

func TestContentHash(t *testing.T) {
	// Same normalized content => same hash (matching for tasksync)
	t1 := &Todo2Task{Content: "  T-123 Fix bug  ", LongDescription: ""}
	t2 := &Todo2Task{Content: "t-123 fix bug", LongDescription: ""}
	h1 := ContentHash(t1)
	h2 := ContentHash(t2)
	if h1 != h2 {
		t.Errorf("ContentHash: same normalized content should match; got %q vs %q", h1, h2)
	}
	// Different content => different hash
	t3 := &Todo2Task{Content: "Other task", LongDescription: ""}
	h3 := ContentHash(t3)
	if h1 == h3 {
		t.Errorf("ContentHash: different content should differ")
	}
}

func TestSetContentHashAndGetContentHash(t *testing.T) {
	task := &Todo2Task{Content: "Test", LongDescription: ""}
	if got := GetContentHash(task); got != "" {
		t.Errorf("GetContentHash before Set = %q, want empty", got)
	}
	SetContentHash(task)
	got := GetContentHash(task)
	if got == "" {
		t.Errorf("GetContentHash after Set should be non-empty")
	}
	if len(got) != 64 {
		t.Errorf("GetContentHash hex length = %d, want 64", len(got))
	}
}

func TestEnsureContentHash(t *testing.T) {
	task := &Todo2Task{Content: "Test", LongDescription: ""}
	EnsureContentHash(task)
	h1 := GetContentHash(task)
	if h1 == "" {
		t.Errorf("EnsureContentHash should set hash")
	}
	EnsureContentHash(task)
	h2 := GetContentHash(task)
	if h1 != h2 {
		t.Errorf("EnsureContentHash should not overwrite existing hash: %q vs %q", h1, h2)
	}
}
