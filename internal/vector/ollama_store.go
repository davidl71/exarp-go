// vector/ollama_store.go — OllamaStore: VectorStore backed by chromem-go + Ollama embeddings.
package vector

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	chromem "github.com/philippgille/chromem-go"
)

// OllamaStore is an in-memory vector index that uses Ollama for embeddings.
// A new instance must be created for each search operation (no persistence).
type OllamaStore struct {
	collection *chromem.Collection
	ollamaURL  string // base URL without /api, e.g. "http://localhost:11434"
}

// NewOllamaStore creates an in-memory OllamaStore using the given Ollama base URL and
// embedding model. Use empty strings to accept defaults ("http://localhost:11434",
// "nomic-embed-text"). Returns an error only if the chromem-go collection cannot be created.
func NewOllamaStore(ollamaBaseURL, model string) (*OllamaStore, error) {
	if ollamaBaseURL == "" {
		ollamaBaseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}

	// chromem-go expects the /api suffix in the Ollama base URL.
	apiURL := ollamaBaseURL + "/api"
	embeddingFunc := chromem.NewEmbeddingFuncOllama(model, apiURL)

	db := chromem.NewDB()
	collection, err := db.CreateCollection("search", nil, embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("vector: create collection: %w", err)
	}

	return &OllamaStore{
		collection: collection,
		ollamaURL:  ollamaBaseURL,
	}, nil
}

// Available does a lightweight HTTP check against the Ollama /api/tags endpoint.
// Uses a 2-second timeout to stay non-blocking.
func (s *OllamaStore) Available() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(s.ollamaURL + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// AddAll upserts all documents into the in-memory index.
// Embeddings are generated concurrently using Ollama.
func (s *OllamaStore) AddAll(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	chromeDocs := make([]chromem.Document, len(docs))
	for i, d := range docs {
		chromeDocs[i] = chromem.Document{ID: d.ID, Content: d.Text}
	}

	if err := s.collection.AddDocuments(ctx, chromeDocs, runtime.NumCPU()); err != nil {
		return fmt.Errorf("vector: add documents: %w", err)
	}

	return nil
}

// Search returns up to n results ranked by cosine similarity to the query.
func (s *OllamaStore) Search(ctx context.Context, query string, n int) ([]SearchResult, error) {
	if n <= 0 {
		n = 10
	}

	// chromem-go panics if nResults > collection size, so cap it.
	count := s.collection.Count()
	if n > count {
		n = count
	}
	if n == 0 {
		return nil, nil
	}

	results, err := s.collection.Query(ctx, query, n, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("vector: query: %w", err)
	}

	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{ID: r.ID, Similarity: r.Similarity}
	}

	return out, nil
}
