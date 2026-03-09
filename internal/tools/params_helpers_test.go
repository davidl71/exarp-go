package tools

import "testing"

func TestDefaultReportOutputPath(t *testing.T) {
	projectRoot := "/tmp/exarp-go"

	tests := []struct {
		name            string
		defaultFilename string
		params          map[string]interface{}
		want            string
	}{
		{
			name:            "generated analysis defaults to out",
			defaultFilename: "TASK_ANALYSIS_NOISE.json",
			params:          map[string]interface{}{},
			want:            "/tmp/exarp-go/out/TASK_ANALYSIS_NOISE.json",
		},
		{
			name:            "project overview defaults to out",
			defaultFilename: "PROJECT_OVERVIEW.md",
			params:          map[string]interface{}{},
			want:            "/tmp/exarp-go/out/PROJECT_OVERVIEW.md",
		},
		{
			name:            "user facing doc stays in docs",
			defaultFilename: "PRD.md",
			params:          map[string]interface{}{},
			want:            "/tmp/exarp-go/docs/PRD.md",
		},
		{
			name:            "explicit output path wins",
			defaultFilename: "TASK_ANALYSIS_NOISE.json",
			params:          map[string]interface{}{"output_path": "custom/report.json"},
			want:            "custom/report.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultReportOutputPath(projectRoot, tt.defaultFilename, tt.params); got != tt.want {
				t.Fatalf("DefaultReportOutputPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
