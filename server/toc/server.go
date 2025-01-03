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

func (c *ChatRegistry) Close() {
	c.m.Lock()
	defer c.m.Unlock()
	for _, sess := range c.sessions {
		sess.Close()
	}
}

type Server struct {
	BOSProxy   BOSProxy
	ChatProxy  ChatProxy
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
			if err := rt.handleNewConnection(conn, connCtx); err != nil {
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

func (rt Server) handleNewConnection(conn net.Conn, ctx context.Context) error {
	defer func() {
		conn.Close()
	}()

	bufCon := newBufferedConn(conn)
	b, err := bufCon.Peek(6)
	if err != nil {
		return fmt.Errorf("Peek: %w", err)
	}
	switch {
	case string(b) == "FLAPON":
		if err := rt.handleTOCOverFLAP(ctx, bufCon); err != nil {
			return fmt.Errorf("handleTOCOverFLAP: %w", err)
		}
	case strings.HasPrefix(string(b), "GET /"):
		if err := rt.handleTOCOverHTTP(bufCon, ctx, conn); err != nil {
			return fmt.Errorf("handleTOCOverHTTP: %w", err)
		}
	}

	return nil
}

type readWriter struct {
	http.Response
	w           io.Writer
	wroteHeader bool
}

func (r *readWriter) Header() http.Header {
	return r.Response.Header
}

func (r *readWriter) Write(i []byte) (int, error) {
	if !r.wroteHeader {
		r.Response.StatusCode = 200
		r.Response.ContentLength = int64(len(i))
		if err := r.Response.Write(r.w); err != nil {
			return 0, err
		}
	}
	return r.w.Write(i)
}

func (r *readWriter) WriteHeader(statusCode int) {
	r.Response.StatusCode = statusCode
	if err := r.Response.Write(r.w); err != nil {
		fmt.Println("error?")
		return
	}
	r.wroteHeader = true
}

func (rt Server) handleTOCOverHTTP(bufCon bufferedConn, thisCtx context.Context, conn net.Conn) error {
	bufReader := bufio.NewReader(bufCon)
	request, err := http.ReadRequest(bufReader)
	if err != nil {
		return errors.New("failed to read HTTP request: " + err.Error())
	}

	switch request.URL.Path {
	case "/info":
		rw := &readWriter{
			w: conn,
			Response: http.Response{
				ContentLength: -1, // disables content-length header, which works for hTTP 1.0
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header: http.Header{
					"Connection": []string{"close"},
				},
				Close: false,
			},
		}
		rt.BOSProxy.Profile(thisCtx, request, rw)
	case "/dir_info":
		rw := &readWriter{
			w: conn,
			Response: http.Response{
				ContentLength: -1, // disables content-length header, which works for hTTP 1.0
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header: http.Header{
					"Connection": []string{"close"},
				},
				Close: false,
			},
		}
		rt.BOSProxy.DirInfoHTTP(thisCtx, request, rw)
	case "/dir_search":
		rw := &readWriter{
			w: conn,
			Response: http.Response{
				ContentLength: -1, // disables content-length header, which works for hTTP 1.0
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header: http.Header{
					"Connection": []string{"close"},
				},
				Close: false,
			},
		}
		rt.BOSProxy.DirSearchHTTP(thisCtx, request, rw)
	}
	return nil
}

func (rt Server) handleTOCOverFLAP(ctx context.Context, clientConn io.ReadWriter) error {
	if err := rt.TOCHandshake(clientConn); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	clientFlap, err := rt.initFLAP(clientConn)
	if err != nil {
		return err
	}

	// buffered so that the go routine has room to exit
	msgCh := make(chan wire.FLAPFrame, 1)
	errCh := make(chan error, 1) // todo handle this

	go func() {
		defer func() {
			fmt.Println("closing handleTOCOverFLAP async function")
		}()
		defer close(msgCh)
		defer close(errCh)

		for {
			clientFrame, err := clientFlap.ReceiveFLAP()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				errCh <- fmt.Errorf("ReceiveFLAP: %w", err)
				return
			}

			if clientFrame.FrameType == wire.FLAPFrameSignoff {
				break // client disconnected
			}
			if clientFrame.FrameType == wire.FLAPFrameKeepAlive {
				continue // keep alive heartbeat
			}
			if clientFrame.FrameType != wire.FLAPFrameData {
				errCh <- fmt.Errorf("unexpected clientFlap clientFrame type: %d", clientFrame.FrameType)
				return
			}
			msgCh <- clientFrame
		}
	}()
	toClient := make(chan []byte, 2)

	var sessBOS *state.Session
	chatRegistry := &ChatRegistry{
		lookup:   make(map[int]string),
		sessions: make(map[int]*state.Session),
		m:        sync.RWMutex{},
	}

	select {
	case <-ctx.Done():
		return nil
	case clientFrame, ok := <-msgCh:
		if !ok {
			return nil
		}
		elems, err := receiveCmd(clientFrame.Payload)
		if err != nil {
			return fmt.Errorf("receive cmd failed: %w %s", err, clientFrame.Payload)
		}
		if len(elems) == 0 || elems[0] != "toc_signon" {
			return errors.New("expected toc_signon as first message")
		}

		var reply []string
		sessBOS, reply = rt.BOSProxy.Login(ctx, elems, chatRegistry, toClient)
		for _, m := range reply {
			if err := clientFlap.SendDataFrame([]byte(m)); err != nil {
				return fmt.Errorf("failed to send data frame %w", err)
			}
		}
		if sessBOS == nil {
			return nil
		}
		fmt.Printf("< client: %+v\n", elems)
	}

	defer func() {
		fmt.Println("closing handleTOCOverFLAP")
		sessBOS.Close()
		rt.BOSProxy.Signout(ctx, sessBOS)
		chatRegistry.Close()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-toClient:
				if !ok {
					fmt.Println("Closing client connections?")
					return
				}
				if err := clientFlap.SendDataFrame(msg); err != nil {
					// todo how to cancel everything?
					rt.Logger.Error("failed to send data frame", "err", err.Error())
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case clientFrame, ok := <-msgCh:
			if !ok {
				fmt.Println("Closing server connections?")
				return nil
			}
			elems, err := receiveCmd(clientFrame.Payload)
			if err != nil {
				return fmt.Errorf("receive cmd failed: %w %s", err, clientFrame.Payload)
			}

			if len(elems) == 0 {
				return errors.New("no cmd in flapon signal")
			}

			fmt.Printf("< client: %+v\n", elems)

			switch elems[0] {
			case "toc_send_im":
				rt.BOSProxy.SendIM(ctx, sessBOS, elems, toClient)
			case "toc_init_done":
				rt.BOSProxy.ClientReady(ctx, sessBOS, toClient)
			case "toc_add_buddy":
				rt.BOSProxy.AddBuddy(ctx, sessBOS, elems, toClient)
			case "toc_remove_buddy":
				rt.BOSProxy.RemoveBuddy(ctx, sessBOS, elems, toClient)
			case "toc_add_permit":
				rt.BOSProxy.AddPermit(ctx, sessBOS, elems, toClient)
			case "toc_add_deny":
				rt.BOSProxy.AddDeny(ctx, sessBOS, elems, toClient)
			case "toc_set_away":
				rt.BOSProxy.SetAway(ctx, sessBOS, elems[1], toClient)
			case "toc_set_caps":
				rt.BOSProxy.SetCaps(ctx, sessBOS, elems, toClient)
			case "toc_evil":
				rt.BOSProxy.Evil(ctx, sessBOS, elems, toClient)
			case "toc_get_info":
				rt.BOSProxy.GetInfoURL(sessBOS, elems, toClient)
			case "toc_chat_join":
				if !rt.ChatProxy.ChatJoin(ctx, sessBOS, chatRegistry, elems, toClient) {
					return nil
				}
			case "toc_chat_send":
				rt.ChatProxy.ChatSend(ctx, chatRegistry, elems, toClient)
			case "toc_chat_accept":
				if !rt.ChatProxy.ChatAccept(ctx, sessBOS, chatRegistry, elems, toClient) {
					return nil
				}
			case "toc_chat_leave":
				rt.ChatProxy.ChatLeave(ctx, chatRegistry, elems, toClient)
			case "toc_set_info":
				rt.BOSProxy.SetInfo(ctx, sessBOS, elems, toClient)
			case "toc_set_dir":
				rt.BOSProxy.SetDir(ctx, sessBOS, elems, toClient)
			case "toc_set_idle":
				rt.BOSProxy.SetIdle(ctx, sessBOS, elems, toClient)
			case "toc_set_config":
				rt.BOSProxy.SetConfig(ctx, sessBOS, elems, toClient)
			case "toc_chat_invite":
				rt.BOSProxy.ChatInvite(ctx, sessBOS, chatRegistry, elems, toClient)
			case "toc_dir_search":
				rt.BOSProxy.GetDirSearchURL(elems, toClient)
			case "toc_get_dir":
				rt.BOSProxy.GetDirURL(elems, toClient)
			default:
				rt.Logger.Error(fmt.Sprintf("unsupported TOC command %s", elems[0]))
			}
		}
	}

	return nil
}

func (rt Server) TOCHandshake(clientConn io.ReadWriter) error {
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
