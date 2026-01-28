package models

import "testing"

func TestIsEpochDate(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"", true},
		{"0", true},
		{"0.0", true},
		{"  0  ", true},
		{"1970-01-01T00:00:00.000Z", true},
		{"1970-01-01T02:00:00.000Z", true},
		{"1970-01-01", true},
		{"2026-01-28T12:00:00Z", false},
		{"2026-01-28", false},
	}
	for _, tt := range tests {
		if got := IsEpochDate(tt.s); got != tt.want {
			t.Errorf("IsEpochDate(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestNormalizeEpochDates(t *testing.T) {
	task := &Todo2Task{
		ID:           "T-1",
		CreatedAt:    "1970-01-01T00:00:00.000Z",
		LastModified: "1970-01-01T00:00:00.000Z",
		CompletedAt:  "2026-01-28T12:00:00Z",
	}
	task.NormalizeEpochDates()
	if task.CreatedAt != "" {
		t.Errorf("CreatedAt should be empty after normalize, got %q", task.CreatedAt)
	}
	if task.LastModified != "" {
		t.Errorf("LastModified should be empty after normalize, got %q", task.LastModified)
	}
	if task.CompletedAt != "2026-01-28T12:00:00Z" {
		t.Errorf("CompletedAt should be unchanged, got %q", task.CompletedAt)
	}
}
