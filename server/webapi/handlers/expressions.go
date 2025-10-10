package handlers

import (
	"log/slog"
	"net/http"
)

// ExpressionsHandler handles Web AIM API expressions/buddy icon endpoints.
type ExpressionsHandler struct {
	Logger *slog.Logger
}

// NewExpressionsHandler creates a new ExpressionsHandler.
func NewExpressionsHandler(logger *slog.Logger) *ExpressionsHandler {
	return &ExpressionsHandler{
		Logger: logger,
	}
}

// Get handles GET /expressions/get requests for buddy icons and expressions.
func (h *ExpressionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	format := r.URL.Query().Get("f")
	targetUser := r.URL.Query().Get("t")
	expressionType := r.URL.Query().Get("type")

	h.Logger.Debug("expressions/get request",
		"format", format,
		"target", targetUser,
		"type", expressionType,
	)

	// For redirect format, return a 404 (no buddy icon)
	// In a real implementation, this would redirect to the actual icon URL
	if format == "redirect" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// For other formats, return an empty response indicating no expressions
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": 200,
			"statusText": "OK",
			"data": map[string]interface{}{
				"expressions": []interface{}{},
			},
		},
	}

	// Send response in requested format
	SendResponse(w, r, response, h.Logger)
}
