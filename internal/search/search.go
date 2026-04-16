// Package search provides web search via a SearXNG instance.
// Layer-1 package — talks to the SearXNG JSON API and nothing else.
package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client talks to a SearXNG instance.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a search client pointed at a SearXNG instance.
// baseURL should be like "https://searxng.example.com" (no trailing slash).
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Result is a single search hit.
type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"` // snippet/description
	Engine  string `json:"engine"`
}

// searxngResponse is the JSON shape returned by SearXNG's /search endpoint.
type searxngResponse struct {
	Results []Result `json:"results"`
	Query   string   `json:"query"`
}

// Search queries SearXNG and returns results.
func (c *Client) Search(query string, limit int) ([]Result, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("web search not configured — set [search] url in ~/.pigo/config.toml")
	}

	if limit <= 0 {
		limit = 10
	}

	params := url.Values{
		"q":      {query},
		"format": {"json"},
	}

	resp, err := c.http.Get(c.baseURL + "/search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search returned %d: %s", resp.StatusCode, string(body))
	}

	var result searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	results := result.Results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
