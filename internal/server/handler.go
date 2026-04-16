package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/scotmcc/pigo/internal/commands"
)

// handleCommand is the HTTP handler for POST /command.
// It decodes the JSON request, dispatches to the registry, and returns JSON.
// HTTP is always sync — async commands should be sent via the pipe.
func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.Command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}

	guid := req.GUID
	if guid == "" {
		guid = uuid.New().String()
	}

	data, err := commands.Dispatch(req.Command, req.Args, commands.NoOpSend)
	if err != nil {
		writeJSON(w, http.StatusOK, Response{GUID: guid, Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{GUID: guid, Success: true, Data: data})
}

// handleHealth is a simple health check endpoint.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, Response{Success: false, Error: msg})
}
