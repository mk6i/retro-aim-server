package oscar

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// AlertServer provides client connection lifecycle management for the Alert
// service. This server, whose handlers are all no-op, exists solely to satisfy
// AIM 4.x, which throws an error when it can't connect to the alert service.
type AlertServer struct {
	AuthService
	Handler
	Logger *slog.Logger
	OnlineNotifier
	config.Config
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to Handler.
func (rt AlertServer) Start() {
	addr := config.Address("", rt.Config.AlertPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind ALERT server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting ALERT service", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			rt.Logger.Error(err.Error())
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		rt.Logger.DebugContext(ctx, "accepted connection")
		go func() {
			if err := rt.handleNewConnection(ctx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}
}

func (rt AlertServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	flapc := &flapClient{
		r:        rwc,
		sequence: 100,
		w:        rwc,
	}

	flap, err := flapc.SignonHandshake()
	if err != nil {
		return err
	}

	var ok bool
	sessionID, ok := flap.Slice(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	bosSess, err := rt.RetrieveBOSSession(string(sessionID))
	if err != nil {
		return err
	}
	if bosSess == nil {
		return errors.New("session not found")
	}

	defer func() {
		bosSess.Close()
		rwc.Close()
		if err := rt.Signout(ctx, bosSess); err != nil {
			rt.Logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
		}
	}()

	ctx = context.WithValue(ctx, "screenName", bosSess.ScreenName())

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	// We copy the session object here to make sure that
	// dispatchIncomingMessages does not consume relayed messages produced by
	// the BOS server. Without this hack, message consumption would be split
	// between the BOS server and Alert server, which would result in
	// incorrect sequence number generation, because each server has its own
	// sequence counter. This hack can be removed by decoupling FLAP routing
	// and message relaying, which are both performed in
	// dispatchIncomingMessages.
	sessCopy := state.NewSession()
	sessCopy.SetScreenName(bosSess.ScreenName())
	sessCopy.SetID(bosSess.ID())

	return dispatchIncomingMessages(ctx, sessCopy, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}
