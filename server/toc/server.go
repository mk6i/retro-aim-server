package toc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
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

type ChatRegistry struct {
	lookup   map[int]string
	sessions map[int]*state.Session
	nextID   int
	m        sync.RWMutex
}

func (c *ChatRegistry) Add(cookie string) int {
	c.m.Lock()
	defer c.m.Unlock()
	for k, v := range c.lookup {
		if v == cookie {
			return k
		}
	}
	id := c.nextID
	c.lookup[id] = cookie
	c.nextID++
	return id
}

func (c *ChatRegistry) Lookup(chatID int) string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.lookup[chatID]
}

func (c *ChatRegistry) Register(chatID int, sess *state.Session) {
	c.m.Lock()
	defer c.m.Unlock()
	c.sessions[chatID] = sess
}

func (c *ChatRegistry) Retrieve(chatID int) *state.Session {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.sessions[chatID]
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

	mux := http.NewServeMux()
	mux.Handle("GET /info", rt.BOSProxy.AuthMiddleware(http.HandlerFunc(rt.BOSProxy.Profile)))
	mux.Handle("GET /dir_info", rt.BOSProxy.AuthMiddleware(http.HandlerFunc(rt.BOSProxy.DirInfoHTTP)))
	mux.Handle("GET /dir_search", rt.BOSProxy.AuthMiddleware(http.HandlerFunc(rt.BOSProxy.DirSearchHTTP)))

	httpServer := &http.Server{
		Handler: mux,
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

	sessBOS, chatRegistry, err := rt.login(ctx, clientFlap)
	if err != nil {
		return fmt.Errorf("rt.login: %w", err)
	}
	if sessBOS == nil {
		return nil // user not found
	}

	defer rt.BOSProxy.Signout(ctx, sessBOS)

	// messages from TOC client
	fromCh := make(chan wire.FLAPFrame, 1)
	// messages to TOC client
	toCh := make(chan []byte, 2)

	// read in messages from client. when client disconnects, it closes fromCh.
	go rt.readFromClient(fromCh, clientFlap)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rt.BOSProxy.ConsumeIncomingBOS(gCtx, sessBOS, chatRegistry, toCh)
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
			elems, err := receiveCmd(clientFrame.Payload)
			if err != nil {
				return fmt.Errorf("receiveCmd: %w", err)
			}

			if len(elems) == 0 {
				return errors.New("no cmd in flapon signal")
			}

			fmt.Printf("< client: %+v\n", elems)

			switch elems[0] {
			case "toc_send_im":
				rt.BOSProxy.SendIM(ctx, sessBOS, elems, toCh)
			case "toc_init_done":
				rt.BOSProxy.BOSReady(ctx, sessBOS, toCh)
			case "toc_add_buddy":
				rt.BOSProxy.AddBuddy(ctx, sessBOS, elems, toCh)
			case "toc_remove_buddy":
				rt.BOSProxy.RemoveBuddy(ctx, sessBOS, elems, toCh)
			case "toc_add_permit":
				rt.BOSProxy.AddPermit(ctx, sessBOS, elems, toCh)
			case "toc_add_deny":
				rt.BOSProxy.AddDeny(ctx, sessBOS, elems, toCh)
			case "toc_set_away":
				rt.BOSProxy.SetAway(ctx, sessBOS, elems[1], toCh)
			case "toc_set_caps":
				rt.BOSProxy.SetCaps(ctx, sessBOS, elems, toCh)
			case "toc_evil":
				rt.BOSProxy.Evil(ctx, sessBOS, elems, toCh)
			case "toc_get_info":
				rt.BOSProxy.GetInfoURL(ctx, sessBOS, elems, toCh)
			case "toc_chat_join", "toc_chat_accept":
				var chatID int
				var joinOK bool
				if elems[0] == "toc_chat_join" {
					chatID, joinOK = rt.BOSProxy.ChatJoin(ctx, sessBOS, chatRegistry, elems, toCh)
				} else {
					chatID, joinOK = rt.BOSProxy.ChatAccept(ctx, sessBOS, chatRegistry, elems, toCh)
				}
				if joinOK {
					doAsync(func() error {
						sess := chatRegistry.Retrieve(chatID)
						rt.BOSProxy.ConsumeIncomingChat(ctx, sess, chatID, toCh)
						return nil
					})
				}
			case "toc_chat_send":
				rt.BOSProxy.ChatSend(ctx, chatRegistry, elems, toCh)
			case "toc_chat_leave":
				rt.BOSProxy.ChatLeave(ctx, chatRegistry, elems, toCh)
			case "toc_set_info":
				rt.BOSProxy.SetInfo(ctx, sessBOS, elems, toCh)
			case "toc_set_dir":
				rt.BOSProxy.SetDir(ctx, sessBOS, elems, toCh)
			case "toc_set_idle":
				rt.BOSProxy.SetIdle(ctx, sessBOS, elems, toCh)
			case "toc_set_config":
				rt.BOSProxy.SetConfig(ctx, sessBOS, elems, toCh)
			case "toc_chat_invite":
				rt.BOSProxy.ChatInvite(ctx, sessBOS, chatRegistry, elems, toCh)
			case "toc_dir_search":
				rt.BOSProxy.GetDirSearchURL(ctx, sessBOS, elems, toCh)
			case "toc_get_dir":
				rt.BOSProxy.GetDirURL(ctx, sessBOS, elems, toCh)
			default:
				rt.Logger.Error(fmt.Sprintf("unsupported TOC command %s", elems[0]))
			}
		}
	}
	return nil
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
		}
	}
}

func (rt Server) login(ctx context.Context, clientFlap *wire.FlapClient) (*state.Session, *ChatRegistry, error) {
	clientFrame, err := clientFlap.ReceiveFLAP()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("clientFlap.ReceiveFLAP: %w", err)
	}

	elems, err := receiveCmd(clientFrame.Payload)
	if err != nil {
		return nil, nil, fmt.Errorf("receiveCmd: %w", err)
	}
	if len(elems) == 0 || elems[0] != "toc_signon" {
		return nil, nil, errors.New("expected toc_signon as first message")
	}

	chatRegistry := &ChatRegistry{
		lookup:   make(map[int]string),
		sessions: make(map[int]*state.Session),
		m:        sync.RWMutex{},
	}

	sessBOS, reply := rt.BOSProxy.Login(ctx, elems)
	for _, m := range reply {
		if err := clientFlap.SendDataFrame([]byte(m)); err != nil {
			return nil, nil, fmt.Errorf("clientFlap.SendDataFrame: %w", err)
		}
	}

	fmt.Printf("< client: %+v\n", elems)
	return sessBOS, chatRegistry, nil
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

func receiveCmd(b []byte) ([]string, error) {
	if b[len(b)-1] == '\x00' {
		b = b[:len(b)-1]
	}
	if bytes.HasPrefix(b, []byte("toc_set_config")) {
		// gaim uses braces instead of quotes for some reason
		first := bytes.IndexByte(b, '{')
		if first != -1 {
			b[first] = '"'
		}
		last := bytes.LastIndexByte(b, '}')
		if last != -1 {
			b[last] = '"'
		}
	}
	reader := csv.NewReader(bytes.NewReader(b))
	reader.Comma = ' '
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return reader.Read()
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
