package tools

import (
	"testing"
)

func TestParseTaskBreakdown(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid json",
			text:    `{"subtasks":[{"name":"A","description":"Do A","complexity":"simple","dependencies":[]},{"name":"B","description":"Do B","complexity":"medium","dependencies":["A"]}]}`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "with prefix text",
			text:    `Here is the breakdown: {"subtasks":[{"name":"Step 1","description":"First","complexity":"simple"}]}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty",
			text:    ``,
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaskBreakdown(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTaskBreakdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && len(got.Subtasks) != tt.wantLen {
				t.Errorf("parseTaskBreakdown() len(Subtasks) = %v, want %v", len(got.Subtasks), tt.wantLen)
			}
		})
	}
}
