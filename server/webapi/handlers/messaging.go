package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// MessageRelayer defines methods for relaying messages between users
type MessageRelayer interface {
	RelayToScreenName(ctx context.Context, recipient state.IdentScreenName, msg wire.SNACMessage)
}

// OfflineMessageManager defines methods for managing offline messages
type OfflineMessageManager interface {
	SaveMessage(ctx context.Context, msg state.OfflineMessage) error
}

// MessagingHandler handles Web AIM API messaging endpoints
type MessagingHandler struct {
	SessionManager        *state.WebAPISessionManager
	MessageRelayer        MessageRelayer
	OfflineMessageManager OfflineMessageManager
	SessionRetriever      SessionRetriever
	Logger                *slog.Logger
}

// MessageResponse is a response structure for messaging API responses.
type MessageResponse struct {
	XMLName  xml.Name `xml:"response" json:"-"`
	Response struct {
		StatusCode int                    `json:"statusCode" xml:"statusCode"`
		StatusText string                 `json:"statusText" xml:"statusText"`
		Data       map[string]interface{} `json:"data,omitempty" xml:"data,omitempty"`
	} `json:"response" xml:"-"`
	// For XML responses, flatten the structure
	StatusCode int                    `json:"-" xml:"statusCode,omitempty"`
	StatusText string                 `json:"-" xml:"statusText,omitempty"`
	Data       map[string]interface{} `json:"-" xml:"data,omitempty"`
}

// SendIMResponseData represents the data portion of a sendIM response.
type SendIMResponseData struct {
	SMSSegmentCount int    `xml:"smsSegmentCount" json:"smsSegmentCount"`
	MessageID       string `xml:"messageId" json:"messageId"`
	Timestamp       int64  `xml:"timestamp" json:"timestamp"`
}

// SendIMXMLResponse is the XML-specific response structure for sendIM.
type SendIMXMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	XMLNS      string   `xml:"xmlns,attr,omitempty"`
	StatusCode int      `xml:"statusCode"`
	StatusText string   `xml:"statusText"`
	RequestID  string   `xml:"requestId,omitempty"`
	Data       struct {
		MsgID string `xml:"msgId"`
		State string `xml:"state"`
	} `xml:"data,omitempty"`
}

// SendIM handles the /im/sendIM endpoint for sending instant messages
func (h *MessagingHandler) SendIM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session from aimsid
	aimsid := r.URL.Query().Get("aimsid")
	if aimsid == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "missing required parameter: aimsid")
		return
	}

	sess, err := h.SessionManager.GetSession(aimsid)
	if err != nil {
		if err == state.ErrNoWebAPISession || err == state.ErrWebAPISessionExpired {
			h.sendErrorResponse(w, http.StatusUnauthorized, "invalid or expired session")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Parse parameters
	recipient := r.URL.Query().Get("t")
	if recipient == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "missing required parameter: t (recipient)")
		return
	}

	message := r.URL.Query().Get("message")
	if message == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "missing required parameter: message")
		return
	}

	// Parse optional parameters
	autoResponse := r.URL.Query().Get("autoResponse") == "1"
	offlineIM := r.URL.Query().Get("offlineIM") != "0" // default to true

	// Create recipient identifier
	recipientIdent := state.NewIdentScreenName(recipient)

	// Check if recipient is online
	recipientSession := h.SessionRetriever.RetrieveSession(recipientIdent)

	// Generate message cookie
	var cookie [8]byte
	if _, err := rand.Read(cookie[:]); err != nil {
		h.Logger.ErrorContext(ctx, "failed to generate message cookie", "error", err)
		h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}
	cookieUint64 := binary.BigEndian.Uint64(cookie[:])

	// Get sender's OSCAR session if available
	var senderInfo wire.TLVUserInfo
	if sess.OSCARSession != nil {
		senderInfo = sess.OSCARSession.TLVUserInfo()
	} else {
		// Create minimal user info for web-only sessions
		senderInfo = wire.TLVUserInfo{
			ScreenName:   sess.ScreenName.String(),
			WarningLevel: 0,
		}
		senderInfo.Append(wire.NewTLVBE(wire.OServiceUserInfoSignonTOD, uint32(sess.CreatedAt.Unix())))
		senderInfo.Append(wire.NewTLVBE(wire.OServiceUserInfoStatus, uint32(0x0000))) // online status
	}

	// Create message ID for response (UUID format like working implementation)
	// Using the cookie bytes to generate a UUID-like string
	messageID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(cookie[:4]),
		binary.BigEndian.Uint16(cookie[4:6]),
		binary.BigEndian.Uint16(cookie[6:8]),
		binary.BigEndian.Uint16([]byte{0x80, 0x00}), // Version bits
		time.Now().UnixNano()&0xffffffffffff)

	if recipientSession == nil {
		// Recipient is offline
		if offlineIM {
			// Save offline message
			offlineMsg := state.OfflineMessage{
				Message: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					Cookie:     cookieUint64,
					ChannelID:  wire.ICBMChannelIM,
					ScreenName: recipient,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ICBMTLVAOLIMData, h.encodeIMMessage(message, autoResponse)),
							wire.NewTLVBE(wire.ICBMTLVStore, uint8(1)), // store offline
						},
					},
				},
				Recipient: recipientIdent,
				Sender:    sess.ScreenName.IdentScreenName(),
				Sent:      time.Now().UTC(),
			}

			if err := h.OfflineMessageManager.SaveMessage(ctx, offlineMsg); err != nil {
				h.Logger.ErrorContext(ctx, "failed to save offline message",
					"from", sess.ScreenName.String(),
					"to", recipient,
					"error", err)
				h.sendErrorResponse(w, http.StatusInternalServerError, "failed to save offline message")
				return
			}

			h.Logger.InfoContext(ctx, "saved offline message",
				"from", sess.ScreenName.String(),
				"to", recipient)
		} else {
			// Recipient is offline and offline delivery is disabled
			h.sendErrorResponse(w, http.StatusNotFound, "recipient is not online")
			return
		}
	} else {
		// Recipient is online, deliver message
		clientIM := wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
			Cookie:       cookieUint64,
			ChannelID:    wire.ICBMChannelIM,
			TLVUserInfo:  senderInfo,
			TLVRestBlock: wire.TLVRestBlock{},
		}

		// Add message data
		clientIM.Append(wire.NewTLVBE(wire.ICBMTLVAOLIMData, h.encodeIMMessage(message, autoResponse)))

		// Add auto-response flag if applicable
		if autoResponse {
			clientIM.Append(wire.NewTLVBE(wire.ICBMTLVAutoResponse, []byte{}))
		}

		// Send message to recipient
		h.MessageRelayer.RelayToScreenName(ctx, recipientIdent, wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMChannelMsgToClient,
				RequestID: wire.ReqIDFromServer,
			},
			Body: clientIM,
		})

		// Queue IM event for the recipient's WebAPI session if they have one
		if recipientWebSession, err := h.SessionManager.GetSessionByUser(recipientIdent); err == nil && recipientWebSession != nil {
			eventData := state.IMEvent{
				From:      sess.ScreenName.String(),
				Message:   message,
				Timestamp: float64(time.Now().Unix()),
				AutoResp:  autoResponse,
			}
			recipientWebSession.EventQueue.Push(state.EventTypeIM, eventData)
		}

		// Also queue sentIM event for the sender's WebAPI session to show in their UI
		senderEventData := state.SentIMEvent{
			Sender: state.UserInfo{
				AimID:     sess.ScreenName.String(),
				DisplayID: sess.ScreenName.String(),
				UserType:  "aim",
			},
			Dest: state.UserInfo{
				AimID:     recipient,
				DisplayID: recipient,
				UserType:  "aim",
			},
			Message:   message,
			Timestamp: float64(time.Now().Unix()),
			AutoResp:  autoResponse,
		}
		sess.EventQueue.Push(state.EventTypeSentIM, senderEventData)

		h.Logger.InfoContext(ctx, "queued sentIM event for sender",
			"from", sess.ScreenName.String(),
			"to", recipient,
			"eventType", state.EventTypeSentIM,
			"queueSize", sess.EventQueue.Size(),
			"subscribedEvents", sess.Events,
			"isSubscribedToSentIM", sess.IsSubscribedTo("sentIM"),
		)

		h.Logger.InfoContext(ctx, "delivered instant message",
			"from", sess.ScreenName.String(),
			"to", recipient)
	}

	// Send success response
	format := strings.ToLower(r.URL.Query().Get("f"))

	if format == "xml" {
		// For XML, use properly typed structure matching working implementation
		response := SendIMXMLResponse{
			XMLNS:      "http://developer.aim.com/xsd/im.xsd",
			StatusCode: 200,
			StatusText: "OK",
			RequestID:  r.URL.Query().Get("r"),
		}
		response.Data.MsgID = messageID
		response.Data.State = "delivered"
		SendXML(w, response, h.Logger)
	} else {
		// For JSON/JSONP/AMF, match working implementation response
		responseData := map[string]interface{}{
			"msgId": messageID,
			"state": "delivered",
		}
		response := MessageResponse{}
		response.Response.StatusCode = 200
		response.Response.StatusText = "OK"
		response.Response.Data = responseData
		SendResponse(w, r, response, h.Logger)
	}
}

// encodeIMMessage encodes a text message into the OSCAR IM format
func (h *MessagingHandler) encodeIMMessage(text string, autoResponse bool) []byte {
	// Create ICBM fragment list for the message
	frags, err := wire.ICBMFragmentList(text)
	if err != nil {
		// If fragment creation fails, return simple text bytes
		return []byte(text)
	}

	// Marshal the fragments
	buf := &bytes.Buffer{}
	for _, frag := range frags {
		if err := wire.MarshalBE(frag, buf); err != nil {
			// If marshaling fails, return simple text bytes
			return []byte(text)
		}
	}
	return buf.Bytes()
}

// sendErrorResponse sends an error response in Web AIM API format
func (h *MessagingHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, errorText string) {
	SendError(w, statusCode, errorText)
}

// SetTyping handles the /im/setTyping endpoint for typing indicators
func (h *MessagingHandler) SetTyping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session from aimsid
	aimsid := r.URL.Query().Get("aimsid")
	if aimsid == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "missing required parameter: aimsid")
		return
	}

	sess, err := h.SessionManager.GetSession(aimsid)
	if err != nil {
		if err == state.ErrNoWebAPISession || err == state.ErrWebAPISessionExpired {
			h.sendErrorResponse(w, http.StatusUnauthorized, "invalid or expired session")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Update session activity
	if err := h.SessionManager.TouchSession(aimsid); err != nil {
		h.Logger.WarnContext(ctx, "failed to touch session", "aimsid", aimsid, "error", err)
	}

	// Parse parameters
	recipient := r.URL.Query().Get("t")
	if recipient == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "missing required parameter: t (recipient)")
		return
	}

	typingStr := r.URL.Query().Get("typing")
	typing := false
	if typingStr != "" {
		var err error
		typing, err = strconv.ParseBool(typingStr)
		if err != nil {
			// Try numeric format (0/1)
			typing = typingStr == "1"
		}
	}

	// Create recipient identifier
	recipientIdent := state.NewIdentScreenName(recipient)

	// Check if recipient is online
	recipientSession := h.SessionRetriever.RetrieveSession(recipientIdent)
	if recipientSession == nil {
		// Silently succeed even if recipient is offline
		h.sendSuccessResponse(w, r, nil)
		return
	}

	// Generate typing notification cookie
	var cookie [8]byte
	if _, err := rand.Read(cookie[:]); err != nil {
		h.Logger.ErrorContext(ctx, "failed to generate typing cookie", "error", err)
		h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}
	cookieUint64 := binary.BigEndian.Uint64(cookie[:])

	// Create typing notification
	var notificationType uint16
	if typing {
		notificationType = 0x0002 // Typing started
	} else {
		notificationType = 0x0001 // Typing stopped
	}

	typingNotification := wire.SNAC_0x04_0x14_ICBMClientEvent{
		Cookie:     cookieUint64,
		ChannelID:  wire.ICBMChannelIM,
		ScreenName: sess.ScreenName.String(),
		Event:      notificationType,
	}

	// Send typing notification to recipient
	h.MessageRelayer.RelayToScreenName(ctx, recipientIdent, wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMClientEvent,
			RequestID: wire.ReqIDFromServer,
		},
		Body: typingNotification,
	})

	// Queue typing event for the recipient's WebAPI session if they have one
	if recipientWebSession, err := h.SessionManager.GetSessionByUser(recipientIdent); err == nil && recipientWebSession != nil {
		eventData := state.TypingEvent{
			From:   sess.ScreenName.String(),
			Typing: typing,
		}
		recipientWebSession.EventQueue.Push(state.EventTypeTyping, eventData)
	}

	h.Logger.DebugContext(ctx, "sent typing notification",
		"from", sess.ScreenName.String(),
		"to", recipient,
		"typing", typing)

	// Send success response
	h.sendSuccessResponse(w, r, nil)
}

// sendSuccessResponse sends a success response in Web AIM API format
func (h *MessagingHandler) sendSuccessResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	format := strings.ToLower(r.URL.Query().Get("f"))

	var responseData map[string]interface{}
	if data != nil {
		if mapData, ok := data.(map[string]interface{}); ok {
			responseData = mapData
		}
	}

	if format == "xml" {
		// For XML, use flattened structure
		response := MessageResponse{}
		response.StatusCode = 200
		response.StatusText = "OK"
		response.Data = responseData
		SendXML(w, response, h.Logger)
	} else {
		// For JSON/JSONP, use nested structure
		response := MessageResponse{}
		response.Response.StatusCode = 200
		response.Response.StatusText = "OK"
		response.Response.Data = responseData
		SendResponse(w, r, response, h.Logger)
	}
}
