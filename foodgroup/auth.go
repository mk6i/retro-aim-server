package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"net"

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
	legacyBuddyListManager LegacyBuddyListManager,
) *AuthService {
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

// AuthService provides client login and session management services. It
// supports both FLAP (AIM v1.0-v3.0) and BUCP (AIM v3.5-v5.9) authentication
// modes.
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

// Signout removes this user's session and notifies users who have this user on
// their buddy list about this user's departure.
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

// BUCPChallenge processes a BUCP authentication challenge request. It
// retrieves the user's auth key based on the screen name provided in the
// request. The client uses the auth key to salt the MD5 password hash provided
// in the subsequent login request. If the account is invalid, an error code is
// set in TLV wire.LoginTLVTagsErrorSubcode. If login credentials are invalid and app
// config DisableAuth is true, a stub auth key is generated and a successful
// challenge response is returned.
func (s AuthService) BUCPChallenge(
	bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest,
	newUUIDFn func() uuid.UUID,
) (wire.SNACMessage, error) {

	screenName, exists := bodyIn.String(wire.LoginTLVTagsScreenName)
	if !exists {
		return wire.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	user, err := s.userManager.User(screenName)
	if err != nil {
		return wire.SNACMessage{}, err
	}

	switch {
	case user != nil:
		// user lookup succeeded
		authKey = user.AuthKey
	case s.config.DisableAuth:
		// can't find user, generate stub auth key
		authKey = newUUIDFn().String()
	default:
		// can't find user, return login error
		return wire.SNACMessage{
			Frame: wire.SNACFrame{
				FoodGroup: wire.BUCP,
				SubGroup:  wire.BUCPLoginResponse,
			},
			Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: []wire.TLV{
						wire.NewTLV(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidUsernameOrPassword),
					},
				},
			},
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

// BUCPLogin processes a BUCP authentication request for AIM v3.5-v5.9. Upon
// successful login, a session is created.
// If login credentials are invalid and app config DisableAuth is true, a stub
// user is created and login continues as normal. DisableAuth allows you to
// skip the account creation procedure, which simplifies the login flow during
// development.
// If login is successful, the SNAC TLV list contains the BOS server address
// (wire.LoginTLVTagsReconnectHere) and an authorization cookie
// (wire.LoginTLVTagsAuthorizationCookie). Else, an error code is set
// (wire.LoginTLVTagsErrorSubcode).
func (s AuthService) BUCPLogin(
	bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest,
	newUUIDFn func() uuid.UUID,
	newUserFn func(screenName string) (state.User, error),
) (wire.SNACMessage, error) {

	block, err := s.login(bodyIn.TLVList, newUserFn, newUUIDFn)
	if err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.BUCP,
			SubGroup:  wire.BUCPLoginResponse,
		},
		Body: wire.SNAC_0x17_0x03_BUCPLoginResponse{
			TLVRestBlock: block,
		},
	}, nil
}

// FLAPLogin processes a FLAP authentication request for AIM v1.0-v3.0. Upon
// successful login, a session is created.
// If login credentials are invalid and app config DisableAuth is true, a stub
// user is created and login continues as normal. DisableAuth allows you to
// skip the account creation procedure, which simplifies the login flow during
// development.
// If login is successful, the SNAC TLV list contains the BOS server address
// (wire.LoginTLVTagsReconnectHere) and an authorization cookie
// (wire.LoginTLVTagsAuthorizationCookie). Else, an error code is set
// (wire.LoginTLVTagsErrorSubcode).
func (s AuthService) FLAPLogin(
	frame wire.FLAPSignonFrame,
	newUUIDFn func() uuid.UUID,
	newUserFn func(screenName string) (state.User, error),
) (wire.TLVRestBlock, error) {
	return s.login(frame.TLVList, newUserFn, newUUIDFn)
}

// login validates a user's credentials and creates their session. it returns
// metadata used in both BUCP and FLAP authentication responses.
func (s AuthService) login(
	TLVList wire.TLVList,
	newUserFn func(screenName string) (state.User, error),
	newUUIDFn func() uuid.UUID,
) (wire.TLVRestBlock, error) {

	screenName, found := TLVList.String(wire.LoginTLVTagsScreenName)
	if !found {
		return wire.TLVRestBlock{}, errors.New("screen name doesn't exist in tlv")
	}

	user, err := s.userManager.User(screenName)
	if err != nil {
		return wire.TLVRestBlock{}, err
	}

	var loginOK bool
	if user != nil {
		if md5Hash, hasMD5 := TLVList.Slice(wire.LoginTLVTagsPasswordHash); hasMD5 {
			loginOK = user.ValidateHash(md5Hash)
		} else if roastedPass, hasRoasted := TLVList.Slice(wire.LoginTLVTagsRoastedPassword); hasRoasted {
			loginOK = user.ValidateRoastedPass(roastedPass)
		} else {
			return wire.TLVRestBlock{}, errors.New("password hash doesn't exist in tlv")
		}
	}

	if loginOK || s.config.DisableAuth {
		if !loginOK {
			// make login succeed anyway. create new user if the account
			// doesn't already exist.
			newUser, err := newUserFn(screenName)
			if err != nil {
				return wire.TLVRestBlock{}, err
			}
			if err := s.userManager.InsertUser(newUser); err != nil {
				if !errors.Is(err, state.ErrDupUser) {
					return wire.TLVRestBlock{}, err
				}
			}
		}

		sess := s.sessionManager.AddSession(newUUIDFn().String(), screenName)

		// auth success
		return wire.TLVRestBlock{
			TLVList: []wire.TLV{
				wire.NewTLV(wire.LoginTLVTagsScreenName, screenName),
				wire.NewTLV(wire.LoginTLVTagsReconnectHere, net.JoinHostPort(s.config.OSCARHost, s.config.BOSPort)),
				wire.NewTLV(wire.LoginTLVTagsAuthorizationCookie, sess.ID()),
			},
		}, nil
	}

	// auth failure
	return wire.TLVRestBlock{
		TLVList: []wire.TLV{
			wire.NewTLV(wire.LoginTLVTagsScreenName, screenName),
			wire.NewTLV(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidUsernameOrPassword),
		},
	}, nil
}
