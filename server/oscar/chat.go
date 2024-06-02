package oscar

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/wire"
)

// ChatServer represents a service that implements a chat room session.
// Clients connect to this service upon creating a chat room or being invited
// to a chat room.
type ChatServer struct {
	AuthService
	Handler
	Logger *slog.Logger
	OnlineNotifier
	config.Config
	CookieCracker
}

// Start creates a TCP server that implements that chat flow.
func (rt ChatServer) Start() {
	addr := net.JoinHostPort("", rt.Config.ChatPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind chat server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting chat service", "host", net.JoinHostPort(rt.Config.OSCARHost, rt.Config.ChatPort))

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

func (rt ChatServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	flapc := wire.NewFlapClient(100, rwc, rwc)
	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	flap, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	var ok bool
	loginCookie, ok := flap.Slice(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get login cookie from payload")
	}

	token, err := rt.CookieCracker.Crack(loginCookie)
	if err != nil {
		return err
	}

	chatSess, err := rt.RegisterChatSession(token)
	if err != nil {
		return err
	}
	if chatSess == nil {
		return errors.New("session not found")
	}

	defer func() {
		chatSess.Close()
		rwc.Close()
		if err := rt.SignoutChat(ctx, chatSess); err != nil {
			rt.Logger.ErrorContext(ctx, "unable to sign out user", "err", err.Error())
		}
	}()

	msg := rt.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())
	return dispatchIncomingMessages(ctx, chatSess, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}
