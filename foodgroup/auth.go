package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
)

// NewAuthService creates a new instance of AuthService.
func NewAuthService(
	cfg config.Config,
	sessionManager SessionRegistry,
	chatSessionRegistry ChatSessionRegistry,
	userManager UserManager,
	cookieBaker CookieBaker,
	chatMessageRelayer ChatMessageRelayer,
	accountManager AccountManager,
	adminServerSessionRetriever SessionRetriever,
) *AuthService {
	return &AuthService{
		chatSessionRegistry: chatSessionRegistry,
		config:              cfg,
		cookieBaker:         cookieBaker,
		sessionManager:      sessionManager,
		userManager:         userManager,
		chatMessageRelayer:  chatMessageRelayer,
		accountManager:      accountManager,
		// hack - adminServerSessionRetriever is just used for admin server
		adminServerSessionRetriever: adminServerSessionRetriever,
	}
}

// AuthService provides client login and session management services. It
// supports both FLAP (AIM v1.0-v3.0) and BUCP (AIM v3.5-v5.9) authentication
// modes.
type AuthService struct {
	chatMessageRelayer          ChatMessageRelayer
	chatSessionRegistry         ChatSessionRegistry
	config                      config.Config
	cookieBaker                 CookieBaker
	sessionManager              SessionRegistry
	userManager                 UserManager
	accountManager              AccountManager
	adminServerSessionRetriever SessionRetriever
}

// RegisterChatSession adds a user to a chat room. The authCookie param is an
// opaque token returned by {{OServiceService.ServiceRequest}} that identifies
// the user and chat room. It returns the session object registered in the
// ChatSessionRegistry.
// This method does not verify that the user and chat room exist because it
// implicitly trusts the contents of the token signed by
// {{OServiceService.ServiceRequest}}.
func (s AuthService) RegisterChatSession(ctx context.Context, authCookie []byte) (*state.Session, error) {
	token, err := s.cookieBaker.Crack(authCookie)
	if err != nil {
		return nil, err
	}
	c := chatLoginCookie{}
	if err := wire.UnmarshalBE(&c, bytes.NewBuffer(token)); err != nil {
		return nil, err
	}
	sess, err := s.chatSessionRegistry.AddSession(ctx, c.ChatCookie, c.ScreenName)
	if err != nil {
		return nil, fmt.Errorf("AddSession: %w", err)
	}
	return sess, err
}

// bosCookie represents a token containing client metadata passed to the BOS
// service upon connection.
type bosCookie struct {
	ScreenName state.DisplayScreenName `oscar:"len_prefix=uint8"`
	ClientID   string                  `oscar:"len_prefix=uint8"`
}

// RegisterBOSSession adds a new session to the session registry.
func (s AuthService) RegisterBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error) {
	buf, err := s.cookieBaker.Crack(authCookie)
	if err != nil {
		return nil, err
	}

	c := bosCookie{}
	if err := wire.UnmarshalBE(&c, bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}

	u, err := s.userManager.User(c.ScreenName.IdentScreenName())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	sess, err := s.sessionManager.AddSession(ctx, u.DisplayScreenName)
	if err != nil {
		return nil, fmt.Errorf("AddSession: %w", err)
	}

	// set the unconfirmed user info flag if this account is unconfirmed
	if confirmed, err := s.accountManager.ConfirmStatusByName(sess.IdentScreenName()); err != nil {
		return nil, fmt.Errorf("error setting unconfirmed user flag: %w", err)
	} else if !confirmed {
		sess.SetUserInfoFlag(wire.OServiceUserFlagUnconfirmed)
	}

	// set string containing OSCAR client name and version
	sess.SetClientID(c.ClientID)

	if u.DisplayScreenName.IsUIN() {
		sess.SetUserInfoFlag(wire.OServiceUserFlagICQ)

		uin, err := strconv.Atoi(u.IdentScreenName.String())
		if err != nil {
			return nil, fmt.Errorf("error converting username to UIN: %w", err)
		}
		sess.SetUIN(uint32(uin))
	}

	return sess, nil
}

// RetrieveBOSSession returns a user's existing session
func (s AuthService) RetrieveBOSSession(authCookie []byte) (*state.Session, error) {
	buf, err := s.cookieBaker.Crack(authCookie)
	if err != nil {
		return nil, err
	}

	c := bosCookie{}
	if err := wire.UnmarshalBE(&c, bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}

	u, err := s.userManager.User(c.ScreenName.IdentScreenName())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.adminServerSessionRetriever.RetrieveSession(u.IdentScreenName), nil
}

// Signout removes this user's session and notifies users who have this user on
// their buddy list about this user's departure. It's guaranteed that the
// session is removed from the session pool.
func (s AuthService) Signout(_ context.Context, sess *state.Session) {
	s.sessionManager.RemoveSession(sess)
}

// SignoutChat removes user from chat room and notifies remaining participants
// of their departure.
func (s AuthService) SignoutChat(ctx context.Context, sess *state.Session) {
	alertUserLeft(ctx, sess, s.chatMessageRelayer)
	s.chatSessionRegistry.RemoveSession(sess)
}

// BUCPChallenge processes a BUCP authentication challenge request. It
// retrieves the user's auth key based on the screen name provided in the
// request. The client uses the auth key to salt the MD5 password hash provided
// in the subsequent login request. If the account is valid, return
// SNAC(0x17,0x07), otherwise return SNAC(0x17,0x03).
func (s AuthService) BUCPChallenge(
	bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest,
	newUUIDFn func() uuid.UUID,
) (wire.SNACMessage, error) {

	screenName, exists := bodyIn.String(wire.LoginTLVTagsScreenName)
	if !exists {
		return wire.SNACMessage{}, errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	user, err := s.userManager.User(state.NewIdentScreenName(screenName))
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
						wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrInvalidUsernameOrPassword),
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
	newUserFn func(screenName state.DisplayScreenName) (state.User, error),
) (wire.SNACMessage, error) {

	block, err := s.login(bodyIn.TLVList, newUserFn)
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
	newUserFn func(screenName state.DisplayScreenName) (state.User, error),
) (wire.TLVRestBlock, error) {
	return s.login(frame.TLVList, newUserFn)
}

// loginProperties represents the properties sent by the client at login.
type loginProperties struct {
	clientID       string
	isBUCPAuth     bool
	isFLAPJavaAuth bool
	isTOCAuth      bool
	isFLAPAuth     bool
	passwordHash   []byte
	roastedPass    []byte
	screenName     state.DisplayScreenName
}

// fromTLV creates an instance of loginProperties from a TLV list.
func (l *loginProperties) fromTLV(list wire.TLVList) error {
	// extract screen name
	if screenName, found := list.String(wire.LoginTLVTagsScreenName); found {
		l.screenName = state.DisplayScreenName(screenName)
	} else {
		return errors.New("screen name doesn't exist in tlv")
	}

	// extract client name and version
	if clientID, found := list.String(wire.LoginTLVTagsClientIdentity); found {
		l.clientID = clientID
	}

	// get the password from the appropriate TLV. older clients have a
	// roasted password, newer clients have a hashed password. ICQ may omit
	// the password TLV when logging in without saved password.
	switch {
	case list.HasTag(wire.LoginTLVTagsPasswordHash):
		// extract password hash for BUCP login
		l.passwordHash, _ = list.Bytes(wire.LoginTLVTagsPasswordHash)
		l.isBUCPAuth = true
	case list.HasTag(wire.LoginTLVTagsRoastedPassword):
		// extract roasted password for FLAP login
		l.roastedPass, _ = list.Bytes(wire.LoginTLVTagsRoastedPassword)
		if strings.HasPrefix(l.clientID, "AOL Instant Messenger (TM) version") &&
			strings.Contains(l.clientID, "for Java") {
			l.isFLAPJavaAuth = true
		} else {
			l.isFLAPAuth = true
		}
	case list.HasTag(wire.LoginTLVTagsRoastedTOCPassword):
		// extract roasted password for TOC FLAP login
		l.roastedPass, _ = list.Bytes(wire.LoginTLVTagsRoastedTOCPassword)
		l.isTOCAuth = true
	default:
		l.isFLAPAuth = true
	}

	return nil
}

// login validates a user's credentials and creates their session. it returns
// metadata used in both BUCP and FLAP authentication responses.
func (s AuthService) login(
	tlv wire.TLVList,
	newUserFn func(screenName state.DisplayScreenName) (state.User, error),
) (wire.TLVRestBlock, error) {

	props := loginProperties{}
	if err := props.fromTLV(tlv); err != nil {
		return wire.TLVRestBlock{}, err
	}

	user, err := s.userManager.User(props.screenName.IdentScreenName())
	if err != nil {
		return wire.TLVRestBlock{}, err
	}

	if user == nil {
		// user not found
		if s.config.DisableAuth {
			// auth disabled, create the user
			return s.createUser(props, newUserFn)
		}
		// auth enabled, return separate login errors for ICQ and AIM
		loginErr := wire.LoginErrInvalidUsernameOrPassword
		if props.screenName.IsUIN() {
			loginErr = wire.LoginErrICQUserErr
		}
		return loginFailureResponse(props, loginErr), nil
	}

	// check if suspended status should prevent login
	if user.SuspendedStatus > 0x0 {
		return loginFailureResponse(props, user.SuspendedStatus), nil
	}

	if s.config.DisableAuth {
		// user exists, but don't validate
		return s.loginSuccessResponse(props)
	}

	var loginOK bool
	switch {
	case props.isBUCPAuth:
		loginOK = user.ValidateHash(props.passwordHash)
	case props.isFLAPAuth:
		loginOK = user.ValidateRoastedPass(props.roastedPass)
	case props.isFLAPJavaAuth:
		loginOK = user.ValidateRoastedJavaPass(props.roastedPass)
	case props.isTOCAuth:
		loginOK = user.ValidateRoastedTOCPass(props.roastedPass)
	}

	if !loginOK {
		return loginFailureResponse(props, wire.LoginErrInvalidPassword), nil
	}

	return s.loginSuccessResponse(props)
}

func (s AuthService) createUser(
	props loginProperties,
	newUserFn func(screenName state.DisplayScreenName) (state.User, error),
) (wire.TLVRestBlock, error) {

	var err error
	if props.screenName.IsUIN() {
		err = props.screenName.ValidateUIN()
	} else {
		err = props.screenName.ValidateAIMHandle()
	}

	if err != nil {
		switch {
		case errors.Is(err, state.ErrAIMHandleInvalidFormat) || errors.Is(err, state.ErrAIMHandleLength):
			return loginFailureResponse(props, wire.LoginErrInvalidUsernameOrPassword), nil
		case errors.Is(err, state.ErrICQUINInvalidFormat):
			return loginFailureResponse(props, wire.LoginErrICQUserErr), nil
		default:
			return wire.TLVRestBlock{}, err
		}
	}

	newUser, err := newUserFn(props.screenName)
	if err != nil {
		return wire.TLVRestBlock{}, err
	}

	err = s.userManager.InsertUser(newUser)
	if err != nil {
		return wire.TLVRestBlock{}, err
	}

	return s.loginSuccessResponse(props)
}

func (s AuthService) loginSuccessResponse(props loginProperties) (wire.TLVRestBlock, error) {
	loginCookie := bosCookie{
		ScreenName: props.screenName,
		ClientID:   props.clientID,
	}

	buf := &bytes.Buffer{}
	if err := wire.MarshalBE(loginCookie, buf); err != nil {
		return wire.TLVRestBlock{}, err
	}
	cookie, err := s.cookieBaker.Issue(buf.Bytes())
	if err != nil {
		return wire.TLVRestBlock{}, fmt.Errorf("failed to issue auth cookie: %w", err)
	}

	return wire.TLVRestBlock{
		TLVList: []wire.TLV{
			wire.NewTLVBE(wire.LoginTLVTagsScreenName, props.screenName),
			wire.NewTLVBE(wire.LoginTLVTagsReconnectHere, net.JoinHostPort(s.config.OSCARHost, s.config.BOSPort)),
			wire.NewTLVBE(wire.LoginTLVTagsAuthorizationCookie, cookie),
		},
	}, nil
}

func loginFailureResponse(props loginProperties, errCode uint16) wire.TLVRestBlock {
	return wire.TLVRestBlock{
		TLVList: []wire.TLV{
			wire.NewTLVBE(wire.LoginTLVTagsScreenName, props.screenName),
			wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, errCode),
		},
	}
}
