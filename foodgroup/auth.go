package foodgroup

import (
	"bytes"
	"context"
	"errors"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
)

// NewAuthService creates a new instance of AuthService.
func NewAuthService(cfg config.Config,
	sessionManager SessionManager,
	messageRelayer MessageRelayer,
	feedbagManager FeedbagManager,
	userManager UserManager,
	chatRegistry ChatRegistry,
	legacyBuddyListManager LegacyBuddyListManager) *AuthService {
	return &AuthService{
		chatRegistry:           chatRegistry,
		config:                 cfg,
		feedbagManager:         feedbagManager,
		legacyBuddyListManager: legacyBuddyListManager,
		messageRelayer:         messageRelayer,
		sessionManager:         sessionManager,
		userManager:            userManager,
	}
}

// AuthService provides user BUCP login and session management services.
type AuthService struct {
	chatRegistry           ChatRegistry
	config                 config.Config
	feedbagManager         FeedbagManager
	legacyBuddyListManager LegacyBuddyListManager
	messageRelayer         MessageRelayer
	sessionManager         SessionManager
	userManager            UserManager
}

// RetrieveChatSession returns a chat room session. Return nil if the session
// does not exist.
func (s AuthService) RetrieveChatSession(loginCookie []byte) (*state.Session, error) {
	c := chatLoginCookie{}
	if err := wire.Unmarshal(&c, bytes.NewBuffer(loginCookie)); err != nil {
		return nil, err
	}
	_, chatSessMgr, err := s.chatRegistry.Retrieve(c.Cookie)
	if err != nil {
		return nil, err
	}
	return chatSessMgr.(SessionManager).RetrieveSession(c.SessID), nil
}

// RetrieveBOSSession returns a user's session. Return nil if the session does
// not exist.
func (s AuthService) RetrieveBOSSession(sessionID string) (*state.Session, error) {
	return s.sessionManager.RetrieveSession(sessionID), nil
}

// Signout removes user from the BOS server and notifies adjacent users (those
// who have this user's screen name on their buddy list) of their departure.
func (s AuthService) Signout(ctx context.Context, sess *state.Session) error {
	if err := broadcastDeparture(ctx, sess, s.messageRelayer, s.feedbagManager, s.legacyBuddyListManager); err != nil {
		return err
	}
	s.sessionManager.RemoveSession(sess)
	s.legacyBuddyListManager.DeleteUser(sess.ScreenName())
	return nil
}

// SignoutChat removes user from chat room and notifies remaining participants
// of their departure. If user is the last to leave, the chat room is deleted.
func (s AuthService) SignoutChat(ctx context.Context, sess *state.Session) error {
	chatRoom, chatSessMgr, err := s.chatRegistry.Retrieve(sess.ChatRoomCookie())
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

// BUCPChallengeRequest satisfies the client request for a random auth
// key. It returns SNAC wire.BUCPChallengeResponse. If the screen name is
// recognized as a valid user, the response contains the account's auth key,
// which salts the user's MD5 password hash. If the account is invalid, an
// error code is set in TLV wire.TLVErrorSubcode. If login credentials are
// invalid and app config DisableAuth is true, a stub auth key is generated and
// a successful challenge response is returned.
func (s AuthService) BUCPChallengeRequest(bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUIDFn func() uuid.UUID) (wire.SNACMessage, error) {
	screenName, exists := bodyIn.String(wire.TLVScreenName)
	if !exists {
		return wire.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	u, err := s.userManager.User(screenName)
	switch {
	case err != nil:
		return wire.SNACMessage{}, err
	case u != nil:
		// user lookup succeeded
		authKey = u.AuthKey
	case s.config.DisableAuth:
		// can't find user, generate stub auth key
		authKey = newUUIDFn().String()
	default:
		// can't find user, return login error
		snacFrameOut := wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPLoginResponse,
		}
		snacPayloadOut := wire.SNAC_0x17_0x03_BUCPLoginResponse{}
		snacPayloadOut.Append(wire.NewTLV(wire.TLVErrorSubcode, uint16(0x01)))
		return wire.SNACMessage{
			Frame: snacFrameOut,
			Body:  snacPayloadOut,
		}, nil
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPChallengeResponse,
		},
		Body: wire.SNAC_0x17_0x07_BUCPChallengeResponse{
			AuthKey: authKey,
		},
	}, nil
}

// BUCPLoginRequest verifies user credentials. Upon successful login, a
// session is created.
// If login credentials are invalid and app config DisableAuth is true, a stub
// user is created and login continues as normal. DisableAuth allows you to
// skip the account creation procedure, which simplifies the login flow during
// development.
// If login is successful, the SNAC TLV list contains the BOS server address
// (wire.TLVReconnectHere) and an authorization cookie
// (wire.TLVAuthorizationCookie). Else, an error code is set
// (wire.TLVErrorSubcode).
func (s AuthService) BUCPLoginRequest(bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUUIDFn func() uuid.UUID, newUserFn func(screenName string) (state.User, error)) (wire.SNACMessage, error) {
	screenName, found := bodyIn.String(wire.TLVScreenName)
	if !found {
		return wire.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}
	md5Hash, found := bodyIn.Slice(wire.TLVPasswordHash)
	if !found {
		return wire.SNACMessage{}, errors.New("password hash doesn't exist in tlv")
	}

	loginOK := false

	u, err := s.userManager.User(screenName)
	switch {
	// runtime error
	case err != nil:
		return wire.SNACMessage{}, err
	// user exists, check password hashes.
	// check both strong password hash (for AIM 4.8+) and weak password hash
	// (for AIM < 4.8).
	// in the future, this could check the appropriate hash based on the client
	// version indicated in the request metadata, but more testing needs to be
	// done first to make sure versioning metadata is consistent across all AIM
	// clients, including 3rd-party implementations, lest we create edge cases
	// that break login for some clients.
	case u != nil && (bytes.Equal(u.StrongMD5Pass, md5Hash) || bytes.Equal(u.WeakMD5Pass, md5Hash)):
		loginOK = true
	// authentication check is disabled, allow unconditional login. create new
	// user if the account doesn't already exist.
	case s.config.DisableAuth:
		user, err := newUserFn(screenName)
		if err != nil {
			return wire.SNACMessage{}, err
		}
		if err := s.userManager.InsertUser(user); err != nil {
			return wire.SNACMessage{}, err
		}
		loginOK = true
	}

	snacPayloadOut := wire.SNAC_0x17_0x03_BUCPLoginResponse{}
	snacPayloadOut.Append(wire.NewTLV(wire.TLVScreenName, screenName))

	if loginOK {
		sess := s.sessionManager.AddSession(newUUIDFn().String(), screenName)
		snacPayloadOut.AppendList([]wire.TLV{
			wire.NewTLV(wire.TLVReconnectHere, config.Address(s.config.OSCARHost, s.config.BOSPort)),
			wire.NewTLV(wire.TLVAuthorizationCookie, sess.ID()),
		})
	} else {
		snacPayloadOut.AppendList([]wire.TLV{
			wire.NewTLV(wire.TLVErrorSubcode, uint16(0x01)),
		})
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPLoginResponse,
		},
		Body: snacPayloadOut,
	}, nil
}
