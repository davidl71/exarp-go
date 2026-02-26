package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestWrapHandler_HappyPath(t *testing.T) {
	handler := WrapHandler(
		"test_tool",
		func(args json.RawMessage) (any, map[string]interface{}, error) {
			var params map[string]interface{}
			if err := json.Unmarshal(args, &params); err != nil {
				return nil, nil, err
			}
			return nil, params, nil
		},
		func(req any) map[string]interface{} {
			return map[string]interface{}{"from_proto": true}
		},
		map[string]interface{}{"action": "default"},
		func(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
			return []framework.TextContent{{Type: "text", Text: fmt.Sprintf("action=%s", params["action"])}}, nil
		},
	)

	result, err := handler(context.Background(), json.RawMessage(`{"action":"run"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Text != "action=run" {
		t.Errorf("got %v, want action=run", result)
	}
}

func TestWrapHandler_ProtoPath(t *testing.T) {
	type fakeProto struct{ Value string }

	handler := WrapHandler(
		"test_tool",
		func(args json.RawMessage) (any, map[string]interface{}, error) {
			return &fakeProto{Value: "hello"}, nil, nil
		},
		func(req any) map[string]interface{} {
			fp := req.(*fakeProto)
			return map[string]interface{}{"value": fp.Value}
		},
		map[string]interface{}{"mode": "auto"},
		func(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
			return []framework.TextContent{{Type: "text", Text: fmt.Sprintf("value=%s mode=%s", params["value"], params["mode"])}}, nil
		},
	)

	result, err := handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Text != "value=hello mode=auto" {
		t.Errorf("got %q, want value=hello mode=auto", result[0].Text)
	}
}

func TestWrapHandler_ParseError(t *testing.T) {
	handler := WrapHandler(
		"test_tool",
		func(args json.RawMessage) (any, map[string]interface{}, error) {
			return nil, nil, fmt.Errorf("bad input")
		},
		nil,
		nil,
		nil,
	)

	_, err := handler(context.Background(), json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "failed to parse arguments: bad input" {
		t.Errorf("got %q, want 'failed to parse arguments: bad input'", err.Error())
	}
}

func TestWrapHandler_NativeError(t *testing.T) {
	handler := WrapHandler(
		"my_tool",
		func(args json.RawMessage) (any, map[string]interface{}, error) {
			return nil, map[string]interface{}{}, nil
		},
		nil,
		nil,
		func(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
			return nil, fmt.Errorf("something broke")
		},
	)

	_, err := handler(context.Background(), json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "my_tool failed: something broke" {
		t.Errorf("got %q, want 'my_tool failed: something broke'", err.Error())
	}
}

func TestWrapHandler_NilDefaults(t *testing.T) {
	handler := WrapHandler(
		"test_tool",
		func(args json.RawMessage) (any, map[string]interface{}, error) {
			return "proto", nil, nil
		},
		func(req any) map[string]interface{} {
			return map[string]interface{}{"key": "val"}
		},
		nil,
		func(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
			return []framework.TextContent{{Type: "text", Text: fmt.Sprintf("key=%s", params["key"])}}, nil
		},
	)

	result, err := handler(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Text != "key=val" {
		t.Errorf("got %q, want key=val", result[0].Text)
	}
}
