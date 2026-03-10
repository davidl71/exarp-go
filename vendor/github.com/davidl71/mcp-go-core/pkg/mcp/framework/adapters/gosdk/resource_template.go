package gosdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResourceTemplateHandlerString returns an SDK ResourceHandler for a URI template with one string parameter.
// uriPrefix is the prefix to match (e.g. "wisdom://advisor/").
// extract gets the param from the URI suffix (e.g. after trimming prefix); it may return an error for invalid suffix.
// handler is called with (ctx, fullURI, param) and returns (body, mimeType, error).
func ResourceTemplateHandlerString(
	uriPrefix string,
	extract func(suffix string) (string, error),
	handler func(ctx context.Context, uri string, param string) ([]byte, string, error),
) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		if req.Params == nil || req.Params.URI == "" {
			return nil, fmt.Errorf("resource URI is required")
		}
		uri := req.Params.URI
		if !strings.HasPrefix(uri, uriPrefix) {
			return nil, fmt.Errorf("invalid URI format: expected prefix %q, got %q", uriPrefix, uri)
		}
		param, err := extract(strings.TrimPrefix(uri, uriPrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to extract parameter from URI %q: %w", uri, err)
		}
		body, mimeType, err := handler(ctx, uri, param)
		if err != nil {
			return nil, err
		}
		if mimeType == "" {
			mimeType = "application/json"
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, MIMEType: mimeType, Text: string(body)},
			},
		}, nil
	}
}

// ResourceTemplateHandlerInt returns an SDK ResourceHandler for a URI template with one int parameter.
// If extract returns an error, defaultVal is used as the param (so a default can be used for empty or invalid suffix).
// handler is called with (ctx, fullURI, param) and returns (body, mimeType, error).
func ResourceTemplateHandlerInt(
	uriPrefix string,
	extract func(suffix string) (int, error),
	defaultVal int,
	handler func(ctx context.Context, uri string, param int) ([]byte, string, error),
) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		if req.Params == nil || req.Params.URI == "" {
			return nil, fmt.Errorf("resource URI is required")
		}
		uri := req.Params.URI
		if !strings.HasPrefix(uri, uriPrefix) {
			return nil, fmt.Errorf("invalid URI format: expected prefix %q, got %q", uriPrefix, uri)
		}
		param, err := extract(strings.TrimPrefix(uri, uriPrefix))
		if err != nil {
			param = defaultVal
		}
		body, mimeType, err := handler(ctx, uri, param)
		if err != nil {
			return nil, err
		}
		if mimeType == "" {
			mimeType = "application/json"
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, MIMEType: mimeType, Text: string(body)},
			},
		}, nil
	}
}
