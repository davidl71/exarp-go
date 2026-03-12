package config

import "testing"

func TestTasksKeybindingsRoundTrip(t *testing.T) {
	cfg := GetDefaults()
	cfg.Tasks.Keybindings["refresh"] = []string{"ctrl+r"}

	pb, err := ToProtobuf(cfg)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	roundTrip, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	got := roundTrip.Tasks.Keybindings["refresh"]
	if len(got) != 1 || got[0] != "ctrl+r" {
		t.Fatalf("refresh keybinding = %v, want [ctrl+r]", got)
	}
}

func TestValidateConfigRejectsEmptyKeybinding(t *testing.T) {
	cfg := GetDefaults()
	cfg.Tasks.Keybindings = map[string][]string{
		"quit": {""},
	}

	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("ValidateConfig() should reject empty keybindings")
	}
}
