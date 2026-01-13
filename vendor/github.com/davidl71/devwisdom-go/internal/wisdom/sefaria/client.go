package sefaria

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides access to the Sefaria API
type Client struct {
	httpClient *http.Client
	baseURL    string
	cache      *Cache
}

// NewClient creates a new Sefaria API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    "https://www.sefaria.org/api",
		cache:      NewCache(),
	}
}

// GetText retrieves text from Sefaria API
// book: Sefaria book ID (e.g., "Pirkei_Avot", "Proverbs")
// chapter: Chapter number (0 for full book)
// verse: Verse number (0 for full chapter)
func (c *Client) GetText(ctx context.Context, book string, chapter, verse int) (*TextResponse, error) {
	// Build cache key
	cacheKey := c.buildCacheKey(book, chapter, verse)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Build API endpoint URL
	endpoint := c.buildEndpoint(book, chapter, verse)
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sefaria API request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "devwisdom-go/1.0")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sefaria API request failed for %q: %w", endpoint, err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Sefaria API returned status %d for %q: %s", resp.StatusCode, endpoint, string(body))
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Sefaria API response: %w", err)
	}

	// Parse JSON response
	var textResp TextResponse
	if err := json.Unmarshal(body, &textResp); err != nil {
		return nil, fmt.Errorf("failed to parse Sefaria API response (invalid JSON): %w", err)
	}

	// Validate response has required fields
	if len(textResp.Text) == 0 && len(textResp.He) == 0 {
		return nil, fmt.Errorf("Sefaria API response has no text content for %q", endpoint)
	}

	// Cache the response
	c.cache.Set(cacheKey, &textResp)

	return &textResp, nil
}

// GetTextBySourceID retrieves text using our internal source ID
// Maps source IDs (pirkei_avot, proverbs, etc.) to Sefaria book IDs
func (c *Client) GetTextBySourceID(ctx context.Context, sourceID string, chapter, verse int) (*TextResponse, error) {
	book, exists := BookMapping[sourceID]
	if !exists {
		return nil, fmt.Errorf("unknown Sefaria source ID %q (available: %v)", sourceID, getBookMappingKeys())
	}

	return c.GetText(ctx, book, chapter, verse)
}

// buildEndpoint constructs the API endpoint path
func (c *Client) buildEndpoint(book string, chapter, verse int) string {
	if chapter == 0 {
		// Full book
		return fmt.Sprintf("texts/%s", book)
	} else if verse == 0 {
		// Full chapter
		return fmt.Sprintf("texts/%s.%d", book, chapter)
	} else {
		// Specific verse
		return fmt.Sprintf("texts/%s.%d.%d", book, chapter, verse)
	}
}

// buildCacheKey creates a cache key for the request
func (c *Client) buildCacheKey(book string, chapter, verse int) string {
	return fmt.Sprintf("sefaria:%s:%d:%d", book, chapter, verse)
}

// getBookMappingKeys returns all available book mapping keys
func getBookMappingKeys() []string {
	keys := make([]string, 0, len(BookMapping))
	for k := range BookMapping {
		keys = append(keys, k)
	}
	return keys
}

// CleanupCache removes expired entries from the cache
func (c *Client) CleanupCache() {
	c.cache.Cleanup()
}
