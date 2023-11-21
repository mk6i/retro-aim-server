package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
	"io"
	"net"
)

func NewAuthService(cfg server.Config, sm SessionManager, fm FeedbagManager, um UserManager, cr *state.ChatRegistry) *AuthService {
	return &AuthService{
		sessionManager: sm,
		feedbagManager: fm,
		config:         cfg,
		userManager:    um,
		chatRegistry:   cr,
	}
}

type AuthService struct {
	sessionManager SessionManager
	feedbagManager FeedbagManager
	userManager    UserManager
	config         server.Config
	chatRegistry   *state.ChatRegistry
}

func (s AuthService) RetrieveChatSession(ctx context.Context, chatID string, sessID string) (*state.Session, error) {
	_, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		return nil, err
	}
	chatSess, found := chatSessMgr.(ChatSessionManager).Retrieve(sessID)
	if !found {
		return nil, fmt.Errorf("unable to find user for session. chat id: %s, sess id: %s", chatID, sessID)
	}
	return chatSess, nil
}

func (s AuthService) Signout(ctx context.Context, sess *state.Session) error {
	if err := broadcastDeparture(ctx, sess, s.sessionManager, s.feedbagManager); err != nil {
		return err
	}
	s.sessionManager.Remove(sess)
	return nil
}

func (s AuthService) SignoutChat(ctx context.Context, sess *state.Session, chatID string) {
	chatRoom, chatSessMgr, err := s.chatRegistry.Retrieve(chatID)
	if err != nil {
		fmt.Println("error getting chat room to remove")
		return
	}
	alertUserLeft(ctx, sess, chatSessMgr.(ChatSessionManager))
	chatSessMgr.(ChatSessionManager).Remove(sess)
	if chatSessMgr.(ChatSessionManager).Empty() {
		s.chatRegistry.RemoveRoom(chatRoom.Cookie)
	}
}

func (s AuthService) VerifyLogin(conn net.Conn) (*state.Session, uint32, error) {
	seq := uint32(100)

	flap, err := s.SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	ID, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	sess, ok := s.sessionManager.Retrieve(string(ID))
	if !ok {
		return nil, 0, fmt.Errorf("unable to find session by id %s", ID)
	}

	return sess, seq, nil
}

func (s AuthService) SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FlapSignonFrame, error) {
	flapFrameOut := oscar.FlapFrame{
		StartMarker:   42,
		FrameType:     oscar.FlapFrameSignon,
		Sequence:      uint16(*sequence),
		PayloadLength: 4, // size of FlapSignonFrame
	}
	if err := oscar.Marshal(flapFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameOut := oscar.FlapSignonFrame{
		FlapVersion: 1,
	}
	if err := oscar.Marshal(flapSignonFrameOut, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	// receive
	flapFrameIn := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flapFrameIn, rw); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	b := make([]byte, flapFrameIn.PayloadLength)
	if _, err := rw.Read(b); err != nil {
		return oscar.FlapSignonFrame{}, err
	}
	flapSignonFrameIn := oscar.FlapSignonFrame{}
	if err := oscar.Unmarshal(&flapSignonFrameIn, bytes.NewBuffer(b)); err != nil {
		return oscar.FlapSignonFrame{}, err
	}

	*sequence++

	return flapSignonFrameIn, nil
}

func (s AuthService) VerifyChatLogin(rw io.ReadWriter) (*server.ChatCookie, uint32, error) {
	seq := uint32(100)

	flap, err := s.SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	buf, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	cookie := server.ChatCookie{}
	err = oscar.Unmarshal(&cookie, bytes.NewBuffer(buf))

	return &cookie, seq, err
}

func (s AuthService) ReceiveAndSendAuthChallenge(snacPayloadIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (oscar.XMessage, error) {
	screenName, exists := snacPayloadIn.GetString(oscar.TLVScreenName)
	if !exists {
		return oscar.XMessage{}, errors.New("screen name doesn't exist in tlv")
	}

	var authKey string

	u, err := s.userManager.GetUser(screenName)
	switch {
	case err != nil:
		return oscar.XMessage{}, err
	case u != nil:
		// user lookup succeeded
		authKey = u.AuthKey
	case s.config.DisableAuth:
		// can't find user, generate stub auth key
		authKey = newUUID().String()
	default:
		// can't find user, return login error
		snacFrameOut := oscar.SnacFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPLoginResponse,
		}
		snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
		snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)))
		return oscar.XMessage{
			SnacFrame: snacFrameOut,
			SnacOut:   snacPayloadOut,
		}, nil
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPChallengeResponse,
		},
		SnacOut: oscar.SNAC_0x17_0x07_BUCPChallengeResponse{
			AuthKey: authKey,
		},
	}, nil
}

func (s AuthService) ReceiveAndSendBUCPLoginRequest(snacPayloadIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID) (oscar.XMessage, error) {

	screenName, found := snacPayloadIn.GetString(oscar.TLVScreenName)
	if !found {
		return oscar.XMessage{}, errors.New("screen name doesn't exist in tlv")
	}
	md5Hash, found := snacPayloadIn.GetSlice(oscar.TLVPasswordHash)
	if !found {
		return oscar.XMessage{}, errors.New("password hash doesn't exist in tlv")
	}

	loginOK := false

	u, err := s.userManager.GetUser(screenName)
	switch {
	case err != nil:
		return oscar.XMessage{}, err
	case u != nil && bytes.Equal(u.PassHash, md5Hash):
		// password check succeeded
		loginOK = true
	case s.config.DisableAuth:
		// login failed but let them in anyway
		newUser, err := newStubUser(screenName)
		if err != nil {
			return oscar.XMessage{}, err
		}
		if err := s.userManager.UpsertUser(newUser); err != nil {
			return oscar.XMessage{}, err
		}
		loginOK = true
	}

	snacPayloadOut := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
	snacPayloadOut.AddTLV(oscar.NewTLV(oscar.TLVScreenName, screenName))

	if loginOK {
		sess := s.sessionManager.NewSessionWithSN(newUUID().String(), screenName)
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVReconnectHere, server.Address(s.config.OSCARHost, s.config.BOSPort)),
			oscar.NewTLV(oscar.TLVAuthorizationCookie, sess.ID()),
		})
	} else {
		snacPayloadOut.AddTLVList([]oscar.TLV{
			oscar.NewTLV(oscar.TLVErrorSubcode, uint16(0x01)),
		})
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUCP,
			SubGroup:  oscar.BUCPLoginResponse,
		},
		SnacOut: snacPayloadOut,
	}, nil
}

func newStubUser(screenName string) (state.User, error) {
	u := state.User{ScreenName: screenName}

	uid, err := uuid.NewRandom()
	if err != nil {
		return u, err
	}
	u.AuthKey = uid.String()

	if err := u.HashPassword("welcome1"); err != nil {
		return u, err
	}
	return u, u.HashPassword("welcome1")
}
