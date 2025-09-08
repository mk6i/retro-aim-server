package state

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/mk6i/retro-aim-server/wire"
)

// WebAPIMessageBridge bridges OSCAR instant messages to WebAPI sessions.
// It handles incoming messages, typing notifications, and offline messages.
type WebAPIMessageBridge struct {
	sessionManager        *WebAPISessionManager
	sessionRetriever      SessionRetriever
	feedbagRetriever      FeedbagRetriever
	offlineMessageManager OfflineMessageManager
	logger                *slog.Logger
}

// OfflineMessageManager provides methods for managing offline messages.
type OfflineMessageManager interface {
	RetrieveMessages(ctx context.Context, screenName IdentScreenName) ([]OfflineMessage, error)
	DeleteMessages(ctx context.Context, screenName IdentScreenName) error
}

// NewWebAPIMessageBridge creates a new message bridge.
func NewWebAPIMessageBridge(
	sessionManager *WebAPISessionManager,
	sessionRetriever SessionRetriever,
	feedbagRetriever FeedbagRetriever,
	offlineMessageManager OfflineMessageManager,
	logger *slog.Logger,
) *WebAPIMessageBridge {
	return &WebAPIMessageBridge{
		sessionManager:        sessionManager,
		sessionRetriever:      sessionRetriever,
		feedbagRetriever:      feedbagRetriever,
		offlineMessageManager: offlineMessageManager,
		logger:                logger,
	}
}

// DeliverMessage delivers an instant message to WebAPI sessions.
func (b *WebAPIMessageBridge) DeliverMessage(ctx context.Context, from IdentScreenName, to IdentScreenName, message string, autoResponse bool) error {
	b.logger.Debug("delivering message to WebAPI",
		"from", from,
		"to", to,
		"autoResponse", autoResponse)

	// Find WebAPI sessions for the recipient
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(to.String()))

	if len(sessions) == 0 {
		b.logger.Debug("no WebAPI sessions found for recipient", "to", to)
		return fmt.Errorf("no sessions found for %s", to)
	}

	// Create IM event
	imEvent := IMEvent{
		From:      from.String(),
		Message:   message,
		Timestamp: float64(time.Now().Unix()),
		AutoResp:  autoResponse,
	}

	// Push to each session's event queue
	delivered := false
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("im") {
			session.EventQueue.Push(EventTypeIM, imEvent)
			delivered = true

			b.logger.Debug("pushed IM event to session",
				"session", session.AimSID,
				"from", from,
				"to", to)
		}
	}

	if !delivered {
		return fmt.Errorf("no sessions subscribed to IM events for %s", to)
	}

	return nil
}

// DeliverTypingNotification delivers a typing notification to WebAPI sessions.
func (b *WebAPIMessageBridge) DeliverTypingNotification(ctx context.Context, from IdentScreenName, to IdentScreenName, typing bool) error {
	b.logger.Debug("delivering typing notification",
		"from", from,
		"to", to,
		"typing", typing)

	// Find WebAPI sessions for the recipient
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(to.String()))

	if len(sessions) == 0 {
		return nil // Typing notifications are non-critical
	}

	// Create typing event
	typingEvent := TypingEvent{
		From:   from.String(),
		Typing: typing,
	}

	// Push to each session's event queue
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("typing") {
			session.EventQueue.Push(EventTypeTyping, typingEvent)

			b.logger.Debug("pushed typing event to session",
				"session", session.AimSID,
				"from", from,
				"typing", typing)
		}
	}

	return nil
}

// DeliverOfflineMessages delivers queued offline messages to a WebAPI session.
func (b *WebAPIMessageBridge) DeliverOfflineMessages(ctx context.Context, session *WebAPISession) error {
	if !session.IsSubscribedTo("offlineIM") {
		return nil
	}

	// Retrieve offline messages
	messages, err := b.offlineMessageManager.RetrieveMessages(ctx, session.ScreenName.IdentScreenName())
	if err != nil {
		return fmt.Errorf("failed to retrieve offline messages: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	b.logger.Debug("delivering offline messages",
		"session", session.AimSID,
		"count", len(messages))

	// Push each offline message as an event
	for _, msg := range messages {
		// Extract message text from the SNAC structure
		messageText := ""
		for _, tlv := range msg.Message.TLVRestBlock.TLVList {
			if tlv.Tag == 0x0002 { // Message data TLV
				// Message fragment parsing simplified for now
				messageText = string(tlv.Value)
				break
			}
		}

		imEvent := IMEvent{
			From:      msg.Sender.String(),
			Message:   messageText,
			Timestamp: float64(msg.Sent.Unix()),
			AutoResp:  false,
		}

		session.EventQueue.Push(EventTypeOfflineIM, imEvent)
	}

	// Delete offline messages after delivery
	if err := b.offlineMessageManager.DeleteMessages(ctx, session.ScreenName.IdentScreenName()); err != nil {
		b.logger.Error("failed to delete offline messages",
			"user", session.ScreenName,
			"error", err)
	}

	return nil
}

// BridgeOSCARMessage bridges an OSCAR message to WebAPI.
// This is called when an OSCAR client sends a message to a user who has WebAPI sessions.
func (b *WebAPIMessageBridge) BridgeOSCARMessage(ctx context.Context, message wire.SNACMessage, from *Session, to IdentScreenName) error {
	// Extract message content based on SNAC type
	var messageText string
	var autoResponse bool

	switch body := message.Body.(type) {
	case wire.SNAC_0x04_0x06_ICBMChannelMsgToHost:
		// Regular IM
		for _, part := range body.TLVRestBlock.TLVList {
			if part.Tag == 0x0002 { // Message data TLV
				// Parse message fragment
				// This is simplified - real implementation would need proper TLV parsing
				messageText = string(part.Value)
				break
			}
		}
	case wire.SNAC_0x04_0x07_ICBMChannelMsgToClient:
		// Message from client
		for _, part := range body.TLVRestBlock.TLVList {
			if part.Tag == 0x0002 {
				messageText = string(part.Value)
				break
			}
		}
		// Check if auto-response (channel 2 is typically used for auto-responses)
		// Note: The Channel field may not exist in this SNAC type
		// autoResponse = false // Default to false for now
	default:
		b.logger.Debug("unsupported OSCAR message type",
			"type", fmt.Sprintf("%T", body))
		return nil
	}

	if messageText == "" {
		return nil
	}

	// Deliver to WebAPI sessions
	return b.DeliverMessage(ctx, from.IdentScreenName(), to, messageText, autoResponse)
}

// BridgeOSCARTyping bridges an OSCAR typing notification to WebAPI.
func (b *WebAPIMessageBridge) BridgeOSCARTyping(ctx context.Context, from *Session, to IdentScreenName, typing bool) error {
	return b.DeliverTypingNotification(ctx, from.IdentScreenName(), to, typing)
}

// HandleDataMessage handles data messages (like file transfer requests).
func (b *WebAPIMessageBridge) HandleDataMessage(ctx context.Context, from IdentScreenName, to IdentScreenName, data []byte, capability string) error {
	b.logger.Debug("handling data message",
		"from", from,
		"to", to,
		"capability", capability,
		"dataLen", len(data))

	// Find WebAPI sessions for the recipient
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(to.String()))

	if len(sessions) == 0 {
		return nil
	}

	// Encode data as base64 for web transport
	encodedData := base64.StdEncoding.EncodeToString(data)

	// Create data IM event
	dataEvent := map[string]interface{}{
		"from":       from.String(),
		"capability": capability,
		"data":       encodedData,
		"timestamp":  float64(time.Now().Unix()),
	}

	// Push to each session's event queue
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("dataIM") {
			session.EventQueue.Push(WebAPIEventType("dataIM"), dataEvent)

			b.logger.Debug("pushed data IM event to session",
				"session", session.AimSID,
				"from", from,
				"capability", capability)
		}
	}

	return nil
}

// NotifyUserAddedToBuddyList notifies a user that someone added them to their buddy list.
func (b *WebAPIMessageBridge) NotifyUserAddedToBuddyList(ctx context.Context, adder IdentScreenName, added IdentScreenName) error {
	b.logger.Debug("notifying user added to buddy list",
		"adder", adder,
		"added", added)

	// Find WebAPI sessions for the added user
	sessions := b.sessionManager.GetSessionsByScreenName(DisplayScreenName(added.String()))

	if len(sessions) == 0 {
		return nil
	}

	// Create notification event
	event := map[string]string{
		"requester": adder.String(),
	}

	// Push to each session's event queue
	for _, session := range sessions {
		if session.EventQueue != nil && session.IsSubscribedTo("userAddedToBuddyList") {
			session.EventQueue.Push(WebAPIEventType("userAddedToBuddyList"), event)

			b.logger.Debug("pushed user added notification",
				"session", session.AimSID,
				"adder", adder)
		}
	}

	return nil
}
