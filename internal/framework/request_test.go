package framework

import (
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/proto"
)

func TestParseRequest_PrefersJSONObjectOverProtoBinary(t *testing.T) {
	req, params, err := ParseRequest(json.RawMessage(`{"action":"validate","dry_run":true}`), func() *proto.TaskAnalysisRequest {
		return &proto.TaskAnalysisRequest{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req != nil {
		t.Fatalf("expected JSON path, got proto request: %#v", req)
	}
	if got := params["action"]; got != "validate" {
		t.Fatalf("action = %#v, want validate", got)
	}
	if got := params["dry_run"]; got != true {
		t.Fatalf("dry_run = %#v, want true", got)
	}
}

func TestParseRequest_EmptyArgsReturnEmptyParams(t *testing.T) {
	req, params, err := ParseRequest(json.RawMessage{}, func() *proto.TaskAnalysisRequest {
		return &proto.TaskAnalysisRequest{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req != nil {
		t.Fatalf("expected nil proto request, got %#v", req)
	}
	if params == nil {
		t.Fatal("expected params map, got nil")
	}
	if len(params) != 0 {
		t.Fatalf("expected empty params, got %#v", params)
	}
}
