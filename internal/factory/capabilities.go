package factory

import (
	"context"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const completionLimit = 64

// newCompletionHandler returns a basic completion handler that supplies tool,
// prompt, and resource candidates based on the argument name.
func newCompletionHandler() func(context.Context, *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
		if req == nil || req.Params == nil {
			return &mcp.CompleteResult{}, nil
		}

		candidates := completionCandidates(req)
		if len(candidates) == 0 {
			return &mcp.CompleteResult{}, nil
		}

		candidates = applyPrefixFilter(candidates, req.Params.Argument.Value)
		if len(candidates) > completionLimit {
			candidates = candidates[:completionLimit]
		}

		return &mcp.CompleteResult{
			Completion: mcp.CompletionResultDetails{
				Values: candidates,
				Total:  len(candidates),
			},
		}, nil
	}
}

func completionCandidates(req *mcp.CompleteRequest) []string {
	name := strings.ToLower(req.Params.Argument.Name)
	switch {
	case name == "tool" || name == "tool_name":
		return tools.ListToolNames()
	case strings.Contains(name, "prompt"):
		return prompts.ListAllPromptNames()
	case strings.Contains(name, "uri"), strings.Contains(name, "resource"), refTargetIsResource(req):
		return tools.ListTrackedResourceURIs()
	default:
		return nil
	}
}

func refTargetIsResource(req *mcp.CompleteRequest) bool {
	if req.Params.Ref == nil {
		return false
	}
	return req.Params.Ref.Type == "ref/resource"
}

func applyPrefixFilter(values []string, prefix string) []string {
	if prefix == "" {
		return values
	}

	lowerPrefix := strings.ToLower(prefix)
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.HasPrefix(strings.ToLower(value), lowerPrefix) {
			filtered = append(filtered, value)
		}
	}

	sort.Strings(filtered)
	return filtered
}
