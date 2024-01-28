package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
)

// BOSService provides client connection lifecycle management for the BOS
// service.
type BOSService struct {
	AuthHandler
	Logger *slog.Logger
	OServiceBOSHandler
	Router
	config.Config
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSService) Start() {
	addr := config.Address("", rt.Config.BOSPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind BOS server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting BOS service", "addr", addr)

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

func (rt BOSService) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	seq := uint32(100)

	flap, err := flapSignonHandshake(rwc, &seq)
	if err != nil {
		return err
	}

	var ok bool
	sessionID, ok := flap.Slice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	sess, err := rt.RetrieveBOSSession(string(sessionID))
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("session not found")
	}

	defer func() {
		sess.Close()
		rwc.Close()
		if err := rt.Signout(ctx, sess); err != nil {
			rt.Logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
		}
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName())

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		return err
	}

	return dispatchIncomingMessages(ctx, sess, seq, rwc, rt.Logger, rt.Router, sendSNAC, rt.Config)
}
