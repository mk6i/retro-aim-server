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
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type bufferedConn struct {
	r        *bufio.Reader
	net.Conn // So that most methods are embedded
}

func newBufferedConn(c net.Conn) bufferedConn {
	return bufferedConn{bufio.NewReader(c), c}
}

func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b bufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

type channelListener struct {
	ch chan net.Conn
}

func (l *channelListener) Accept() (net.Conn, error) {
	ch, ok := <-l.ch
	if !ok {
		return nil, io.EOF
	}
	return ch, nil
}

func (l *channelListener) Close() error {
	return nil
}

func (l *channelListener) Addr() net.Addr {
	return nil
}

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

	rt.Logger.Info("starting server", "listen_host", rt.ListenAddr)

	go func() {
		<-ctx.Done()
		listener.Close()
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
			rt.Logger.Error("accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			bufCon := newBufferedConn(conn)
			b, err := bufCon.Peek(6)
			if err != nil {
				rt.Logger.Error("peek failed", "err", err.Error())
				return
			}
			switch {
			case string(b) == "FLAPON":
				ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
				if err := rt.handleTOCOverFLAP(ctx, bufCon); err != nil {
					rt.Logger.Error("handleTOCOverFLAP failed", "err", err.Error())
					return
				}
			case strings.HasPrefix(string(b), "GET /"):
				select {
				case httpCh <- bufCon:
					fmt.Println("Sent off connection")
				case <-ctx.Done():
					return
				}
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

func (rt Server) handleTOCOverFLAP(ctx context.Context, conn io.ReadWriteCloser) error {
	defer func() {
		conn.Close()
	}()

	if err := rt.handshake(conn); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

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

	defer rt.BOSProxy.Signout(ctx, sessBOS)

	// messages from TOC client
	fromCh := make(chan wire.FLAPFrame, 1)
	// messages to TOC client
	toCh := make(chan []byte, 2)

	// read in messages from client. when client disconnects, it closes fromCh.
	go rt.readFromClient(fromCh, clientFlap)

	g, gCtx := errgroup.WithContext(ctx)

	chatRegistry := newChatRegistry()

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
	defer func() {
		fmt.Println("closing processCommands")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case clientFrame, ok := <-fromCh:
			if !ok {
				return errDisconnect
			}
			clientFrame.Payload = bytes.TrimRight(clientFrame.Payload, "\x00") // trim null terminator

			if len(clientFrame.Payload) == 0 {
				return errors.New("no givenPayload in flapon signal")
			}

			cmd := clientFrame.Payload
			if idx := bytes.IndexByte(clientFrame.Payload, ' '); idx > -1 {
				cmd = clientFrame.Payload[:idx]
			}

			rt.logClientCommand(ctx, clientFrame, cmd)

			reply := func(msg string) {
				if len(msg) == 0 {
					return
				}
				select {
				case toCh <- []byte(msg):
				case <-ctx.Done():
					return
				}
				// todo disconnect when internal svc err
			}

			switch string(cmd) {
			case "toc_send_im":
				reply(rt.BOSProxy.SendIM(ctx, sessBOS, clientFrame.Payload))
			case "toc_init_done":
				reply(rt.BOSProxy.InitDone(ctx, sessBOS, clientFrame.Payload))
			case "toc_add_buddy":
				reply(rt.BOSProxy.AddBuddy(ctx, sessBOS, clientFrame.Payload))
			case "toc_remove_buddy":
				reply(rt.BOSProxy.RemoveBuddy(ctx, sessBOS, clientFrame.Payload))
			case "toc_add_permit":
				reply(rt.BOSProxy.AddPermit(ctx, sessBOS, clientFrame.Payload))
			case "toc_add_deny":
				reply(rt.BOSProxy.AddDeny(ctx, sessBOS, clientFrame.Payload))
			case "toc_set_away":
				reply(rt.BOSProxy.SetAway(ctx, sessBOS, clientFrame.Payload))
			case "toc_set_caps":
				reply(rt.BOSProxy.SetCaps(ctx, sessBOS, clientFrame.Payload))
			case "toc_evil":
				reply(rt.BOSProxy.Evil(ctx, sessBOS, clientFrame.Payload))
			case "toc_get_info":
				reply(rt.BOSProxy.GetInfoURL(ctx, sessBOS, clientFrame.Payload))
			case "toc_chat_join", "toc_chat_accept":
				var chatID int
				var msg string

				if string(cmd) == "toc_chat_join" {
					chatID, msg = rt.BOSProxy.ChatJoin(ctx, sessBOS, chatRegistry, clientFrame.Payload)
				} else {
					chatID, msg = rt.BOSProxy.ChatAccept(ctx, sessBOS, chatRegistry, clientFrame.Payload)
				}
				reply(msg)

				if msg == cmdInternalSvcErr {
					return nil
				}

				doAsync(func() error {
					sess := chatRegistry.RetrieveSess(chatID)
					rt.BOSProxy.RecvChat(ctx, sess, chatID, toCh)
					return nil
				})
			case "toc_chat_send":
				reply(rt.BOSProxy.ChatSend(ctx, chatRegistry, clientFrame.Payload))
			case "toc_chat_leave":
				reply(rt.BOSProxy.ChatLeave(ctx, chatRegistry, clientFrame.Payload))
			case "toc_set_info":
				reply(rt.BOSProxy.SetInfo(ctx, sessBOS, clientFrame.Payload))
			case "toc_set_dir":
				reply(rt.BOSProxy.SetDir(ctx, sessBOS, clientFrame.Payload))
			case "toc_set_idle":
				reply(rt.BOSProxy.SetIdle(ctx, sessBOS, clientFrame.Payload))
			case "toc_set_config":
				reply(rt.BOSProxy.SetConfig(ctx, sessBOS, clientFrame.Payload))
			case "toc_chat_invite":
				reply(rt.BOSProxy.ChatInvite(ctx, sessBOS, chatRegistry, clientFrame.Payload))
			case "toc_dir_search":
				reply(rt.BOSProxy.GetDirSearchURL(ctx, sessBOS, clientFrame.Payload))
			case "toc_get_dir":
				reply(rt.BOSProxy.GetDirURL(ctx, sessBOS, clientFrame.Payload))
			default:
				rt.Logger.Error(fmt.Sprintf("unsupported TOC command %s", cmd))
			}
		}
	}
	return nil
}

func (rt Server) logClientCommand(ctx context.Context, clientFrame wire.FLAPFrame, cmd []byte) {
	if rt.Logger.Enabled(ctx, slog.LevelDebug) {
		rt.Logger.InfoContext(ctx, "client request", "command", clientFrame.Payload)
	} else {
		rt.Logger.InfoContext(ctx, "client request", "command", cmd)
	}
}

func (rt Server) sendToClient(ctx context.Context, toClient <-chan []byte, clientFlap *wire.FlapClient) error {
	defer func() {
		fmt.Println("closing sendToClient")
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-toClient:
			if err := clientFlap.SendDataFrame(msg); err != nil {
				return fmt.Errorf("clientFlap.SendDataFrame: %w", err)
			}
			rt.logServerResponse(ctx, msg)
		}
	}
}

func (rt Server) logServerResponse(ctx context.Context, msg []byte) {
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

	fmt.Printf("< client: %+v\n", clientFrame.Payload)
	return sessBOS, nil
}

func (rt Server) readFromClient(msgCh chan<- wire.FLAPFrame, clientFlap *wire.FlapClient) {
	defer func() {
		fmt.Println("closing readFromClient")
	}()
	defer close(msgCh)

	for {
		clientFrame, err := clientFlap.ReceiveFLAP()
		if err != nil {
			if !(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				rt.Logger.Error("ReceiveFLAP error", "err", err.Error())
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
			rt.Logger.Error("unexpected clientFlap clientFrame type", "type", clientFrame.FrameType)
			break
		}
		msgCh <- clientFrame
	}
}

func (rt Server) handshake(clientConn io.ReadWriter) error {
	reader := bufio.NewReader(clientConn)

	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read line failed: %w", err)
	}
	if string(line) != "FLAPON" {
		return fmt.Errorf("unexpected line: %s", string(line))
	}
	line, _, err = reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read line failed: %w", err)
	}
	return nil
}

func (rt Server) initFLAP(clientConn io.ReadWriter) (*wire.FlapClient, error) {
	clientFlap := wire.NewFlapClient(0, clientConn, clientConn)

	fmt.Printf("sending signon frame\n")
	if err := clientFlap.SendSignonFrame(nil); err != nil {
		return nil, fmt.Errorf("send flapon signal failed: %w", err)
	}

	signonFrame, err := clientFlap.ReceiveSignonFrame()
	if err != nil {
		return nil, fmt.Errorf("send flapon signal failed: %w", err)
	}

	fmt.Printf("received signon frame: %v\n", signonFrame)
	return clientFlap, nil
}

func (rt Server) respond(s string, rwc io.ReadWriteCloser) error {
	fmt.Printf("server: %s\n", s)
	if _, err := io.WriteString(rwc, s); err != nil {
		return fmt.Errorf("error writing FLAPON: %w", err)
	}
	return nil
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
