package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mk6i/retro-aim-server/state"
)

// AuthHandler handles Web AIM API authentication endpoints.
type AuthHandler struct {
	UserManager UserManager
	TokenStore  TokenStore
	Logger      *slog.Logger
}

// UserManager defines methods for user authentication.
type UserManager interface {
	// AuthenticateUser verifies username and password
	AuthenticateUser(username, password string) (*state.User, error)
	// FindUserByScreenName finds a user by their screen name
	FindUserByScreenName(screenName state.IdentScreenName) (*state.User, error)
}

// TokenStore manages authentication tokens.
type TokenStore interface {
	// StoreToken saves an authentication token for a user
	StoreToken(token string, screenName state.IdentScreenName, expiresAt time.Time) error
	// ValidateToken checks if a token is valid and returns the associated screen name
	ValidateToken(token string) (state.IdentScreenName, error)
	// DeleteToken removes a token
	DeleteToken(token string) error
}

// ClientLoginRequest represents the request body for clientLogin.
type ClientLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DevID    string `json:"devId"`
}

// ClientLoginResponse represents the response for clientLogin endpoint.
type ClientLoginResponse struct {
	Response struct {
		StatusCode int    `json:"statusCode"`
		StatusText string `json:"statusText"`
		Data       struct {
			Token struct {
				A string `json:"a"`
			} `json:"token"`
			LoginID        string `json:"loginId"`
			ScreenName     string `json:"screenName"`
			SessionSecret  string `json:"sessionSecret"`
			HostTime       int64  `json:"hostTime"`
			TokenExpiresIn int    `json:"tokenExpiresIn"`
		} `json:"data"`
	} `json:"response"`
}

// ClientLoginXMLResponse represents the XML response for clientLogin endpoint.
type ClientLoginXMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	StatusCode int      `xml:"statusCode"`
	StatusText string   `xml:"statusText"`
	Data       struct {
		Token struct {
			A string `xml:"a"`
		} `xml:"token"`
		LoginID        string `xml:"loginId"`
		ScreenName     string `xml:"screenName"`
		SessionSecret  string `xml:"sessionSecret"`
		HostTime       int64  `xml:"hostTime"`
		TokenExpiresIn int    `xml:"tokenExpiresIn"`
	} `xml:"data"`
}

// ClientLogin handles POST /auth/clientLogin requests.
// This endpoint authenticates users and returns an authentication token.
func (h *AuthHandler) ClientLogin(w http.ResponseWriter, r *http.Request) {
	var username, password, devID, format string

	// Check Content-Type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// Parse JSON body
		var req ClientLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("failed to parse JSON clientLogin request", "error", err)
			h.sendError(w, http.StatusBadRequest, "invalid JSON format")
			return
		}
		username = req.Username
		password = req.Password
		devID = req.DevID
		format = "json" // default to JSON for JSON requests
	} else {
		// Parse form-encoded or URL parameters
		if err := r.ParseForm(); err != nil {
			h.Logger.Error("failed to parse form data", "error", err)
			h.sendError(w, http.StatusBadRequest, "invalid form data")
			return
		}

		// Try form values first, then fall back to query parameters
		username = r.FormValue("s")
		if username == "" {
			username = r.FormValue("username")
		}
		password = r.FormValue("pwd")
		if password == "" {
			password = r.FormValue("password")
		}
		devID = r.FormValue("devId")
		format = r.FormValue("f") // Get format parameter
		if format == "" {
			format = "json" // default to JSON
		}

		h.Logger.Debug("form-encoded login attempt",
			"username", username,
			"has_password", password != "",
			"devId", devID,
			"format", format,
			"form", r.Form)
	}

	// Validate required fields
	if username == "" || password == "" {
		h.sendError(w, http.StatusBadRequest, "username and password required")
		return
	}

	// Authenticate user
	user, err := h.UserManager.AuthenticateUser(username, password)
	if err != nil {
		h.Logger.Warn("authentication failed",
			"username", username,
			"error", err)
		h.sendError(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	// Generate authentication token
	token, err := h.generateToken()
	if err != nil {
		h.Logger.Error("failed to generate token", "error", err)
		h.sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Store token with 24 hour expiry
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := h.TokenStore.StoreToken(token, user.IdentScreenName, expiresAt); err != nil {
		h.Logger.Error("failed to store token", "error", err)
		h.sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Generate session secret (for signing subsequent requests)
	sessionSecret, err := h.generateToken()
	if err != nil {
		h.Logger.Error("failed to generate session secret", "error", err)
		h.sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Build and send response based on requested format
	if format == "xml" {
		// Build XML response
		resp := ClientLoginXMLResponse{}
		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Data.Token.A = token
		resp.Data.LoginID = string(user.DisplayScreenName)
		resp.Data.ScreenName = string(user.DisplayScreenName)
		resp.Data.SessionSecret = sessionSecret
		resp.Data.HostTime = time.Now().Unix()
		resp.Data.TokenExpiresIn = 86400 // 24 hours in seconds

		// Send XML response
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
		if err := xml.NewEncoder(w).Encode(resp); err != nil {
			h.Logger.Error("failed to encode XML response", "error", err)
		}
	} else {
		// Build JSON response
		resp := ClientLoginResponse{}
		resp.Response.StatusCode = 200
		resp.Response.StatusText = "OK"
		resp.Response.Data.Token.A = token
		resp.Response.Data.LoginID = string(user.DisplayScreenName)
		resp.Response.Data.ScreenName = string(user.DisplayScreenName)
		resp.Response.Data.SessionSecret = sessionSecret
		resp.Response.Data.HostTime = time.Now().Unix()
		resp.Response.Data.TokenExpiresIn = 86400 // 24 hours in seconds

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.Logger.Error("failed to encode JSON response", "error", err)
		}
	}

	h.Logger.Info("user authenticated successfully",
		"username", username,
		"screenName", user.DisplayScreenName,
		"format", format)
}

// generateToken generates a secure random token.
func (h *AuthHandler) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// sendError sends an error response in the requested format.
func (h *AuthHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	// Default to JSON for error responses
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

// sendXMLError sends an XML error response.
func (h *AuthHandler) sendXMLError(w http.ResponseWriter, statusCode int, message string) {
	resp := struct {
		XMLName    xml.Name `xml:"response"`
		StatusCode int      `xml:"statusCode"`
		StatusText string   `xml:"statusText"`
	}{
		StatusCode: statusCode,
		StatusText: message,
	}

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(statusCode)
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	xml.NewEncoder(w).Encode(resp)
}
