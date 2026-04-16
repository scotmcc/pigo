// Package detect provides runtime checks for external dependencies — whether
// a binary is on PATH, whether a local service is reachable, whether a model
// is pulled, etc. The install flow uses these checks to decide what to ask
// the user to set up.
package detect

import (
	"encoding/json"
	"net/http"
	"time"
)

// OllamaOnPath reports whether the ollama binary is available on PATH.
func OllamaOnPath() bool {
	return OnPath("ollama")
}

// OllamaReachable reports whether an Ollama daemon is responding at endpoint.
// Uses a short timeout so it never blocks the install flow for long.
func OllamaReachable(endpoint string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(endpoint + "/api/version")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// OllamaHasModel reports whether the named model is pulled on the daemon at
// endpoint. Matches both untagged names (e.g. "nomic-embed-text", which Ollama
// stores as "nomic-embed-text:latest") and explicit tags.
func OllamaHasModel(endpoint, name string) (bool, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint + "/api/tags")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, err
	}

	for _, m := range body.Models {
		if m.Name == name || m.Name == name+":latest" {
			return true, nil
		}
	}
	return false, nil
}
