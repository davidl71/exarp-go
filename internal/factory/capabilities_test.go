package factory

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCompletionHandlerProvidesToolNames(t *testing.T) {
	handler := newCompletionHandler()
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Argument: mcp.CompleteParamsArgument{Name: "tool"},
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("completion handler returned error: %v", err)
	}

	if got := len(res.Completion.Values); got == 0 {
		t.Fatalf("no completion values returned")
	}

	if !containsValue(res.Completion.Values, "report") {
		t.Errorf("completion values missing expected tool \"report\"; got %v", res.Completion.Values[:3])
	}
}

func TestCompletionHandlerFiltersByPrefix(t *testing.T) {
	handler := newCompletionHandler()
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Argument: mcp.CompleteParamsArgument{
				Name:  "tool",
				Value: "project_",
			},
		},
	}

	res, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("completion handler returned error: %v", err)
	}

	if res.Completion.Total != len(res.Completion.Values) {
		t.Fatalf("expected total %d to match returned values %d", res.Completion.Total, len(res.Completion.Values))
	}

	if len(res.Completion.Values) == 0 {
		t.Fatalf("expected matches for prefix project_")
	}

	if !containsValue(res.Completion.Values, "project_scorecard") {
		t.Errorf("completion values %v missing project_scorecard", res.Completion.Values)
	}
}

func containsValue(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
