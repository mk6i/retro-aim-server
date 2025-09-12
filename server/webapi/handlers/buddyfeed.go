package handlers

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/mk6i/retro-aim-server/state"
)

// BuddyFeedHandler handles Web AIM API buddy feed endpoints.
type BuddyFeedHandler struct {
	SessionManager   *state.WebAPISessionManager
	FeedManager      *state.BuddyFeedManager
	SessionRetriever SessionRetriever
	Logger           *slog.Logger
}

// GetUser handles GET /buddyfeed/getUser requests to retrieve a user's feed.
func (h *BuddyFeedHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get target user from 't' parameter as per spec
	targetUser := r.URL.Query().Get("t")
	if targetUser == "" {
		SendError(w, http.StatusBadRequest, "missing 't' parameter")
		return
	}

	// Get format parameter
	format := strings.ToLower(r.URL.Query().Get("f"))

	// Handle RSS as native format
	if format == "rss" || format == "native" {
		feedType := "rss"
		if r.URL.Query().Get("f") == "atom" {
			feedType = "atom"
		}

		feedData, err := h.FeedManager.GetUserFeed(ctx, targetUser, feedType)
		if err != nil {
			h.Logger.ErrorContext(ctx, "failed to get user feed",
				"user", targetUser,
				"error", err,
			)
			SendError(w, http.StatusInternalServerError, "failed to retrieve feed")
			return
		}

		if feedType == "atom" {
			h.sendAtomFeed(w, feedData)
		} else {
			h.sendRSSFeed(w, feedData)
		}
		return
	}

	h.Logger.DebugContext(ctx, "retrieving user feed",
		"user", targetUser,
		"format", format,
	)

	// Get the user's feed in JSON format for standard response
	feedData, err := h.FeedManager.GetUserFeed(ctx, targetUser, "json")
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to get user feed",
			"user", targetUser,
			"error", err,
		)
		SendError(w, http.StatusInternalServerError, "failed to retrieve feed")
		return
	}

	// Use centralized response handler for all formats
	response := BaseResponse{}
	response.Response.StatusCode = 200
	response.Response.StatusText = "OK"
	response.Response.Data = feedData
	SendResponse(w, r, response, h.Logger)
}

// GetBuddylist handles GET /buddyfeed/getBuddylist requests to retrieve aggregated buddy feeds.
func (h *BuddyFeedHandler) GetBuddylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authentication required for buddy list feed
	aimsid := r.URL.Query().Get("aimsid")
	if aimsid == "" {
		SendError(w, http.StatusBadRequest, "missing aimsid parameter")
		return
	}

	// Get session
	session, err := h.SessionManager.GetSession(aimsid)
	if err != nil {
		SendError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Get format and limit parameters
	format := strings.ToLower(r.URL.Query().Get("f"))
	if format == "" {
		format = "rss"
	}

	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 500 {
				limit = 500 // Max limit
			}
		}
	}

	h.Logger.DebugContext(ctx, "retrieving buddy list feed",
		"screenName", session.ScreenName.String(),
		"format", format,
		"limit", limit,
	)

	// Get buddy list from OSCAR session if available
	var buddies []state.IdentScreenName

	// Return empty feed for now - buddy list integration pending
	if len(buddies) == 0 {
		h.Logger.InfoContext(ctx, "no buddies found for feed aggregation",
			"screenName", session.ScreenName.String(),
		)

		// Return empty feed
		emptyFeed := map[string]interface{}{
			"title":       fmt.Sprintf("%s's Buddy Feed", session.ScreenName.String()),
			"description": "Aggregated feed from your buddy list",
			"items":       []interface{}{},
		}

		if format == "json" {
			response := BaseResponse{}
			response.Response.StatusCode = 200
			response.Response.StatusText = "OK"
			response.Response.Data = emptyFeed
			SendResponse(w, r, response, h.Logger)
		} else {
			h.sendEmptyRSSFeed(w, session.ScreenName.String())
		}
		return
	}

	// Get aggregated feed
	feedData, err := h.FeedManager.GetBuddyListFeed(ctx, buddies, format, limit)
	if err != nil {
		h.Logger.ErrorContext(ctx, "failed to get buddy list feed",
			"screenName", session.ScreenName.String(),
			"error", err,
		)
		SendError(w, http.StatusInternalServerError, "failed to retrieve feed")
		return
	}

	// Send response based on format
	switch format {
	case "atom":
		h.sendAtomFeed(w, feedData)
	case "json":
		response := BaseResponse{}
		response.Response.StatusCode = 200
		response.Response.StatusText = "OK"
		response.Response.Data = feedData
		SendResponse(w, r, response, h.Logger)
	case "rss", "":
		h.sendRSSFeed(w, feedData)
	default:
		h.sendRSSFeed(w, feedData)
	}
}

// PushFeed handles GET /buddyfeed/pushFeed requests to submit feed updates.
func (h *BuddyFeedHandler) PushFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authentication token or session
	token := r.URL.Query().Get("a")
	aimsid := r.URL.Query().Get("aimsid")

	if token == "" && aimsid == "" {
		SendError(w, http.StatusBadRequest, "authentication required")
		return
	}

	var screenName string
	if aimsid != "" {
		session, err := h.SessionManager.GetSession(aimsid)
		if err != nil {
			SendError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		screenName = session.ScreenName.String()

		if err := h.SessionManager.TouchSession(aimsid); err != nil {
			h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
		}
	} else {
		// Extract screen name from token authentication
		screenName = r.URL.Query().Get("s")
		if screenName == "" {
			SendError(w, http.StatusBadRequest, "missing source user")
			return
		}
	}

	// Extract feed parameters as per spec
	feedTitle := r.URL.Query().Get("feedTitle")
	feedLink := r.URL.Query().Get("feedLink")
	feedDesc := r.URL.Query().Get("feedDesc")
	itemTitle := r.URL.Query().Get("itemTitle")
	itemLink := r.URL.Query().Get("itemLink")
	itemGuid := r.URL.Query().Get("itemGuid")

	// Validate required parameters
	if feedTitle == "" || feedLink == "" || feedDesc == "" ||
		itemTitle == "" || itemLink == "" || itemGuid == "" {
		SendError(w, http.StatusBadRequest, "missing required feed parameters")
		return
	}

	// Build feed data structure
	feedData := map[string]interface{}{
		"title":       itemTitle,
		"description": r.URL.Query().Get("itemDesc"),
		"link":        itemLink,
		"guid":        itemGuid,
		"feedTitle":   feedTitle,
		"feedLink":    feedLink,
		"feedDesc":    feedDesc,
	}

	// Optional parameters
	if publisher := r.URL.Query().Get("feedPublisher"); publisher != "" {
		feedData["publisher"] = publisher
	}
	if pubDate := r.URL.Query().Get("itemPubDate"); pubDate != "" {
		feedData["pubDate"] = pubDate
	}
	if category := r.URL.Query().Get("itemCategory"); category != "" {
		feedData["categories"] = []string{category}
	}

	h.Logger.InfoContext(ctx, "pushing feed update",
		"screenName", screenName,
		"itemTitle", itemTitle,
	)

	// Push the feed update
	if err := h.FeedManager.PushFeed(ctx, screenName, feedData); err != nil {
		h.Logger.ErrorContext(ctx, "failed to push feed",
			"screenName", screenName,
			"error", err,
		)
		SendError(w, http.StatusInternalServerError, "failed to push feed")
		return
	}

	// Send success response
	response := BaseResponse{}
	response.Response.StatusCode = 200
	response.Response.StatusText = "OK"
	response.Response.Data = map[string]interface{}{
		"success": true,
	}

	SendResponse(w, r, response, h.Logger)
}

// sendRSSFeed sends an RSS feed response.
func (h *BuddyFeedHandler) sendRSSFeed(w http.ResponseWriter, feed interface{}) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	// Add XML declaration
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))

	// Marshal and write the feed
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		h.Logger.Error("failed to encode RSS feed", "error", err)
	}
}

// sendAtomFeed sends an Atom feed response.
func (h *BuddyFeedHandler) sendAtomFeed(w http.ResponseWriter, feed interface{}) {
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

	// Add XML declaration
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))

	// Marshal and write the feed
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		h.Logger.Error("failed to encode Atom feed", "error", err)
	}
}

// sendEmptyRSSFeed sends an empty RSS feed.
func (h *BuddyFeedHandler) sendEmptyRSSFeed(w http.ResponseWriter, screenName string) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	emptyFeed := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>%s's Buddy Feed</title>
    <description>Aggregated feed from your buddy list</description>
    <link>/buddyfeed/getBuddylist</link>
    <language>en-US</language>
  </channel>
</rss>`, screenName)

	w.Write([]byte(emptyFeed))
}
