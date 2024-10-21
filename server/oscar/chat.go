package oscar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

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
}

// Start creates a TCP server that implements that chat flow.
func (rt ChatServer) Start(ctx context.Context) error {
	addr := net.JoinHostPort("", rt.Config.ChatPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("unable to start chat sever: %w", err)
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

	wg.Wait()
	rt.Logger.Info("shutdown complete")
	return nil
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

	authCookie, ok := flap.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get login cookie from payload")
	}

	chatSess, err := rt.RegisterChatSession(authCookie)
	if err != nil {
		return err
	}
	if chatSess == nil {
		return errors.New("session not found")
	}

	defer func() {
		chatSess.Close()
		rwc.Close()
		rt.SignoutChat(ctx, chatSess)
	}()

	msg := rt.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "screenName", chatSess.IdentScreenName())
	return dispatchIncomingMessages(ctx, chatSess, flapc, rwc, rt.Logger, rt.Handler)
}
