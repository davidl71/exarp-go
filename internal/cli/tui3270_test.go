package cli

import (
	"reflect"
	"testing"
)

func TestStripTUI3270DaemonFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "empty",
			args: nil,
			want: []string{},
		},
		{
			name: "strips daemon and pid-file pair",
			args: []string{"tui3270", "Todo", "--daemon", "--pid-file", "/tmp/x.pid", "3270"},
			want: []string{"tui3270", "Todo", "3270"},
		},
		{
			name: "strips equals form",
			args: []string{"tui3270", "--pid-file=/data/p.pid", "-d"},
			want: []string{"tui3270"},
		},
		{
			name: "keeps foreground",
			args: []string{"tui3270", "--foreground", "In Progress"},
			want: []string{"tui3270", "--foreground", "In Progress"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stripTUI3270DaemonFlags(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("stripTUI3270DaemonFlags() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
