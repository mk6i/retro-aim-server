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

// ChatNavServer provides client connection lifecycle management for the
// ChatNav service. This service is only used by AIM 4.x clients that make a
// separate ChatNav TCP connection. AIM 5.x clients call the ChatNav food group
// provided by BOS without creating an additional TCP connection.
type ChatNavServer struct {
	AuthService
	Handler
	Logger *slog.Logger
	OnlineNotifier
	config.Config
}

// Start starts a TCP server and listens for ChatNav connections.
func (rt ChatNavServer) Start() {
	addr := config.Address("", rt.Config.ChatNavPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind chat nav server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting chat nav service", "addr", addr)

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

func (rt ChatNavServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
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
		rwc.Close()
	}()

	ctx = context.WithValue(ctx, "screenName", bosSess.ScreenName())

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	// We copy the session object here to make sure that
	// dispatchIncomingMessages does not consume relayed messages produced by
	// the BOS server. Without this hack, message consumption would be split
	// between the BOS server and ChatNav server, which would result in
	// incorrect sequence number generation, because each server has its own
	// sequence counter. This hack can be removed by decoupling FLAP routing
	// and message relaying, which are both performed in
	// dispatchIncomingMessages.
	sessCopy := state.NewSession()
	sessCopy.SetScreenName(bosSess.ScreenName())
	sessCopy.SetID(bosSess.ID())

	return dispatchIncomingMessages(ctx, sessCopy, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}
