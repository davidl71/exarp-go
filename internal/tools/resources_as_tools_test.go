package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandleReadResourceMatchesTrackedTemplateURI(t *testing.T) {
	templateURI := "stdio://test/template/{id}"
	TrackResource(templateURI, "Template Test", "template resource", "application/json", func(ctx context.Context, uri string) ([]byte, string, error) {
		return []byte(`{"ok":true}`), "application/json", nil
	})

	args, _ := json.Marshal(map[string]string{
		"uri": "stdio://test/template/abc123",
	})

	result, err := handleReadResource(context.Background(), args)
	if err != nil {
		t.Fatalf("handleReadResource: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if payload["uri"] != "stdio://test/template/abc123" {
		t.Fatalf("uri = %v, want actual URI", payload["uri"])
	}
	if payload["mime_type"] != "application/json" {
		t.Fatalf("mime_type = %v, want application/json", payload["mime_type"])
	}
}
