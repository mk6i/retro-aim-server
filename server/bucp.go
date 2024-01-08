package server

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

type AuthHandler interface {
	BUCPChallengeRequestHandler(bodyIn oscar.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (oscar.SNACMessage, error)
	BUCPLoginRequestHandler(bodyIn oscar.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID, fn func(screenName string) (state.User, error)) (oscar.SNACMessage, error)
	RetrieveChatSession(chatID string, sessionID string) (*state.Session, error)
	RetrieveBOSSession(sessionID string) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session, chatID string) error
}

// BUCPAuthService represents a service that implements the OSCAR BUCP. This is
// the first service that the AIM client connects to in the login flow.
type BUCPAuthService struct {
	AuthHandler
	config.Config
	RouteLogger
}

// Start creates a TCP server that implements the BUCP authentication flow. It
// validates users credentials and, upon success, provides an auth cookie and
// hostname information for connecting to the BOS service.
func (rt BUCPAuthService) Start() {
	addr := config.Address("", rt.Config.BUCPPort)
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

		go func() {
			if err := rt.handleNewConnection(conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}
}

func (rt BUCPAuthService) handleNewConnection(rwc io.ReadWriteCloser) error {
	defer rwc.Close()
	seq := uint32(100)
	if _, err := flapSignonHandshake(rwc, &seq); err != nil {
		return err
	}

	challengeRequest := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := receiveSNAC(&oscar.SNACFrame{}, &challengeRequest, rwc); err != nil {
		return err
	}

	outSNAC, err := rt.BUCPChallengeRequestHandler(challengeRequest, uuid.New)
	if err != nil {
		return err
	}
	if err := sendSNAC(outSNAC.Frame, outSNAC.Body, &seq, rwc); err != nil {
		return err
	}

	loginRequest := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := receiveSNAC(&oscar.SNACFrame{}, &loginRequest, rwc); err != nil {
		return err
	}

	outSNAC, err = rt.BUCPLoginRequestHandler(loginRequest, uuid.New, state.NewStubUser)
	if err != nil {
		return err
	}

	return sendSNAC(outSNAC.Frame, outSNAC.Body, &seq, rwc)
}

func flapSignonHandshake(rw io.ReadWriter, sequence *uint32) (oscar.FLAPSignonFrame, error) {
	// send FLAPFrameSignon to client
	flap := oscar.FLAPFrame{
		StartMarker:   42,
		FrameType:     oscar.FLAPFrameSignon,
		Sequence:      uint16(*sequence),
		PayloadLength: 4, // size of FLAPSignonFrame
	}
	if err := oscar.Marshal(flap, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	flapSignonFrameOut := oscar.FLAPSignonFrame{
		FLAPVersion: 1,
	}
	if err := oscar.Marshal(flapSignonFrameOut, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}

	// receive FLAPFrameSignon from client
	flap = oscar.FLAPFrame{}
	if err := oscar.Unmarshal(&flap, rw); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	buf, err := flap.ReadBody(rw)
	if err != nil {
		return oscar.FLAPSignonFrame{}, err
	}
	flapSignonFrameIn := oscar.FLAPSignonFrame{}
	if err := oscar.Unmarshal(&flapSignonFrameIn, buf); err != nil {
		return oscar.FLAPSignonFrame{}, err
	}

	*sequence++

	return flapSignonFrameIn, nil
}
