package oscar

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
)

type AuthService interface {
	BUCPChallenge(bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	FLAPLogin(frame wire.FLAPSignonFrame, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.TLVRestBlock, error)
	RegisterBOSSession(screenName state.IdentScreenName) (*state.Session, error)
	RegisterChatSession(loginCookie []byte) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session) error
}

// AuthServer is an authentication server for both FLAP (AIM v1.0-3.0) and BUCP
// (AIM v3.5-5.9) authentication flows.
type AuthServer struct {
	AuthService
	config.Config
	Logger *slog.Logger
	CookieCracker
}

// Start starts the authentication server and listens for new connections.
func (rt AuthServer) Start() {
	addr := net.JoinHostPort("", rt.Config.AuthPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind server address", "host", addr, "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting server", "listen_host", addr, "oscar_host", rt.Config.OSCARHost)

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

func (rt AuthServer) handleNewConnection(rwc io.ReadWriteCloser) error {
	defer rwc.Close()

	flapc := wire.NewFlapClient(100, rwc, rwc)
	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	signonFrame, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	if _, hasRoastedPassword := signonFrame.Uint16(wire.LoginTLVTagsRoastedPassword); hasRoastedPassword {
		return rt.processFLAPAuth(signonFrame, flapc)
	}

	return rt.processBUCPAuth(flapc, err)
}

func (rt AuthServer) processFLAPAuth(signonFrame wire.FLAPSignonFrame, flapc *wire.FlapClient) error {
	tlv, err := rt.AuthService.FLAPLogin(signonFrame, state.NewStubUser)
	if err != nil {
		return err
	}
	return flapc.SendSignoffFrame(tlv)
}

func (rt AuthServer) processBUCPAuth(flapc *wire.FlapClient, err error) error {
	challengeRequest := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := flapc.ReceiveSNAC(&wire.SNACFrame{}, &challengeRequest); err != nil {
		return err
	}

	outSNAC, err := rt.BUCPChallenge(challengeRequest, uuid.New)
	if err != nil {
		return err
	}
	if err := flapc.SendSNAC(outSNAC.Frame, outSNAC.Body); err != nil {
		return err
	}

	loginRequest := wire.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := flapc.ReceiveSNAC(&wire.SNACFrame{}, &loginRequest); err != nil {
		return err
	}

	outSNAC, err = rt.BUCPLogin(loginRequest, state.NewStubUser)
	if err != nil {
		return err
	}

	return flapc.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
