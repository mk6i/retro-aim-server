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
	CookieCracker
	Handler
	ListenAddr string
	Logger     *slog.Logger
	OnlineNotifier
	config.Config
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt BOSServer) Start() {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		rt.Logger.Error("unable to bind BOS server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting service", "host", rt.ListenAddr)

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

func (rt BOSServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	flapc := wire.NewFlapClient(100, rwc, rwc)

	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	flap, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	authCookie, ok := flap.Slice(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	token, err := rt.CookieCracker.Crack(authCookie)
	if err != nil {
		return err
	}

	sess, err := rt.RegisterBOSSession(string(token))
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

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	return dispatchIncomingMessages(ctx, sess, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}
