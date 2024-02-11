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
	BUCPChallengeRequest(bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLoginRequest(bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUUID func() uuid.UUID, fn func(screenName string) (state.User, error)) (wire.SNACMessage, error)
	RetrieveChatSession(loginCookie []byte) (*state.Session, error)
	RetrieveBOSSession(sessionID string) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session) error
	SignoutChat(ctx context.Context, sess *state.Session) error
}

// BUCPAuthService represents a service that implements the OSCAR BUCP. This is
// the first service that the AIM client connects to in the login flow.
type BUCPAuthService struct {
	AuthService
	config.Config
	Logger *slog.Logger
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
	flapc := flapClient{
		r:        rwc,
		sequence: 100,
		w:        rwc,
	}
	defer rwc.Close()
	if _, err := flapc.SignonHandshake(); err != nil {
		return err
	}

	challengeRequest := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := flapc.ReceiveSNAC(&wire.SNACFrame{}, &challengeRequest); err != nil {
		return err
	}

	outSNAC, err := rt.BUCPChallengeRequest(challengeRequest, uuid.New)
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

	outSNAC, err = rt.BUCPLoginRequest(loginRequest, uuid.New, state.NewStubUser)
	if err != nil {
		return err
	}

	return flapc.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
