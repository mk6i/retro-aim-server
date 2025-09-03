package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mk6i/retro-aim-server/state"
)

// SessionRetriever provides methods to retrieve OSCAR sessions.
type SessionRetriever interface {
	AllSessions() []*state.Session
	RetrieveSession(screenName state.IdentScreenName) *state.Session
}

// CommonHandler provides shared utilities for all Web API handlers.
type CommonHandler struct {
	Logger *slog.Logger
}

// SendError sends an error response in Web AIM API format.
func SendError(w http.ResponseWriter, statusCode int, message string) {
	resp := struct {
		Response struct {
			StatusCode int    `json:"statusCode"`
			StatusText string `json:"statusText"`
		} `json:"response"`
	}{}

	resp.Response.StatusCode = statusCode
	resp.Response.StatusText = message

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// SendJSON sends a JSON response.
func SendJSON(w http.ResponseWriter, data interface{}, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		if logger != nil {
			logger.Error("failed to encode response", "err", err.Error())
		}
	}
}

// SendJSONP sends a JSONP response with the specified callback.
func SendJSONP(w http.ResponseWriter, callback string, data interface{}, logger *slog.Logger) {
	// Validate callback to prevent XSS
	if !IsValidCallback(callback) {
		SendError(w, http.StatusBadRequest, "invalid callback parameter")
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		if logger != nil {
			logger.Error("failed to marshal response", "err", err.Error())
		}
		SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(callback))
	w.Write([]byte("("))
	w.Write(jsonData)
	w.Write([]byte(");"))
}

// IsValidCallback validates a JSONP callback name to prevent XSS.
func IsValidCallback(callback string) bool {
	if len(callback) == 0 || len(callback) > 100 {
		return false
	}

	// Allow alphanumeric, underscore, dollar sign, and dot (for namespace)
	for _, r := range callback {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '$' || r == '.') {
			return false
		}
	}

	return true
}
