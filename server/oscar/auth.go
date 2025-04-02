package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
)

type AuthService interface {
	BUCPChallenge(ctx context.Context, bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(ctx context.Context, bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	FLAPLogin(ctx context.Context, frame wire.FLAPSignonFrame, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.TLVRestBlock, error)
	RegisterBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RetrieveBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RegisterChatSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session)
	SignoutChat(ctx context.Context, sess *state.Session)
}

// AuthServer is an authentication server for both FLAP (AIM v1.0-3.0) and BUCP
// (AIM v3.5-5.9) authentication flows.
type AuthServer struct {
	AuthService
	config.Config
	Logger *slog.Logger
}

// Start starts the authentication server and listens for new connections.
func (rt AuthServer) Start(ctx context.Context) error {
	addr := net.JoinHostPort("", rt.Config.AuthPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("unable to start auth server: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	rt.Logger.Info("starting server", "listen_host", addr, "oscar_host", rt.Config.OSCARHost)

	wg := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			rt.Logger.Error("accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			connCtx := context.WithValue(ctx, "ip", conn.RemoteAddr().String())
			rt.Logger.DebugContext(connCtx, "accepted connection")
			if err := rt.handleNewConnection(connCtx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}

	if !waitForShutdown(&wg) {
		rt.Logger.Error("shutdown complete, but connections didn't close cleanly")
	} else {
		rt.Logger.Info("shutdown complete")
	}

	return nil
}

func (rt AuthServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	defer rwc.Close()

	flapc := wire.NewFlapClient(100, rwc, rwc)
	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	signonFrame, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	// decide whether the client is using BUCP or FLAP authentication based on
	// the presence of the screen name TLV. this block used to check for the
	// presence of the roasted password TLV, however that proved an unreliable
	// indicator of FLAP-auth because older ICQ clients appear to omit the
	// roasted password TLV when the password is not stored client-side.
	if _, hasScreenName := signonFrame.Uint16BE(wire.LoginTLVTagsScreenName); hasScreenName {
		return rt.processFLAPAuth(ctx, signonFrame, flapc)
	}

	return rt.processBUCPAuth(ctx, flapc)
}

func (rt AuthServer) processFLAPAuth(ctx context.Context, signonFrame wire.FLAPSignonFrame, flapc *wire.FlapClient) error {
	tlv, err := rt.AuthService.FLAPLogin(ctx, signonFrame, state.NewStubUser)
	if err != nil {
		return err
	}
	return flapc.SendSignoffFrame(tlv)
}

func (rt AuthServer) processBUCPAuth(ctx context.Context, flapc *wire.FlapClient) error {
	for {
		frame, err := flapc.ReceiveFLAP()
		if err != nil {
			return err
		}

		switch frame.FrameType {
		case wire.FLAPFrameSignoff:
			rt.Logger.Debug("signed off mid-login")
			return io.EOF // client disconnected
		case wire.FLAPFrameKeepAlive:
			rt.Logger.Debug("received flap keepalive frame")
		case wire.FLAPFrameData:
			buf := bytes.NewReader(frame.Payload)
			fr := wire.SNACFrame{}
			if err := wire.UnmarshalBE(&fr, buf); err != nil {
				return err
			}
			switch {
			case fr.FoodGroup == wire.BUCP && fr.SubGroup == wire.BUCPChallengeRequest:
				challengeRequest := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
				if err := wire.UnmarshalBE(&challengeRequest, buf); err != nil {
					return err
				}
				outSNAC, err := rt.BUCPChallenge(ctx, challengeRequest, uuid.New)
				if err != nil {
					return err
				}
				if err := flapc.SendSNAC(outSNAC.Frame, outSNAC.Body); err != nil {
					return err
				}

				if outSNAC.Frame.SubGroup == wire.BUCPLoginResponse {
					screenName, _ := challengeRequest.String(wire.LoginTLVTagsScreenName)
					rt.Logger.Debug("failed BUCP challenge: user does not exist", "screen_name", screenName)
					return nil // account does not exist
				}
			case fr.FoodGroup == wire.BUCP && fr.SubGroup == wire.BUCPLoginRequest:
				loginRequest := wire.SNAC_0x17_0x02_BUCPLoginRequest{}
				if err := wire.UnmarshalBE(&loginRequest, buf); err != nil {
					return err
				}
				outSNAC, err := rt.BUCPLogin(ctx, loginRequest, state.NewStubUser)
				if err != nil {
					return err
				}

				return flapc.SendSNAC(outSNAC.Frame, outSNAC.Body)
			default:
				rt.Logger.Debug("unexpected SNAC received during login",
					"foodgroup", wire.FoodGroupName(fr.FoodGroup),
					"subgroup", wire.SubGroupName(fr.FoodGroup, fr.SubGroup))
				return io.EOF
			}
		default:
			rt.Logger.Debug("unexpected frame type received during login", "type", frame.FrameType)
			return io.EOF
		}
	}
}
