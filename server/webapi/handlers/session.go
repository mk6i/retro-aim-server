package handlers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mk6i/retro-aim-server/server/webapi/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// SessionHandler handles Web AIM API session management endpoints.
type SessionHandler struct {
	SessionManager   *state.WebAPISessionManager
	OSCARAuthService AuthService
	BuddyListService BuddyListService
	TokenStore       TokenStore
	Logger           *slog.Logger
}

// AuthService defines methods needed for authentication.
type AuthService interface {
	BUCPChallenge(ctx context.Context, bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(ctx context.Context, bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error), advertisedHost string) (wire.SNACMessage, error)
	RegisterBOSSession(ctx context.Context, authCookie state.ServerCookie) (*state.Session, error)
}

// BuddyListService defines methods for buddy list operations.
type BuddyListService interface {
	GetBuddyList(ctx context.Context, screenName state.IdentScreenName) ([]BuddyGroup, error)
}

// BuddyGroup represents a group of buddies.
type BuddyGroup struct {
	Name    string  `json:"name"`
	Buddies []Buddy `json:"buddies"`
}

// Buddy represents a buddy in the buddy list.
type Buddy struct {
	AimID     string `json:"aimId"`
	State     string `json:"state"`
	StatusMsg string `json:"statusMsg,omitempty"`
	AwayMsg   string `json:"awayMsg,omitempty"`
	UserType  string `json:"userType"`
}

// StartSessionResponse represents the response for startSession endpoint.
type StartSessionResponse struct {
	Response struct {
		StatusCode int    `json:"statusCode"`
		StatusText string `json:"statusText"`
		Data       struct {
			AimSID          string                 `json:"aimsid"`
			FetchTimeout    int                    `json:"fetchTimeout"`
			TimeToNextFetch int                    `json:"timeToNextFetch"`
			Events          map[string]interface{} `json:"events,omitempty"`
		} `json:"data"`
	} `json:"response"`
}

// StartSessionXMLResponse represents the XML response for startSession endpoint.
type StartSessionXMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	StatusCode int      `xml:"statusCode"`
	StatusText string   `xml:"statusText"`
	Data       struct {
		AimSID          string `xml:"aimsid"`
		FetchTimeout    int    `xml:"fetchTimeout"`
		TimeToNextFetch int    `xml:"timeToNextFetch"`
		MyInfo          *struct {
			AimID     string `xml:"aimId"`
			DisplayID string `xml:"displayId"`
			Buddylist struct {
				Groups *[]BuddyGroup `xml:"group,omitempty"`
			} `xml:"buddylist,omitempty"`
		} `xml:"myInfo,omitempty"`
		Events *struct {
			BuddyList struct {
				Groups *[]BuddyGroup `xml:"group,omitempty"`
			} `xml:"buddylist"`
		} `xml:"events,omitempty"`
	} `xml:"data"`
}

// EndSessionResponse represents the response for endSession endpoint.
type EndSessionResponse struct {
	Response struct {
		StatusCode int    `json:"statusCode"`
		StatusText string `json:"statusText"`
	} `json:"response"`
}

// StartSession handles GET /aim/startSession requests.
func (h *SessionHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get API key info from context (set by auth middleware)
	apiKey, ok := ctx.Value(middleware.ContextKeyAPIKey).(*state.WebAPIKey)
	if !ok {
		h.sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Parse parameters
	params := r.URL.Query()

	// Get authentication token if provided
	authToken := params.Get("a")

	// Get client info
	clientName := params.Get("clientName")
	if clientName == "" {
		clientName = "WebAIM"
	}
	clientVersion := params.Get("clientVersion")
	if clientVersion == "" {
		clientVersion = "1.0"
	}

	// Get events to subscribe to
	eventsParam := params.Get("events")
	var events []string
	if eventsParam != "" {
		events = strings.Split(eventsParam, ",")
	} else {
		// Default events if none specified
		events = []string{"buddylist", "presence", "im"}
	}

	// Get timeout settings
	timeout := 30000 // Default 30 seconds
	if t := params.Get("timeout"); t != "" {
		if val, err := strconv.Atoi(t); err == nil && val > 0 {
			timeout = val * 1000 // Convert to milliseconds
		}
	}

	// Determine screen name from auth token or anonymous
	var screenName state.DisplayScreenName

	if authToken != "" {
		// Validate auth token and get screen name
		if h.TokenStore == nil {
			h.Logger.Error("TokenStore not configured")
			h.sendError(w, http.StatusInternalServerError, "authentication not configured")
			return
		}
		identScreenName, err := h.TokenStore.ValidateToken(authToken)
		if err != nil {
			h.Logger.Warn("invalid authentication token",
				"error", err)
			h.sendError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		// For WebAPI sessions, we can use the IdentScreenName directly as DisplayScreenName
		// since WRAITH handles the display formatting
		screenName = state.DisplayScreenName(identScreenName.String())
		tokenPreview := authToken
		if len(tokenPreview) > 8 {
			tokenPreview = tokenPreview[:8] + "..."
		}
		h.Logger.Info("authenticated session requested",
			"token", tokenPreview,
			"screenName", screenName)
	} else {
		// Anonymous session - generate guest name
		screenName = state.DisplayScreenName("Guest_" + strconv.FormatInt(time.Now().Unix(), 36))
		h.Logger.Info("anonymous session requested",
			"screenName", screenName)
	}

	// Create session
	session, err := h.SessionManager.CreateSession(screenName, apiKey.DevID, events)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to create session", "err", err.Error())
		h.sendError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Store client info
	session.ClientName = clientName
	session.ClientVersion = clientVersion
	session.FetchTimeout = timeout
	session.RemoteAddr = r.RemoteAddr

	// Prepare response
	resp := StartSessionResponse{}
	resp.Response.StatusCode = 200
	resp.Response.StatusText = "OK"
	resp.Response.Data.AimSID = session.AimSID
	resp.Response.Data.FetchTimeout = session.FetchTimeout
	resp.Response.Data.TimeToNextFetch = session.TimeToNextFetch

	// If buddy list event is subscribed, include initial buddy list
	for _, event := range events {
		if event == "buddylist" {
			// TODO: Fetch actual buddy list from service
			resp.Response.Data.Events = make(map[string]interface{})
			resp.Response.Data.Events["buddylist"] = map[string]interface{}{
				"groups": []BuddyGroup{},
			}
			break
		}
	}

	// Check response format
	format := r.URL.Query().Get("f")
	if format == "" {
		format = "json" // default to JSON
	}

	// Send response in requested format
	if format == "xml" {
		// Build XML response
		xmlResp := StartSessionXMLResponse{}
		xmlResp.StatusCode = 200
		xmlResp.StatusText = "OK"
		xmlResp.Data.AimSID = session.AimSID
		xmlResp.Data.FetchTimeout = timeout
		xmlResp.Data.TimeToNextFetch = 500

		// Add myInfo with user data
		xmlResp.Data.MyInfo = &struct {
			AimID     string `xml:"aimId"`
			DisplayID string `xml:"displayId"`
			Buddylist struct {
				Groups *[]BuddyGroup `xml:"group,omitempty"`
			} `xml:"buddylist,omitempty"`
		}{
			AimID:     session.ScreenName.String(),
			DisplayID: session.ScreenName.String(),
		}

		// Add buddy list if requested in myInfo or events
		for _, event := range events {
			if event == "buddylist" || event == "myInfo" {
				// Add to myInfo buddylist
				emptyGroups := []BuddyGroup{}
				xmlResp.Data.MyInfo.Buddylist.Groups = &emptyGroups

				// Also add to events if specifically requested
				if event == "buddylist" {
					if xmlResp.Data.Events == nil {
						xmlResp.Data.Events = &struct {
							BuddyList struct {
								Groups *[]BuddyGroup `xml:"group,omitempty"`
							} `xml:"buddylist"`
						}{}
					}
					xmlResp.Data.Events.BuddyList.Groups = &emptyGroups
				}
				break
			}
		}

		// Send XML response
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")

		// Build complete XML string first
		xmlData, err := xml.Marshal(xmlResp)
		if err != nil {
			h.Logger.Error("failed to marshal XML response", "error", err)
			h.sendError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// Write XML declaration and data as one response
		xmlOutput := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>%s`, xmlData)
		w.Header().Set("Content-Length", strconv.Itoa(len(xmlOutput)))
		fmt.Fprint(w, xmlOutput)
	} else {
		// Send JSON response
		SendJSON(w, resp, h.Logger)
	}

	h.Logger.InfoContext(ctx, "session started",
		"aimsid", session.AimSID,
		"screen_name", screenName,
		"dev_id", apiKey.DevID,
		"events", events,
		"format", format,
	)
}

// EndSession handles GET /aim/endSession requests.
func (h *SessionHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from parameters
	aimsid := r.URL.Query().Get("aimsid")
	if aimsid == "" {
		h.sendError(w, http.StatusBadRequest, "missing aimsid parameter")
		return
	}

	// Get session
	session, err := h.SessionManager.GetSession(aimsid)
	if err != nil {
		if err == state.ErrNoWebAPISession {
			h.sendError(w, http.StatusNotFound, "session not found")
		} else if err == state.ErrWebAPISessionExpired {
			h.sendError(w, http.StatusGone, "session expired")
		} else {
			h.sendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// TODO: Broadcast offline presence to buddies
	// TODO: Clean up OSCAR session if bridged

	// Remove session
	if err := h.SessionManager.RemoveSession(aimsid); err != nil {
		h.Logger.ErrorContext(ctx, "failed to remove session", "err", err.Error())
		h.sendError(w, http.StatusInternalServerError, "failed to end session")
		return
	}

	// Send response
	resp := EndSessionResponse{}
	resp.Response.StatusCode = 200
	resp.Response.StatusText = "OK"

	SendJSON(w, resp, h.Logger)

	h.Logger.InfoContext(ctx, "session ended",
		"aimsid", aimsid,
		"screen_name", session.ScreenName,
	)
}

// sendError is a convenience method that wraps the common SendError function.
func (h *SessionHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	SendError(w, statusCode, message)
}
