package cli

import "testing"

func TestCollectRepeatedStringFlag(t *testing.T) {
	argv := []string{"exarp-go", "task", "create", "Name", "--tag", "a", "--tag=b", "--tag", " c "}
	got := CollectRepeatedStringFlag(argv, "tag")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
	if CollectRepeatedStringFlag([]string{"task", "list", "--tag", "x"}, "tag") == nil {
		t.Fatal("expected one value")
	}
	if CollectRepeatedStringFlag(nil, "tag") != nil {
		t.Fatal("nil argv")
	}
}

func TestMergeTaskTagsFromCSVAndRepeated(t *testing.T) {
	argv := []string{"task", "create", "N", "--tag", "b", "--tag", "c"}
	got := MergeTaskTagsFromCSVAndRepeated("a,b", argv)
	want := []string{"#a", "#b", "#c"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
	// CSV first; repeated --tag deduped against earlier tokens
	got2 := MergeTaskTagsFromCSVAndRepeated("one,two", []string{"--tag", "two", "--tag", "three"})
	want2 := []string{"#one", "#two", "#three"}
	if len(got2) != len(want2) {
		t.Fatalf("got2 %#v want %#v", got2, want2)
	}
	for i := range want2 {
		if got2[i] != want2[i] {
			t.Fatalf("got2 %#v want %#v", got2, want2)
		}
	}
}
