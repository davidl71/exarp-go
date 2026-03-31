package database

import "testing"

func TestDefaultProjectIDFromEnv(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("PROJECT_ROOT", "")
		if got := DefaultProjectIDFromEnv(); got != "default" {
			t.Fatalf("got %q, want default", got)
		}
	})
	t.Run("basename of path", func(t *testing.T) {
		t.Setenv("PROJECT_ROOT", "/Users/dev/Projects/Trading/Aether")
		if got := DefaultProjectIDFromEnv(); got != "Aether" {
			t.Fatalf("got %q, want Aether", got)
		}
	})
}
