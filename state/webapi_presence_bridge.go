package state

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mk6i/retro-aim-server/wire"
)

// RelationshipFetcher defines methods for fetching user relationships.
type RelationshipFetcher interface {
	Relationship(ctx context.Context, me IdentScreenName, them IdentScreenName) (Relationship, error)
	AllRelationships(ctx context.Context, me IdentScreenName, filter []IdentScreenName) ([]Relationship, error)
}

// WebAPIPresenceBridge bridges OSCAR presence events to WebAPI sessions.
// It listens for buddy arrival/departure events and pushes them to WebAPI event queues.
type WebAPIPresenceBridge struct {
	sessionManager      *WebAPISessionManager
	buddyListManager    *WebAPIBuddyListManager
	feedbagRetriever    FeedbagRetriever
	relationshipFetcher RelationshipFetcher
	logger              *slog.Logger
}

// NewWebAPIPresenceBridge creates a new presence bridge.
func NewWebAPIPresenceBridge(
	sessionManager *WebAPISessionManager,
	buddyListManager *WebAPIBuddyListManager,
	feedbagRetriever FeedbagRetriever,
	relationshipFetcher RelationshipFetcher,
	logger *slog.Logger,
) *WebAPIPresenceBridge {
	return &WebAPIPresenceBridge{
		sessionManager:      sessionManager,
		buddyListManager:    buddyListManager,
		feedbagRetriever:    feedbagRetriever,
		relationshipFetcher: relationshipFetcher,
		logger:              logger,
	}
}

// BroadcastBuddyArrived handles when a buddy comes online.
func (b *WebAPIPresenceBridge) BroadcastBuddyArrived(ctx context.Context, arrivingSession *Session) error {
	arrivingScreenName := arrivingSession.IdentScreenName()
	b.logger.Debug("broadcasting buddy arrival",
		"buddy", arrivingScreenName)

	// Find all WebAPI sessions that have this user as a buddy
	sessions := b.findSessionsWithBuddy(ctx, arrivingScreenName)

	// Get presence info for the arriving buddy
	buddyInfo := b.buddyListManager.GetPresenceForBuddy(arrivingScreenName.String())

	// Push presence event to each session
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("presence") {
			presenceEvent := PresenceEvent{
				AimID:      buddyInfo.AimID,
				State:      buddyInfo.State,
				StatusMsg:  buddyInfo.StatusMsg,
				AwayMsg:    buddyInfo.AwayMsg,
				IdleTime:   buddyInfo.IdleTime,
				OnlineTime: buddyInfo.OnlineTime,
				UserType:   buddyInfo.UserType,
			}

			session.EventQueue.Push(EventTypePresence, presenceEvent)

			b.logger.Debug("pushed buddy arrival event",
				"session", session.AimSID,
				"buddy", arrivingScreenName)
		}
	}

	return nil
}

// BroadcastBuddyDeparted handles when a buddy goes offline.
func (b *WebAPIPresenceBridge) BroadcastBuddyDeparted(ctx context.Context, departingSession *Session) error {
	departingScreenName := departingSession.IdentScreenName()
	b.logger.Debug("broadcasting buddy departure",
		"buddy", departingScreenName)

	// Find all WebAPI sessions that have this user as a buddy
	sessions := b.findSessionsWithBuddy(ctx, departingScreenName)

	// Create offline presence event
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("presence") {
			presenceEvent := PresenceEvent{
				AimID:    departingScreenName.String(),
				State:    "offline",
				UserType: "aim",
			}

			session.EventQueue.Push(EventTypePresence, presenceEvent)

			b.logger.Debug("pushed buddy departure event",
				"session", session.AimSID,
				"buddy", departingScreenName)
		}
	}

	return nil
}

// BroadcastBuddyStatusUpdate handles when a buddy's status changes (away, idle, etc).
func (b *WebAPIPresenceBridge) BroadcastBuddyStatusUpdate(ctx context.Context, updatedSession *Session) error {
	updatedScreenName := updatedSession.IdentScreenName()
	b.logger.Debug("broadcasting buddy status update",
		"buddy", updatedScreenName,
		"awayMessage", updatedSession.AwayMessage() != "",
		"idle", updatedSession.Idle())

	// Find all WebAPI sessions that have this user as a buddy
	sessions := b.findSessionsWithBuddy(ctx, updatedScreenName)

	// Get updated presence info
	buddyInfo := b.buddyListManager.GetPresenceForBuddy(updatedScreenName.String())

	// Push presence event to each session
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("presence") {
			presenceEvent := PresenceEvent{
				AimID:      buddyInfo.AimID,
				State:      buddyInfo.State,
				StatusMsg:  buddyInfo.StatusMsg,
				AwayMsg:    buddyInfo.AwayMsg,
				IdleTime:   buddyInfo.IdleTime,
				OnlineTime: buddyInfo.OnlineTime,
				UserType:   buddyInfo.UserType,
			}

			session.EventQueue.Push(EventTypePresence, presenceEvent)

			b.logger.Debug("pushed buddy status update event",
				"session", session.AimSID,
				"buddy", updatedScreenName,
				"state", buddyInfo.State)
		}
	}

	return nil
}

// findSessionsWithBuddy finds all WebAPI sessions that have the specified user as a buddy.
func (b *WebAPIPresenceBridge) findSessionsWithBuddy(ctx context.Context, buddyScreenName IdentScreenName) []*WebAPISession {
	var sessions []*WebAPISession

	// Get all active WebAPI sessions
	allSessions := b.sessionManager.GetAllSessions()

	for _, session := range allSessions {
		// Skip anonymous sessions
		if session.ScreenName.String() == "" {
			continue
		}

		// Check if this session has the buddy in their list
		if b.hasBuddy(ctx, session.ScreenName.IdentScreenName(), buddyScreenName) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// hasBuddy checks if a user has another user as a buddy in their feedbag.
func (b *WebAPIPresenceBridge) hasBuddy(ctx context.Context, userScreenName, buddyScreenName IdentScreenName) bool {
	// First check blocking relationship (OSCAR compliant)
	if b.relationshipFetcher != nil {
		rel, err := b.relationshipFetcher.Relationship(ctx, userScreenName, buddyScreenName)
		if err != nil {
			b.logger.Error("failed to get relationship",
				"user", userScreenName,
				"buddy", buddyScreenName,
				"error", err)
		} else {
			// OSCAR compliance: if either blocks the other, they don't see presence
			if rel.YouBlock || rel.BlocksYou {
				return false
			}
		}
	}

	// Retrieve user's feedbag
	items, err := b.feedbagRetriever.RetrieveFeedbag(ctx, userScreenName)
	if err != nil {
		b.logger.Error("failed to retrieve feedbag",
			"user", userScreenName,
			"error", err)
		return false
	}

	// Check if buddy exists in feedbag
	for _, item := range items {
		// Use wire.FeedbagClassIdBuddy constant instead of hardcoded value
		if item.ClassID == wire.FeedbagClassIdBuddy &&
			NewIdentScreenName(item.Name) == buddyScreenName {
			return true
		}
	}

	return false
}

// PushInitialBuddyList pushes the initial buddy list to a session's event queue.
func (b *WebAPIPresenceBridge) PushInitialBuddyList(ctx context.Context, session *WebAPISession) error {
	// Skip if not subscribed to buddylist events
	if !session.IsSubscribedTo("buddylist") {
		return nil
	}

	// Get buddy list for user
	groups, err := b.buddyListManager.GetBuddyListForUser(ctx, session.ScreenName.IdentScreenName())
	if err != nil {
		return fmt.Errorf("failed to get buddy list: %w", err)
	}

	// Push buddy list event
	buddyListData := b.buddyListManager.FormatBuddyListEvent(groups)
	session.EventQueue.Push(EventTypeBuddyList, buddyListData)

	b.logger.Debug("pushed initial buddy list",
		"session", session.AimSID,
		"groups", len(groups))

	// Also push initial presence events for online buddies
	for _, group := range groups {
		for _, buddy := range group.Buddies {
			if buddy.State != "offline" {
				presenceEvent := PresenceEvent{
					AimID:      buddy.AimID,
					State:      buddy.State,
					StatusMsg:  buddy.StatusMsg,
					AwayMsg:    buddy.AwayMsg,
					IdleTime:   buddy.IdleTime,
					OnlineTime: buddy.OnlineTime,
					UserType:   buddy.UserType,
				}
				session.EventQueue.Push(EventTypePresence, presenceEvent)
			}
		}
	}

	return nil
}

// PushBuddyAddedEvent pushes an event when a buddy is added to the list.
func (b *WebAPIPresenceBridge) PushBuddyAddedEvent(ctx context.Context, userScreenName IdentScreenName, buddyName string, groupName string) {
	// Find WebAPI sessions for this user
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(userScreenName.String()))

	// Get buddy info
	buddyInfo := b.buddyListManager.GetPresenceForBuddy(buddyName)

	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("buddylist") {
			// Convert buddy to map for proper AMF3 encoding
			buddyMap := map[string]interface{}{
				"aimId":     buddyInfo.AimID,
				"displayId": buddyInfo.DisplayID,
				"state":     buddyInfo.State,
				"userType":  buddyInfo.UserType,
				"bot":       buddyInfo.Bot,
				"service":   buddyInfo.Service,
			}
			// Add optional fields
			if buddyInfo.StatusMsg != "" {
				buddyMap["statusMsg"] = buddyInfo.StatusMsg
			}
			if buddyInfo.AwayMsg != "" {
				buddyMap["awayMsg"] = buddyInfo.AwayMsg
			}
			if buddyInfo.OnlineTime > 0 {
				buddyMap["onlineTime"] = float64(buddyInfo.OnlineTime)
			}
			if buddyInfo.IdleTime > 0 {
				buddyMap["idleTime"] = buddyInfo.IdleTime
			}

			event := BuddyListEvent{
				Action: "add",
				Buddy:  buddyMap,
				Group:  groupName,
			}
			session.EventQueue.Push(EventTypeBuddyList, event)
		}
	}
}

// PushBuddyRemovedEvent pushes an event when a buddy is removed from the list.
func (b *WebAPIPresenceBridge) PushBuddyRemovedEvent(ctx context.Context, userScreenName IdentScreenName, buddyName string) {
	// Find WebAPI sessions for this user
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(userScreenName.String()))

	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("buddylist") {
			event := BuddyListEvent{
				Action: "remove",
				Buddy: map[string]string{
					"aimId": buddyName,
				},
			}
			session.EventQueue.Push(EventTypeBuddyList, event)
		}
	}
}
