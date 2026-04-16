// Package server provides the pigo HTTP and TCP servers.
// It accepts JSON command requests and dispatches them through the command registry.
package server

// Request is the JSON body sent to the server (HTTP or pipe).
//
//	{ "guid": "abc-123", "command": "vault.search", "args": { "q": "..." } }
//
// GUID is optional on HTTP (server generates one). Required on pipe for routing.
type Request struct {
	GUID    string         `json:"guid,omitempty"`
	Type    string         `json:"type,omitempty"` // "register_pipe" for pipe setup
	Command string         `json:"command,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
}

// Response is the JSON body returned by the server.
//
// Sync (HTTP):
//
//	{ "guid": "abc-123", "success": true, "data": ... }
//
// Async ack:
//
//	{ "guid": "abc-123", "status": "accepted" }
//
// Async update:
//
//	{ "guid": "abc-123", "status": "update", "message": "Processing..." }
//
// Async done:
//
//	{ "guid": "abc-123", "status": "done", "data": {...} }
type Response struct {
	GUID    string `json:"guid,omitempty"`
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"` // accepted, update, done, error
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}
