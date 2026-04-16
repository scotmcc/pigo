// Package ollama provides an HTTP client for the Ollama embedding API.
// Layer-1 package — talks to Ollama and nothing else.
package ollama

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

// Client talks to the Ollama /api/embed endpoint.
type Client struct {
	endpoint string
	model    string
	http     *http.Client
}

// NewClient creates an Ollama client with the given endpoint and model.
func NewClient(endpoint, model string) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		http:     &http.Client{},
	}
}

// embedRequest is the JSON body sent to /api/embed.
type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// embedResponse is the JSON body returned from /api/embed.
type embedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// Embed returns the embedding vector for the given text.
// The result is a float32 byte slice suitable for SQLite storage.
func (c *Client) Embed(text string) ([]byte, error) {
	body, err := json.Marshal(embedRequest{
		Model: c.model,
		Input: text,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.http.Post(c.endpoint+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return float64sToBytes(result.Embeddings[0]), nil
}

// float64sToBytes converts a float64 slice to a float32 byte slice.
// SQLite-vec expects float32 vectors, but Ollama returns float64.
func float64sToBytes(fs []float64) []byte {
	buf := new(bytes.Buffer)
	for _, f := range fs {
		binary.Write(buf, binary.LittleEndian, float32(math.Float64frombits(math.Float64bits(f))))
	}
	return buf.Bytes()
}
