package tools

import "testing"

func TestMapToStructViaJSON(t *testing.T) {
	t.Parallel()

	type healthArgs struct {
		Action string `json:"action"`
		Path   string `json:"path"`
	}

	params := map[string]interface{}{
		"action": "docs",
		"path":   "README.md",
		"extra":  1,
	}

	var dst healthArgs
	if err := MapToStructViaJSON(params, &dst); err != nil {
		t.Fatal(err)
	}

	if dst.Action != "docs" || dst.Path != "README.md" {
		t.Fatalf("%+v", dst)
	}
}
