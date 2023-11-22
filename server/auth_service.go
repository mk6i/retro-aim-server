package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type AuthHandler interface {
	ReceiveAndSendAuthChallenge(snacPayloadIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (oscar.SNACMessage, error)
	ReceiveAndSendBUCPLoginRequest(snacPayloadIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID) (oscar.SNACMessage, error)
	RetrieveChatSession(ctx context.Context, chatID string, sessID string) (*state.Session, error)
	SendAndReceiveSignonFrame(rw io.ReadWriter, sequence *uint32) (oscar.FLAPSignonFrame, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session, chatID string)
	VerifyChatLogin(rw io.ReadWriter) (*ChatCookie, uint32, error)
	VerifyLogin(rwc io.ReadWriteCloser) (*state.Session, uint32, error)
}

type AuthService struct {
	AuthHandler
	Config
	RouteLogger
}

func (rt AuthService) handleAuthConnection(rwc io.ReadWriteCloser) {
	defer rwc.Close()
	seq := uint32(100)
	_, err := rt.SendAndReceiveSignonFrame(rwc, &seq)
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

	snacPayloadIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	msg, err := rt.ReceiveAndSendAuthChallenge(snacPayloadIn, uuid.New)
	if err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	if err := sendSNAC(oscar.SNACFrame{}, msg.Frame, msg.Body, &seq, rwc); err != nil {
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

	snacPayloadIn2 := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn2, buf); err != nil {
		rt.Logger.Error(err.Error())
		return
	}

	msg, err = rt.ReceiveAndSendBUCPLoginRequest(snacPayloadIn2, uuid.New)
	if err != nil {
		rt.Logger.Error(err.Error())
		return
	}
	if err := sendSNAC(oscar.SNACFrame{}, msg.Frame, msg.Body, &seq, rwc); err != nil {
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
