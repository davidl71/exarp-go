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

func TestFillRFC3339FromUnix(t *testing.T) {
	t.Parallel()

	const ts = int64(1700000000) // 2023-11-14T22:13:20Z
	want := "2023-11-14T22:13:20Z"

	task := &Todo2Task{
		ID:           "T-1",
		CreatedAt:    "",
		LastModified: "1970-01-01T00:00:00Z",
		CompletedAt:  "",
	}
	task.FillRFC3339FromUnix(ts, ts, 0)

	if task.CreatedAt != want {
		t.Errorf("CreatedAt = %q, want %q", task.CreatedAt, want)
	}
	if task.LastModified != want {
		t.Errorf("LastModified = %q, want %q", task.LastModified, want)
	}
	if task.CompletedAt != "" {
		t.Errorf("CompletedAt should stay empty when completedAtTS is 0, got %q", task.CompletedAt)
	}

	// Do not overwrite non-epoch strings or when unix is zero.
	task2 := &Todo2Task{ID: "T-2", CreatedAt: "2026-06-01T12:00:00Z", LastModified: ""}
	task2.FillRFC3339FromUnix(0, ts, 0)
	if task2.CreatedAt != "2026-06-01T12:00:00Z" {
		t.Errorf("CreatedAt should be unchanged, got %q", task2.CreatedAt)
	}
	if task2.LastModified != want {
		t.Errorf("LastModified = %q, want %q", task2.LastModified, want)
	}
}
