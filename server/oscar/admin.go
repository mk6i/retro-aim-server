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
func (rt AdminServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		return fmt.Errorf("unable to start admin server: %w", err)
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

	if !waitForShutdown(&wg) {
		rt.Logger.Error("shutdown complete, but connections didn't close cleanly")
	} else {
		rt.Logger.Info("shutdown complete")
	}

	return nil
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

	authCookie, ok := flap.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	sess, err := rt.RetrieveBOSSession(authCookie)
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("session not found")
	}

	defer func() {
		rwc.Close()
	}()

	ctx = context.WithValue(ctx, "screenName", sess.IdentScreenName())

	msg := rt.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	return dispatchIncomingMessagesSimple(ctx, sess, flapc, rwc, rt.Logger, rt.Handler)
}

func dispatchIncomingMessagesSimple(ctx context.Context, sess *state.Session, flapc *wire.FlapClient, r io.Reader, logger *slog.Logger, router Handler) error {
	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	// buffered so that the go routine has room to exit
	msgCh := make(chan wire.FLAPFrame, 1)
	errCh := make(chan error, 1)

	// consume flap frames
	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			frame := wire.FLAPFrame{}
			if err := wire.UnmarshalBE(&frame, r); err != nil {
				errCh <- err
				return
			}
			msgCh <- frame
		}
	}()

	for {
		select {
		case flap, ok := <-msgCh:
			if !ok {
				return nil
			}
			switch flap.FrameType {
			case wire.FLAPFrameData:
				flapBuf := bytes.NewBuffer(flap.Payload)

				inFrame := wire.SNACFrame{}
				if err := wire.UnmarshalBE(&inFrame, flapBuf); err != nil {
					return err
				}
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Handle(ctx, sess, inFrame, flapBuf, flapc); err != nil {
					middleware.LogRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrRouteNotFound) {
						if err1 := sendInvalidSNACErr(inFrame, flapc); err1 != nil {
							return errors.Join(err1, err)
						}
						break
					}
					return err
				}
			case wire.FLAPFrameSignon:
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", flap)
			case wire.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", flap)
			case wire.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", flap)
				return nil
			case wire.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", flap)
			}
		case <-ctx.Done():
			// application is shutting down
			if err := flapc.Disconnect(); err != nil {
				return fmt.Errorf("unable to gracefully disconnect user. %w", err)
			}
			return nil
		case err := <-errCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return nil
		}
	}
}
