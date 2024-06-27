package oscar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// AdminServer provides client connection lifecycle management for the BOS
// service.
type AdminServer struct {
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
func (rt AdminServer) Start() {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		rt.Logger.Error("unable to bind server address", "host", rt.ListenAddr, "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting server", "listen_host", rt.ListenAddr, "oscar_host", rt.Config.OSCARHost)

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

func (rt AdminServer) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
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

	sess, err := rt.RetrieveBOSSession(authCookie)
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

	return dispatchIncomingMessagesSimple(ctx, sess, flapc, rwc, rt.Logger, rt.Handler, rt.Config)
}

func dispatchIncomingMessagesSimple(ctx context.Context, sess *state.Session, flapc *wire.FlapClient, r io.Reader, logger *slog.Logger, router Handler, config config.Config) error {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(r, msgCh, readErrCh)

	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	for {
		select {
		case m, ok := <-msgCh:
			if !ok {
				return nil
			}
			switch m.flap.FrameType {
			case wire.FLAPFrameData:
				inFrame := wire.SNACFrame{}
				if err := wire.Unmarshal(&inFrame, m.payload); err != nil {
					return err
				}
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Handle(ctx, sess, inFrame, m.payload, flapc); err != nil {
					middleware.LogRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrRouteNotFound) {
						if err1 := sendInvalidSNACErr(inFrame, flapc); err1 != nil {
							return errors.Join(err1, err)
						}
						if config.FailFast {
							panic(err.Error())
						}
						break
					}
					return err
				}
			case wire.FLAPFrameSignon:
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", m.flap)
			case wire.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", m.flap)
			case wire.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", m.flap)
				return nil
			case wire.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", m.flap)
			}
		case err := <-readErrCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return nil
		}
	}
}
