package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// PresenceHandler handles Web AIM API presence-related endpoints.
type PresenceHandler struct {
	SessionManager   *state.WebAPISessionManager
	SessionRetriever SessionRetriever
	FeedbagRetriever FeedbagRetriever
	BuddyBroadcaster BuddyBroadcaster
	ProfileManager   ProfileManager
	Logger           *slog.Logger
}

// FeedbagRetriever provides methods to retrieve buddy list data.
type FeedbagRetriever interface {
	RetrieveFeedbag(ctx context.Context, screenName state.IdentScreenName) ([]wire.FeedbagItem, error)
	RelationshipsByUser(ctx context.Context, screenName state.IdentScreenName) ([]state.IdentScreenName, error)
}

// BuddyBroadcaster broadcasts buddy presence updates
type BuddyBroadcaster interface {
	BroadcastBuddyArrived(ctx context.Context, sess *state.Session) error
	BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error
}

// ProfileManager manages user profiles
type ProfileManager interface {
	SetProfile(ctx context.Context, screenName state.IdentScreenName, profile string) error
	Profile(ctx context.Context, screenName state.IdentScreenName) (string, error)
}

// PresenceGetResponse represents the response for presence/get endpoint.
type PresenceGetResponse struct {
	Response struct {
		StatusCode int          `json:"statusCode"`
		StatusText string       `json:"statusText"`
		Data       PresenceData `json:"data"`
	} `json:"response"`
}

// PresenceData contains presence information.
type PresenceData struct {
	Groups []BuddyGroupInfo    `json:"groups,omitempty"`
	Users  []BuddyPresenceInfo `json:"users,omitempty"`
}

// BuddyGroupInfo represents a buddy group with its members.
type BuddyGroupInfo struct {
	Name    string              `json:"name"`
	Buddies []BuddyPresenceInfo `json:"buddies"`
}

// BuddyPresenceInfo represents presence information for a buddy.
type BuddyPresenceInfo struct {
	AimID      string `json:"aimId"`
	State      string `json:"state"` // "online", "offline", "away", "idle"
	StatusMsg  string `json:"statusMsg,omitempty"`
	AwayMsg    string `json:"awayMsg,omitempty"`
	IdleTime   int    `json:"idleTime,omitempty"`
	OnlineTime int64  `json:"onlineTime,omitempty"`
	UserType   string `json:"userType"` // "aim", "icq", "admin"
}

// GetPresence handles GET /presence/get requests.
func (h *PresenceHandler) GetPresence(w http.ResponseWriter, r *http.Request) {
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

	// Touch the session
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Check if buddy list is requested
	getBuddyList := r.URL.Query().Get("bl") == "1"

	// Get target users if specified
	targetUsers := r.URL.Query().Get("t")

	// Prepare response
	resp := PresenceGetResponse{}
	resp.Response.StatusCode = 200
	resp.Response.StatusText = "OK"

	if getBuddyList {
		// Retrieve buddy list from feedbag
		groups, err := h.getBuddyListGroups(ctx, session.ScreenName.IdentScreenName())
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to get buddy list", "err", err.Error())
			// Return empty buddy list on error instead of failing
			groups = []BuddyGroupInfo{}
		}
		resp.Response.Data.Groups = groups
	} else if targetUsers != "" {
		// Get presence for specific users
		users := strings.Split(targetUsers, ",")
		presenceList := make([]BuddyPresenceInfo, 0, len(users))

		for _, user := range users {
			user = strings.TrimSpace(user)
			if user == "" {
				continue
			}

			presence := h.getUserPresence(state.NewIdentScreenName(user))
			presenceList = append(presenceList, presence)
		}

		resp.Response.Data.Users = presenceList
	} else {
		// No specific request, return empty data
		resp.Response.Data.Groups = []BuddyGroupInfo{}
	}

	// Check for JSONP callback
	if callback := r.URL.Query().Get("f"); callback != "" {
		SendJSONP(w, callback, resp, h.Logger)
	} else {
		SendJSON(w, resp, h.Logger)
	}

	h.Logger.DebugContext(ctx, "presence retrieved",
		"aimsid", aimsid,
		"buddy_list", getBuddyList,
		"targets", targetUsers,
	)
}

// getBuddyListGroups retrieves the buddy list organized by groups.
func (h *PresenceHandler) getBuddyListGroups(ctx context.Context, screenName state.IdentScreenName) ([]BuddyGroupInfo, error) {
	// Get feedbag items
	items, err := h.FeedbagRetriever.RetrieveFeedbag(ctx, screenName)
	if err != nil {
		return nil, err
	}

	// Organize items into groups
	groupMap := make(map[uint16]*BuddyGroupInfo)
	buddyToGroup := make(map[string]uint16)

	// First pass: identify groups
	for _, item := range items {
		if item.ClassID == wire.FeedbagClassIdGroup {
			name := item.Name
			if name == "" {
				name = "Buddies" // Default group name
			}

			groupMap[item.ItemID] = &BuddyGroupInfo{
				Name:    name,
				Buddies: []BuddyPresenceInfo{},
			}
		}
	}

	// Second pass: add buddies to groups
	for _, item := range items {
		if item.ClassID == wire.FeedbagClassIdBuddy {
			// Get buddy screen name
			buddyName := item.Name
			if buddyName == "" {
				continue
			}

			// Find buddy's group
			groupID := item.GroupID

			buddyToGroup[buddyName] = groupID
		}
	}

	// If no groups exist, create a default one
	if len(groupMap) == 0 {
		groupMap[0] = &BuddyGroupInfo{
			Name:    "Buddies",
			Buddies: []BuddyPresenceInfo{},
		}
	}

	// Add buddies to their groups with presence info
	for buddyName, groupID := range buddyToGroup {
		group, exists := groupMap[groupID]
		if !exists {
			// Put in first available group if group doesn't exist
			for _, g := range groupMap {
				group = g
				break
			}
		}

		// Get presence information for buddy
		presence := h.getUserPresence(state.NewIdentScreenName(buddyName))
		group.Buddies = append(group.Buddies, presence)
	}

	// Convert map to slice
	groups := make([]BuddyGroupInfo, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, *group)
	}

	return groups, nil
}

// getUserPresence gets the current presence state for a user.
func (h *PresenceHandler) getUserPresence(screenName state.IdentScreenName) BuddyPresenceInfo {
	// Default offline presence
	presence := BuddyPresenceInfo{
		AimID:    screenName.String(),
		State:    "offline",
		UserType: "aim",
	}

	// Check if user is online by looking for their OSCAR session
	if session := h.SessionRetriever.RetrieveSession(screenName); session != nil {
		presence.State = "online"

		// Check user status
		statusBitmask := session.UserStatusBitmask()
		if statusBitmask&wire.OServiceUserStatusAway != 0 {
			presence.State = "away"
			// TODO: Get away message from session
		} else if statusBitmask&wire.OServiceUserStatusDND != 0 {
			presence.State = "dnd"
		}

		// Check idle time
		if session.Idle() {
			presence.State = "idle"
			idleTime := time.Since(session.IdleTime())
			presence.IdleTime = int(idleTime.Minutes())
		}

		// Get online time
		presence.OnlineTime = session.SignonTime().Unix()

		// TODO: Get status message from profile
	}

	// Determine user type
	if strings.HasPrefix(screenName.String(), "admin") {
		presence.UserType = "admin"
	} else if isICQScreenName(screenName.String()) {
		presence.UserType = "icq"
	}

	return presence
}

// isICQScreenName checks if a screen name is an ICQ number.
func isICQScreenName(screenName string) bool {
	if len(screenName) == 0 {
		return false
	}
	for _, r := range screenName {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// sendError is a convenience method that wraps the common SendError function.
func (h *PresenceHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	SendError(w, statusCode, message)
}

// SetState handles GET /presence/setState requests to update user's presence state.
func (h *PresenceHandler) SetState(w http.ResponseWriter, r *http.Request) {
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
		h.sendError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Get the requested state
	stateParam := r.URL.Query().Get("state")
	awayMsg := r.URL.Query().Get("awayMsg")

	// Get OSCAR session if available
	oscarSession := session.OSCARSession
	if oscarSession == nil {
		// For web-only sessions, we'll need to track state in the WebAPI session
		// For now, just store in event data
		h.Logger.WarnContext(ctx, "no OSCAR session for presence update", "aimsid", aimsid)

		// Still send success response
		response := map[string]interface{}{
			"response": map[string]interface{}{
				"statusCode": 200,
				"statusText": "OK",
			},
		}
		SendJSON(w, response, h.Logger)
		return
	}

	// Map web state to OSCAR status bits
	var statusBitmask uint32
	switch stateParam {
	case "online":
		statusBitmask = 0x0000 // Clear all status bits
		oscarSession.SetAwayMessage("")
	case "away":
		statusBitmask = wire.OServiceUserStatusAway
		if awayMsg != "" {
			oscarSession.SetAwayMessage(awayMsg)
		}
	case "invisible":
		statusBitmask = wire.OServiceUserStatusInvisible
	case "dnd":
		statusBitmask = wire.OServiceUserStatusDND
	default:
		h.sendError(w, http.StatusBadRequest, "invalid state parameter")
		return
	}

	// Update OSCAR session status
	oscarSession.SetUserStatusBitmask(statusBitmask)

	// Broadcast presence update
	if statusBitmask&wire.OServiceUserStatusInvisible != 0 {
		// User going invisible - broadcast departure
		if err := h.BuddyBroadcaster.BroadcastBuddyDeparted(ctx, oscarSession); err != nil {
			h.Logger.ErrorContext(ctx, "failed to broadcast buddy departed", "err", err.Error())
		}
	} else {
		// User visible - broadcast arrival/update
		if err := h.BuddyBroadcaster.BroadcastBuddyArrived(ctx, oscarSession); err != nil {
			h.Logger.ErrorContext(ctx, "failed to broadcast buddy arrived", "err", err.Error())
		}
	}

	// Queue presence event for other WebAPI sessions watching this user
	h.broadcastPresenceEvent(session.ScreenName.IdentScreenName(), stateParam, awayMsg, "")

	h.Logger.InfoContext(ctx, "presence state updated",
		"screenName", session.ScreenName.String(),
		"state", stateParam,
		"hasAwayMsg", awayMsg != "",
	)

	// Send success response
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": 200,
			"statusText": "OK",
		},
	}

	// Check for JSONP callback
	if callback := r.URL.Query().Get("f"); callback != "" {
		SendJSONP(w, callback, response, h.Logger)
	} else {
		SendJSON(w, response, h.Logger)
	}
}

// SetStatus handles GET /presence/setStatus requests to update user's status message.
func (h *PresenceHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
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
		h.sendError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Get the status message
	statusMsg := r.URL.Query().Get("statusMsg")
	statusCode := r.URL.Query().Get("statusCode")

	// Store status message in session (this would normally be stored in a profile/status service)
	// For now, we'll broadcast it as part of presence

	// Get OSCAR session if available
	if oscarSession := session.OSCARSession; oscarSession != nil {
		// In OSCAR, status messages are typically part of the profile
		// We'll need to extend this based on the actual implementation

		// Broadcast presence update with new status
		if err := h.BuddyBroadcaster.BroadcastBuddyArrived(ctx, oscarSession); err != nil {
			h.Logger.ErrorContext(ctx, "failed to broadcast status update", "err", err.Error())
		}
	}

	// Queue status event for other WebAPI sessions
	h.broadcastPresenceEvent(session.ScreenName.IdentScreenName(), "", "", statusMsg)

	h.Logger.InfoContext(ctx, "status message updated",
		"screenName", session.ScreenName.String(),
		"statusMsg", statusMsg,
		"statusCode", statusCode,
	)

	// Send success response
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": 200,
			"statusText": "OK",
		},
	}

	// Check for JSONP callback
	if callback := r.URL.Query().Get("f"); callback != "" {
		SendJSONP(w, callback, response, h.Logger)
	} else {
		SendJSON(w, response, h.Logger)
	}
}

// SetProfile handles GET /presence/setProfile requests to update user's profile.
func (h *PresenceHandler) SetProfile(w http.ResponseWriter, r *http.Request) {
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
		h.sendError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Get the profile content
	profile := r.URL.Query().Get("profile")

	// Limit profile size (4KB max)
	if len(profile) > 4096 {
		h.sendError(w, http.StatusBadRequest, "profile too large (max 4KB)")
		return
	}

	// Save profile using ProfileManager
	if err := h.ProfileManager.SetProfile(ctx, session.ScreenName.IdentScreenName(), profile); err != nil {
		h.Logger.ErrorContext(ctx, "failed to set profile", "err", err.Error())
		h.sendError(w, http.StatusInternalServerError, "failed to save profile")
		return
	}

	h.Logger.InfoContext(ctx, "profile updated",
		"screenName", session.ScreenName.String(),
		"profileSize", len(profile),
	)

	// Send success response
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": 200,
			"statusText": "OK",
		},
	}

	// Check for JSONP callback
	if callback := r.URL.Query().Get("f"); callback != "" {
		SendJSONP(w, callback, response, h.Logger)
	} else {
		SendJSON(w, response, h.Logger)
	}
}

// GetProfile handles GET /presence/getProfile requests to retrieve user's profile.
func (h *PresenceHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
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
		h.sendError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Get target screen name (optional - defaults to self)
	targetSN := r.URL.Query().Get("sn")
	if targetSN == "" {
		targetSN = session.ScreenName.String()
	}

	// Retrieve profile using ProfileManager
	profile, err := h.ProfileManager.Profile(ctx, state.NewIdentScreenName(targetSN))
	if err != nil {
		h.Logger.WarnContext(ctx, "failed to get profile", "err", err.Error())
		// Return empty profile on error
		profile = ""
	}

	// Send response
	response := map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": 200,
			"statusText": "OK",
			"data": map[string]interface{}{
				"screenName":  targetSN,
				"profile":     profile,
				"lastUpdated": time.Now().Unix(),
			},
		},
	}

	// Check for JSONP callback
	if callback := r.URL.Query().Get("f"); callback != "" {
		SendJSONP(w, callback, response, h.Logger)
	} else {
		SendJSON(w, response, h.Logger)
	}
}

// Icon handles GET /presence/icon requests for presence icons.
func (h *PresenceHandler) Icon(w http.ResponseWriter, r *http.Request) {
	// Get parameters
	name := r.URL.Query().Get("name")
	size := r.URL.Query().Get("size")
	iconType := r.URL.Query().Get("type")

	if name == "" {
		h.sendError(w, http.StatusBadRequest, "missing name parameter")
		return
	}

	// Default values
	if size == "" {
		size = "32"
	}
	if iconType == "" {
		iconType = "aim"
	}

	// For now, redirect to a placeholder icon
	// In production, this would redirect to actual icon storage/CDN
	iconURL := "/static/icons/default_" + iconType + "_" + size + ".png"

	// If it's an email lookup, extract username
	if strings.Contains(name, "@") {
		parts := strings.Split(name, "@")
		if len(parts) > 0 {
			name = parts[0]
		}
	}

	// Check if user is online and get their state
	screenName := state.NewIdentScreenName(name)
	if session := h.SessionRetriever.RetrieveSession(screenName); session != nil {
		statusBitmask := session.UserStatusBitmask()
		if statusBitmask&wire.OServiceUserStatusAway != 0 {
			iconURL = "/static/icons/away_" + iconType + "_" + size + ".png"
		} else if session.Idle() {
			iconURL = "/static/icons/idle_" + iconType + "_" + size + ".png"
		} else {
			iconURL = "/static/icons/online_" + iconType + "_" + size + ".png"
		}
	} else {
		iconURL = "/static/icons/offline_" + iconType + "_" + size + ".png"
	}

	// Redirect to icon URL
	http.Redirect(w, r, iconURL, http.StatusFound)
}

// broadcastPresenceEvent sends presence updates to all WebAPI sessions watching this user
func (h *PresenceHandler) broadcastPresenceEvent(screenName state.IdentScreenName, stateStr, awayMsg, statusMsg string) {
	// Get all sessions that have this user in their buddy list
	// For now, we'll broadcast to all sessions (this should be optimized)
	for _, sess := range h.SessionManager.GetAllSessions() {
		if sess.EventQueue != nil && sess.Events != nil {
			// Check if session is subscribed to presence events
			for _, event := range sess.Events {
				if event == "presence" || event == "myInfo" {
					eventData := state.PresenceEvent{
						AimID:     screenName.String(),
						State:     stateStr,
						AwayMsg:   awayMsg,
						StatusMsg: statusMsg,
					}
					sess.EventQueue.Push(state.EventTypePresence, eventData)
					break
				}
			}
		}
	}
}
