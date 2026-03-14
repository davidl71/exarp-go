// vector/text_fallback.go — TextFallbackStore: no-op VectorStore for when Ollama is unavailable.
// Always returns Available()=false so callers fall back to their text-search paths.
package vector

import "context"

// TextFallbackStore is a VectorStore that is never available.
// It is used as a safe default when Ollama is not reachable.
type TextFallbackStore struct{}

func (TextFallbackStore) Available() bool { return false }

func (TextFallbackStore) AddAll(_ context.Context, _ []Document) error { return nil }

func (TextFallbackStore) Search(_ context.Context, _ string, _ int) ([]SearchResult, error) {
	return nil, nil
}
