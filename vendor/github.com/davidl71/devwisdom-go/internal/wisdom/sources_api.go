package wisdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APISourceLoader handles loading sources from remote APIs
type APISourceLoader struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

// NewAPISourceLoader creates a new API source loader
func NewAPISourceLoader(baseURL string, timeout time.Duration) *APISourceLoader {
	return &APISourceLoader{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		timeout: timeout,
	}
}

// LoadSource loads a source from an API endpoint
func (al *APISourceLoader) LoadSource(ctx context.Context, endpoint string) (*SourceConfig, error) {
	url := fmt.Sprintf("%s/%s", al.baseURL, endpoint)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "devwisdom-go/1.0")

	// Make request with timeout
	resp, err := al.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d for source %q: check API endpoint and source ID", resp.StatusCode, endpoint)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON
	var config SourceConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &config, nil
}

// LoadSourceWithRetry loads a source with retry logic
func (al *APISourceLoader) LoadSourceWithRetry(ctx context.Context, endpoint string, maxRetries int) (*SourceConfig, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			backoff := time.Duration(i) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		config, err := al.LoadSource(ctx, endpoint)
		if err == nil {
			return config, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed to load source %q after %d retries: %w", endpoint, maxRetries, lastErr)
}

// LoadSourceWithTimeout loads a source with a custom timeout
func (al *APISourceLoader) LoadSourceWithTimeout(endpoint string, timeout time.Duration) (*SourceConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return al.LoadSource(ctx, endpoint)
}
