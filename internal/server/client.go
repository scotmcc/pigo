package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client sends commands to a running pigo server.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a client pointed at the given server address.
func NewClient(host string, port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Send dispatches a command to the server and returns the response.
func (c *Client) Send(command string, args map[string]any) (*Response, error) {
	body, err := json.Marshal(Request{Command: command, Args: args})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.http.Post(c.baseURL+"/command", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("send command: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// IsRunning checks if a pigo server is reachable at the configured address.
func (c *Client) IsRunning() bool {
	resp, err := c.http.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
