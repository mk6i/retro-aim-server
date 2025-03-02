package toc

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// bufferedConn is a wrapper around net.Conn that allows peeking into the
// incoming connection without consuming data. It is useful for multiplexing
// TOC/HTTP and TOC/FLAP connections.
//
// It embeds net.Conn, so all standard connection methods remain available.
type bufferedConn struct {
	r *bufio.Reader
	net.Conn
}

// newBufferedConn wraps a net.Conn with buffered reading capabilities.
func newBufferedConn(c net.Conn) bufferedConn {
	return bufferedConn{bufio.NewReader(c), c}
}

// Peek returns the next n bytes from the buffer without advancing the reader.
// If fewer than n bytes are available, it returns an error.
func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

// Read reads data into p from the buffered connection.
// It prioritizes buffered data before reading from the underlying connection.
func (b bufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// channelListener is an implementation of net.Listener that accepts connections
// from a channel instead of a network socket. It is useful for attaching an
// HTTP service to a connection on the fly.
type channelListener struct {
	ch chan net.Conn // Channel used to receive connections.
}

// Accept waits for and returns the next connection from the channel.
// If the channel is closed, it returns io.EOF to indicate no more connections.
func (l *channelListener) Accept() (net.Conn, error) {
	ch, ok := <-l.ch
	if !ok {
		return nil, io.EOF
	}
	return ch, nil
}

// Close closes the listener. Since channelListener does not manage an actual
// network connection, this is a no-op and always returns nil.
func (l *channelListener) Close() error {
	return nil
}

// Addr returns the network address of the listener.
// Since channelListener is not bound to a real network address, it returns nil.
func (l *channelListener) Addr() net.Addr {
	return nil
}

// Server implements a TOC protocol server that multiplexes TOC/HTTP and
// TOC/FLAP requests. It acts as a gateway, forwarding all TOC requests
// to the OSCAR server for processing.
type Server struct {
	BOSProxy   OSCARProxy
	ListenAddr string
	Logger     *slog.Logger
}

func (rt Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		return fmt.Errorf("unable to start TOC server: %w", err)
	}

	rt.Logger.InfoContext(ctx, "starting server", "listen_host", rt.ListenAddr)

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	httpServer := &http.Server{
		Handler: rt.BOSProxy.NewServeMux(),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	httpCh := make(chan net.Conn)
	defer close(httpCh)

	go func() {
		_ = httpServer.Serve(&channelListener{ch: httpCh})
	}()

	wg := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			rt.Logger.ErrorContext(ctx, "accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			rt.dispatchConn(conn, ctx, httpCh)
		}()
	}

	if !waitForShutdown(&wg) {
		rt.Logger.ErrorContext(ctx, "shutdown complete, but connections didn't close cleanly")
	} else {
		rt.Logger.InfoContext(ctx, "shutdown complete")
	}

	return nil
}

// dispatchConn inspects and routes an incoming connection. If the connection
// starts with "FLAP", handle as TOC/FLAP; otherwise, dispatch for HTTP
// processing.
func (rt Server) dispatchConn(conn net.Conn, ctx context.Context, httpCh chan net.Conn) error {
	bufCon := newBufferedConn(conn)

	doFlap := "FLAP"
	buf, err := bufCon.Peek(len(doFlap))
	if err != nil {
		return fmt.Errorf("bufCon.Peek: %w", err)
	}

	if string(buf) == doFlap {
		if err = rt.dispatchFLAP(ctx, bufCon); err != nil {
			return fmt.Errorf("dispatchFLAP: %w", err)
		}
		return nil
	}

	select {
	case httpCh <- bufCon:
		return nil
	case <-ctx.Done():
		return nil
	}
}

func (rt Server) dispatchFLAP(ctx context.Context, conn net.Conn) error {
	defer func() {
		_ = conn.Close()
	}()
	ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())

	clientFlap, err := rt.initFLAP(conn)
	if err != nil {
		return err
	}

	sessBOS, err := rt.login(ctx, clientFlap)
	if err != nil {
		return fmt.Errorf("rt.login: %w", err)
	}
	if sessBOS == nil {
		return nil // user not found
	}

	ctx = context.WithValue(ctx, "screenName", sessBOS.IdentScreenName())

	remoteAddr, ok := ctx.Value("ip").(string)
	if ok {
		ip, err := netip.ParseAddrPort(remoteAddr)
		if err != nil {
			return errors.New("unable to parse ip addr")
		}
		sessBOS.SetRemoteAddr(&ip)
	}

	chatRegistry := NewChatRegistry()

	defer rt.BOSProxy.Signout(ctx, sessBOS, chatRegistry)

	// messages from TOC client
	fromCh := make(chan wire.FLAPFrame, 1)
	// messages to TOC client
	toCh := make(chan []byte, 2)

	// read in messages from client. when client disconnects, it closes fromCh and toCh.
	go rt.readFromClient(ctx, fromCh, toCh, clientFlap)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rt.BOSProxy.RecvBOS(gCtx, sessBOS, chatRegistry, toCh)
	})
	g.Go(func() error {
		return rt.sendToClient(gCtx, toCh, clientFlap)
	})
	g.Go(func() error {
		return rt.processCommands(gCtx, g.Go, sessBOS, chatRegistry, fromCh, toCh)
	})

	err = g.Wait()
	if errors.Is(err, errDisconnect) {
		err = nil
	}
	return err
}

func (rt Server) processCommands(
	ctx context.Context,
	doAsync func(f func() error),
	sessBOS *state.Session,
	chatRegistry *ChatRegistry,
	fromCh <-chan wire.FLAPFrame,
	toCh chan<- []byte,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case clientFrame, ok := <-fromCh:
			if !ok {
				return nil
			}
			clientFrame.Payload = bytes.TrimRight(clientFrame.Payload, "\x00") // trim null terminator

			if len(clientFrame.Payload) == 0 {
				return errors.New("TOC command is empty")
			}
			if len(clientFrame.Payload) > 2048 {
				return errors.New("TOC command exceeds maximum length (2048)")
			}

			msg, ok := rt.BOSProxy.RecvClientCmd(ctx, sessBOS, chatRegistry, clientFrame.Payload, toCh, doAsync)
			if !ok {
				return nil
			}
			if len(msg) > 0 {
				select {
				case toCh <- []byte(msg):
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func (rt Server) sendToClient(ctx context.Context, toClient <-chan []byte, clientFlap *wire.FlapClient) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-toClient:
			if err := clientFlap.SendDataFrame(msg); err != nil {
				return fmt.Errorf("clientFlap.SendDataFrame: %w", err)
			}
			if rt.Logger.Enabled(ctx, slog.LevelDebug) {
				rt.Logger.DebugContext(ctx, "server response", "command", msg)
			} else {
				// just log the command, omit params
				idx := len(msg)
				if col := bytes.IndexByte(msg, ':'); col > -1 {
					idx = col
				}
				rt.Logger.InfoContext(ctx, "server response", "command", msg[0:idx])
			}
		}
	}
}

func (rt Server) login(ctx context.Context, clientFlap *wire.FlapClient) (*state.Session, error) {
	clientFrame, err := clientFlap.ReceiveFLAP()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("clientFlap.ReceiveFLAP: %w", err)
	}

	sessBOS, reply := rt.BOSProxy.Signon(ctx, clientFrame.Payload)
	for _, m := range reply {
		if err := clientFlap.SendDataFrame([]byte(m)); err != nil {
			return nil, fmt.Errorf("clientFlap.SendDataFrame: %w", err)
		}
	}

	return sessBOS, nil
}

func (rt Server) readFromClient(ctx context.Context, msgCh chan<- wire.FLAPFrame, toCh chan<- []byte, clientFlap *wire.FlapClient) {
	defer close(msgCh)
	defer close(toCh)

	for {
		clientFrame, err := clientFlap.ReceiveFLAP()
		if err != nil {
			if !(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				rt.Logger.ErrorContext(ctx, "ReceiveFLAP error", "err", err.Error())
			}
			break
		}

		if clientFrame.FrameType == wire.FLAPFrameSignoff {
			break // client disconnected
		}
		if clientFrame.FrameType == wire.FLAPFrameKeepAlive {
			continue // keep alive heartbeat
		}
		if clientFrame.FrameType != wire.FLAPFrameData {
			rt.Logger.ErrorContext(ctx, "unexpected clientFlap clientFrame type", "type", clientFrame.FrameType)
			break
		}
		msgCh <- clientFrame
	}
}

// initFLAP sets up a new FLAP connection. It returns a flap client if the
// connection successfully initialized.
func (rt Server) initFLAP(rw io.ReadWriter) (*wire.FlapClient, error) {
	expected := "FLAPON\r\n\r\n"
	buf := make([]byte, len(expected))

	_, err := io.ReadFull(rw, buf)
	if err != nil {
		return nil, fmt.Errorf("io.ReadFull: %w", err)
	}
	if expected != string(buf) {
		return nil, fmt.Errorf("expected FLAPON, got %s", buf)
	}

	clientFlap := wire.NewFlapClient(0, rw, rw)

	if err := clientFlap.SendSignonFrame(nil); err != nil {
		return nil, fmt.Errorf("clientFlap.SendSignonFrame: %w", err)
	}
	if _, err := clientFlap.ReceiveSignonFrame(); err != nil {
		return nil, fmt.Errorf("clientFlap.ReceiveSignonFrame: %w", err)
	}

	return clientFlap, nil
}

// waitForShutdown returns when either the wg completes or 5 seconds has
// passed. This is a temporary hack to ensure that the server shuts down even
// if all the TCP connections do not drain. Return true if the shutdown is
// clean.
func waitForShutdown(wg *sync.WaitGroup) bool {
	ch := make(chan struct{})

	go func() {
		wg.Wait() // goroutine leak if wg never completes
		close(ch)
	}()

	select {
	case <-ch:
		return true
	case <-time.After(time.Second * 5):
		return false
	}
}
