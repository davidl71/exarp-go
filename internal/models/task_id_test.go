package models

import "testing"

func TestGenerateTaskIDMonotonicAndUnique(t *testing.T) {
	const count = 1000
	seen := make(map[string]bool, count)
	var last string
	for i := 0; i < count; i++ {
		id := GenerateTaskID()
		if !IsValidTaskID(id) {
			t.Fatalf("GenerateTaskID() produced invalid id %q", id)
		}
		if seen[id] {
			t.Fatalf("GenerateTaskID() produced duplicate id %q", id)
		}
		seen[id] = true
		if last != "" && id <= last {
			t.Fatalf("GenerateTaskID() not monotonic: prev=%q current=%q", last, id)
		}
		last = id
	}
}
