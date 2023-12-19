package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type AuthHandler interface {
	BUCPChallengeRequestHandler(bodyIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (oscar.SNACMessage, error)
	BUCPLoginRequestHandler(bodyIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID, fn func(screenName string) (state.User, error)) (oscar.SNACMessage, error)
	RetrieveChatSession(chatID string, sessionID string) (*state.Session, error)
	RetrieveBOSSession(sessionID string) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session, chatID string) error
}

type AuthService struct {
	AuthHandler
	Config
	RouteLogger
}

func SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FLAPSignonFrame, error) {
	flapFrameOut := oscar.FLAPFrame{
		StartMarker:   42,
		FrameType:     oscar.FLAPFrameSignon,
		Sequence:      uint16(*sequence),
		PayloadLength: 4, // size of FLAPSignonFrame
	}
	if err := oscar.Marshal(flapFrameOut, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	flapSignonFrameOut := oscar.FLAPSignonFrame{
		FLAPVersion: 1,
	}
	if err := oscar.Marshal(flapSignonFrameOut, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}

	// receive
	flapFrameIn := oscar.FLAPFrame{}
	if err := oscar.Unmarshal(&flapFrameIn, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	b := make([]byte, flapFrameIn.PayloadLength)
	if _, err := rw.Read(b); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	flapSignonFrameIn := oscar.FLAPSignonFrame{}
	if err := oscar.Unmarshal(&flapSignonFrameIn, bytes.NewBuffer(b)); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}

	*sequence++

	return flapSignonFrameIn, nil
}

func verifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error) {
	seq := uint32(100)

	flap, err := SendAndReceiveSignonFrame(rw, &seq)
	if err != nil {
		return nil, 0, err
	}

	var ok bool
	buf, ok := flap.GetSlice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return nil, 0, errors.New("unable to get session id from payload")
	}

	cookie := ChatCookie{}
	err = oscar.Unmarshal(&cookie, bytes.NewBuffer(buf))

	return &cookie, seq, err
}

func (rt AuthService) handleAuthConnection(rwc io.ReadWriteCloser) {
	defer rwc.Close()
	seq := uint32(100)
	_, err := SendAndReceiveSignonFrame(rwc, &seq)
	if err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	flap := oscar.FLAPFrame{}
	if err := oscar.Unmarshal(&flap, rwc); err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	b := make([]byte, flap.PayloadLength)
	if _, err := rwc.Read(b); err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	snac := oscar.SNACFrame{}
	buf := bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	bodyIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := oscar.Unmarshal(&bodyIn, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	msg, err := rt.BUCPChallengeRequestHandler(bodyIn, uuid.New)
	if err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	flap = oscar.FLAPFrame{}
	if err := oscar.Unmarshal(&flap, rwc); err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	snac = oscar.SNACFrame{}
	b = make([]byte, flap.PayloadLength)
	if _, err := rwc.Read(b); err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	buf = bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	bodyIn2 := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&bodyIn2, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	msg, err = rt.BUCPLoginRequestHandler(bodyIn2, uuid.New, state.NewStubUser)
	if err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		rt.Logger.Error(err.Error())
		return
	}
}

func (rt AuthService) Start() {
	addr := Address("", rt.Config.OSCARPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind OSCAR server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting auth service", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			rt.Logger.Error(err.Error())
			continue
		}

		go rt.handleAuthConnection(conn)
	}
}
