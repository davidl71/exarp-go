package tools

import (
	"testing"
)

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty/whitespace
		{"empty string", "", "todo"},
		{"whitespace", "   ", "todo"},

		// Todo variants
		{"Todo", "Todo", "todo"},
		{"TODO", "TODO", "todo"},
		{"pending", "pending", "todo"},
		{"not started", "not started", "todo"},
		{"new", "new", "todo"},

		// In Progress variants
		{"in_progress", "in_progress", "in_progress"},
		{"in-progress", "in-progress", "in_progress"},
		{"in progress", "in progress", "in_progress"},
		{"working", "working", "in_progress"},
		{"active", "active", "in_progress"},
		{"inprogress", "inprogress", "in_progress"},

		// Review variants
		{"review", "review", "review"},
		{"Review", "Review", "review"},
		{"needs review", "needs review", "review"},
		{"awaiting review", "awaiting review", "review"},

		// Completed variants
		{"done", "done", "completed"},
		{"DONE", "DONE", "completed"},
		{"completed", "completed", "completed"},
		{"finished", "finished", "completed"},
		{"closed", "closed", "completed"},

		// Blocked variants
		{"blocked", "blocked", "blocked"},
		{"waiting", "waiting", "blocked"},

		// Cancelled variants
		{"cancelled", "cancelled", "cancelled"},
		{"canceled", "canceled", "cancelled"},
		{"abandoned", "abandoned", "cancelled"},

		// Custom statuses (should return lowercase)
		{"custom status", "Custom Status", "custom status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeStatus(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeStatusToTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty/whitespace
		{"empty string", "", "Todo"},
		{"whitespace", "   ", "Todo"},

		// Todo variants
		{"todo", "todo", "Todo"},
		{"TODO", "TODO", "Todo"},
		{"pending", "pending", "Todo"},

		// In Progress variants
		{"in_progress", "in_progress", "In Progress"},
		{"in-progress", "in-progress", "In Progress"},
		{"in progress", "in progress", "In Progress"},
		{"working", "working", "In Progress"},

		// Review variants
		{"review", "review", "Review"},
		{"Review", "Review", "Review"},

		// Completed variants (should map to "Done")
		{"done", "done", "Done"},
		{"DONE", "DONE", "Done"},
		{"completed", "completed", "Done"},
		{"finished", "finished", "Done"},

		// Blocked variants
		{"blocked", "blocked", "Blocked"},

		// Cancelled variants
		{"cancelled", "cancelled", "Cancelled"},
		{"canceled", "canceled", "Cancelled"},

		// Custom statuses (should capitalize)
		{"custom status", "custom status", "Custom Status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeStatusToTitleCase(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeStatusToTitleCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty/whitespace
		{"empty string", "", "medium"},
		{"whitespace", "   ", "medium"},

		// Standard priorities
		{"low", "low", "low"},
		{"Low", "Low", "low"},
		{"LOW", "LOW", "low"},
		{"medium", "medium", "medium"},
		{"Medium", "Medium", "medium"},
		{"MEDIUM", "MEDIUM", "medium"},
		{"high", "high", "high"},
		{"High", "High", "high"},
		{"HIGH", "HIGH", "high"},
		{"critical", "critical", "critical"},
		{"Critical", "Critical", "critical"},
		{"CRITICAL", "CRITICAL", "critical"},

		// Variants
		{"lowest", "lowest", "low"},
		{"highest", "highest", "critical"},
		{"urgent", "urgent", "critical"},
		{"normal", "normal", "medium"},
		{"standard", "standard", "medium"},

		// Custom priorities (should return lowercase)
		{"custom priority", "Custom Priority", "custom priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePriority(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizePriority(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsPendingStatusNormalized(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"todo", "todo", true},
		{"Todo", "Todo", true},
		{"pending", "pending", true},
		{"in_progress", "in_progress", false},
		{"done", "done", false},
		{"review", "review", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPendingStatusNormalized(tt.input)
			if got != tt.expected {
				t.Errorf("IsPendingStatusNormalized(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsCompletedStatusNormalized(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"done", "done", true},
		{"Done", "Done", true},
		{"completed", "completed", true},
		{"cancelled", "cancelled", true},
		{"todo", "todo", false},
		{"in_progress", "in_progress", false},
		{"review", "review", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCompletedStatusNormalized(tt.input)
			if got != tt.expected {
				t.Errorf("IsCompletedStatusNormalized(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsActiveStatusNormalized(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"todo", "todo", true},
		{"in_progress", "in_progress", true},
		{"review", "review", true},
		{"blocked", "blocked", true},
		{"done", "done", false},
		{"completed", "completed", false},
		{"cancelled", "cancelled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsActiveStatusNormalized(tt.input)
			if got != tt.expected {
				t.Errorf("IsActiveStatusNormalized(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsReviewStatusNormalized(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"review", "review", true},
		{"Review", "Review", true},
		{"todo", "todo", false},
		{"done", "done", false},
		{"in_progress", "in_progress", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsReviewStatusNormalized(tt.input)
			if got != tt.expected {
				t.Errorf("IsReviewStatusNormalized(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
