package handler

import (
	"bytes"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
)

// NewAuthService creates a new instance of AuthService.
func NewAuthService(cfg server.Config, sessionManager SessionManager, messageRelayer MessageRelayer, feedbagManager FeedbagManager, userManager UserManager, chatRegistry ChatRegistry) *AuthService {
	return &AuthService{
		chatRegistry:   chatRegistry,
		config:         cfg,
		feedbagManager: feedbagManager,
		messageRelayer: messageRelayer,
		sessionManager: sessionManager,
		userManager:    userManager,
	}
}

// AuthService provides user BUCP login and session management services.
type AuthService struct {
	chatRegistry   ChatRegistry
	config         server.Config
	feedbagManager FeedbagManager
	messageRelayer MessageRelayer
	sessionManager SessionManager
	userManager    UserManager
}

// RetrieveChatSession returns a chat room session. Return nil if the session
// does not exist.
func (s AuthService) RetrieveChatSession(chatID string, sessionID string) (*state.Session, error) {
	_, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return nil, err
	}
	return chatSessMgr.(SessionManager).RetrieveSession(sessionID), nil
}

// RetrieveBOSSession returns a user's session. Return nil if the session does
// not exist.
func (s AuthService) RetrieveBOSSession(sessionID string) (*state.Session, error) {
	return s.sessionManager.RetrieveSession(sessionID), nil
}

// Signout removes user from the BOS server and notifies adjacent users (those
// who have this user's screen name on their buddy list) of their departure.
func (s AuthService) Signout(ctx context.Context, sess *state.Session) error {
	if err := broadcastDeparture(ctx, sess, s.messageRelayer, s.feedbagManager); err != nil {
		return err
	}
	s.sessionManager.RemoveSession(sess)
	return nil
}

// SignoutChat removes user from chat room and notifies remaining participants
// of their departure. If user is the last to leave, the chat room is deleted.
func (s AuthService) SignoutChat(ctx context.Context, sess *state.Session, chatID string) error {
	chatRoom, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return err
	}
	alertUserLeft(ctx, sess, chatSessMgr.(ChatMessageRelayer))
	chatSessMgr.(SessionManager).RemoveSession(sess)
	if chatSessMgr.(SessionManager).Empty() {
		s.chatRegistry.Remove(chatRoom.Cookie)
	}
	return nil
}

// BUCPChallengeRequestHandler satisfies the client request for a random auth
// key. It returns SNAC oscar.BUCPChallengeResponse. If the screen name in
// TLV oscar.TLVScreenName in bodyIn is recognized as a valid user, the
// response contains the account's auth key, which salt's the user's MD5
// password hash. If the account is invalid, an error code is set in TLV
// oscar.TLVErrorSubcode. If login credentials are invalid and app config
// DisableAuth is true, a stub auth key is generated and a successful challenge
// response is returned.
func (s AuthService) BUCPChallengeRequestHandler(bodyIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUIDFn func() uuid.UUID) (oscar.SNACMessage, error) {
	screenName, exists := bodyIn.GetString(oscar.TLVScreenName)
	if !exists {
		return oscar.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	u, err := s.userManager.GetUser(screenName)
	switch {
	case err != nil:
		return oscar.SNACMessage{}, err
	case u != nil:
		// user lookup succeeded
		authKey = u.AuthKey
	case s.config.DisableAuth:
		// can't find user, generate stub auth key
		authKey = newUUIDFn().String()
	default:
		// can't find user, return login error
		snacFrameOut := oscar.SNACFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPLoginResponse,
		}
		snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
		snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)))
		return oscar.SNACMessage{
			Frame: snacFrameOut,
			Body:  snacPayloadOut,
		}, nil
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPChallengeResponse,
		},
		Body: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
			AuthKey: authKey,
		},
	}, nil
}

// BUCPLoginRequestHandler verifies user credentials. Upon successful login, a
// session is created.
// If login credentials are invalid and app config DisableAuth is true, a stub
// user is created and login continues as normal. DisableAuth allows you to
// skip the account creation procedure, which simplifies the login flow during
// development.
// If login is successful, the SNAC TLV list contains the BOS server address
// (oscar.TLVReconnectHere) and an authorization cookie
// (oscar.TLVAuthorizationCookie). Else, an error code is set
// (oscar.TLVErrorSubcode).
func (s AuthService) BUCPLoginRequestHandler(bodyIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUIDFn func() uuid.UUID, newUserFn func(screenName string) (state.User, error)) (oscar.SNACMessage, error) {
	screenName, found := bodyIn.GetString(oscar.TLVScreenName)
	if !found {
		return oscar.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}
	md5Hash, found := bodyIn.GetSlice(oscar.TLVPasswordHash)
	if !found {
		return oscar.SNACMessage{}, errors.New("password hash doesn't exist in tlv")
	}

	loginOK := false

	u, err := s.userManager.GetUser(screenName)
	switch {
	case err != nil:
		return oscar.SNACMessage{}, err
	case u != nil && bytes.Equal(u.PassHash, md5Hash):
		// password check succeeded
		loginOK = true
	case s.config.DisableAuth:
		// login failed but let them in anyway
		user, err := newUserFn(screenName)
		if err != nil {
			return oscar.SNACMessage{}, err
		}
		if err := s.userManager.UpsertUser(user); err != nil {
			return oscar.SNACMessage{}, err
		}
		loginOK = true
	}

	snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
	snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVScreenName, screenName))

	if loginOK {
		sess := s.sessionManager.AddSession(newUUIDFn().String(), screenName)
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVReconnectHere, server.Address(s.config.OSCARHost, s.config.BOSPort)),
			oscar.NewTLV(oscar.TLVAuthorizationCookie, sess.ID()),
		})
	} else {
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
		})
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPLoginResponse,
		},
		Body: snacPayloadOut,
	}, nil
}
