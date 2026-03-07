package tools

import "testing"

func TestFormatCodebaseSummary(t *testing.T) {
	tests := []struct {
		name    string
		metrics map[string]interface{}
		want    string
	}{
		{
			name: "multilang summary",
			metrics: map[string]interface{}{
				"total_files":  42,
				"go_files":     10,
				"python_files": 7,
				"cpp_files":    3,
				"rust_files":   0,
			},
			want: "42 files (Go: 10, Python: 7, C/C++: 3)",
		},
		{
			name: "single language summary",
			metrics: map[string]interface{}{
				"total_files": 12,
				"go_files":    12,
			},
			want: "12 files (Go: 12)",
		},
		{
			name: "no tracked language summary",
			metrics: map[string]interface{}{
				"total_files": 5,
			},
			want: "5 files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCodebaseSummary(tt.metrics); got != tt.want {
				t.Fatalf("formatCodebaseSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
