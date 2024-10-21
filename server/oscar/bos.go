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

// OnlineNotifier returns a OServiceHostOnline SNAC that is sent to the client
// at the beginning of the protocol sequence which lists all food groups
// managed by the server.
type OnlineNotifier interface {
	HostOnline() wire.SNACMessage
}

// BOSServer provides client connection lifecycle management for the BOS
// service.
type BOSServer struct {
	AuthService
	Handler
	ListenAddr string
	Logger     *slog.Logger
	OnlineNotifier
	config.Config
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		return fmt.Errorf("unable to start BOS server: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	rt.Logger.Info("starting server", "listen_host", rt.ListenAddr, "oscar_host", rt.Config.OSCARHost)

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

func (rt BOSServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
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
		return errors.New("unable to get session id from payload")
	}

	sess, err := rt.RegisterBOSSession(authCookie)
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

	ctx = context.WithValue(ctx, "screenName", sess.IdentScreenName())

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	return dispatchIncomingMessages(ctx, sess, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}
