// vector/store.go — VectorStore interface for semantic search with graceful fallback.
// OllamaStore wraps chromem-go; TextFallbackStore is a no-op used when Ollama is unavailable.
package vector

import "context"

// Document is the unit added to a VectorStore.
type Document struct {
	ID   string
	Text string
}

// SearchResult is returned from VectorStore.Search.
type SearchResult struct {
	ID         string
	Similarity float32
}

// VectorStore is an in-process semantic search index backed by Ollama embeddings.
// Both methods must be safe for concurrent use.
type VectorStore interface {
	// AddAll upserts a batch of documents into the index.
	AddAll(ctx context.Context, docs []Document) error
	// Search returns up to n results ranked by semantic similarity.
	Search(ctx context.Context, query string, n int) ([]SearchResult, error)
	// Available reports whether the embedding backend (Ollama) is reachable.
	Available() bool
}
