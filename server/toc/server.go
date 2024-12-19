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

func newBufferedConnSize(c net.Conn, n int) bufferedConn {
	return bufferedConn{bufio.NewReaderSize(c, n), c}
}

func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b bufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// Server provides client connection lifecycle management for the BOS
// service.
type Server struct {
	ListenAddr    string
	Logger        *slog.Logger
	LocateService LocateService
	BOSProxy      BOSProxy
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
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
		go func() {
			<-ctx.Done()
			conn.Close()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			connCtx := context.WithValue(ctx, "ip", conn.RemoteAddr().String())

			defer conn.Close()

			bufCon := newBufferedConn(conn)
			b, err := bufCon.Peek(6)
			if err != nil {
				rt.Logger.Error("peek failed", "err", err.Error())
			}
			switch {
			case string(b) == "FLAPON":
				if err := rt.handleTOCOverFlap(connCtx, bufCon); err != nil {
					rt.Logger.Info("user session failed", "err", err.Error())
				}
			case strings.HasPrefix(string(b), "GET /"):
				bufReader := bufio.NewReader(bufCon)
				request, err := http.ReadRequest(bufReader)
				if err != nil {
					fmt.Println("Error reading HTTP request:", err)
					return
				}

				switch request.URL.Path {
				case "/info":
					from := request.URL.Query().Get("from")
					if from == "" {
						rt.Logger.Error("no from query parameter")
					}
					user := request.URL.Query().Get("user")
					if user == "" {
						rt.Logger.Error("no user query parameter")
					}

					sess := state.NewSession()
					sess.SetIdentScreenName(state.NewIdentScreenName(from))
					inBody := wire.SNAC_0x02_0x05_LocateUserInfoQuery{
						Type:       uint16(wire.LocateTypeSig),
						ScreenName: user,
					}

					info, err := rt.LocateService.UserInfoQuery(ctx, sess, wire.SNACFrame{}, inBody)
					if err != nil {
						rt.Logger.Error("user session failed", "err", err.Error())
						return
					}
					if !(info.Frame.FoodGroup == wire.Locate && info.Frame.SubGroup == wire.LocateUserInfoReply) {
						rt.Logger.Error("didn't get expected locate response")
						return
					}

					locateInfoReply := info.Body.(wire.SNAC_0x02_0x06_LocateUserInfoReply)
					profile, hasProf := locateInfoReply.LocateInfo.Bytes(wire.LocateTLVTagsInfoSigData)
					if !hasProf {
						rt.Logger.Error("didn't get expected location info")
						return
					}
					response := http.Response{
						Status:        http.StatusText(http.StatusOK),
						StatusCode:    http.StatusOK,
						Proto:         "HTTP/1.0",
						ProtoMajor:    1,
						ProtoMinor:    0,
						Header:        make(http.Header),
						Body:          nil,
						ContentLength: int64(len(profile)),
						Close:         true,
					}

					response.Header.Set("Content-Type", "text/plain")
					response.Header.Set("Content-Length", fmt.Sprintf("%d", len(profile)))

					if err := response.Write(conn); err != nil {
						fmt.Println("Error writing response:", err)
						return
					}

					if _, err = conn.Write([]byte(profile)); err != nil {
						fmt.Println("Error writing myProfile:", err)
					}
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

func (rt Server) handleTOCOverFlap(ctx context.Context, clientConn io.ReadWriter) error {
	if err := rt.TOCHandshake(clientConn); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	clientFlap, err := rt.initFLAP(clientConn)
	if err != nil {
		return err
	}

	clientCh := make(chan []byte)

	defer func() {
		close(clientCh)
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				rt.Logger.Info("closing client writer")
				return
			case msg := <-clientCh:
				if err := clientFlap.SendDataFrame(msg); err != nil {
					rt.Logger.Error("failed to send data frame", "err", err.Error())
				}
			}
		}
	}()

	var sessBOS *state.Session
	var sessChat *state.Session

	for {
		clientFrame, err := clientFlap.ReceiveFLAP()
		if err != nil {
			return fmt.Errorf("send flapon signal failed: %w", err)
		}

		if clientFrame.FrameType == wire.FLAPFrameSignoff {
			break // client disconnected
		}
		if clientFrame.FrameType == wire.FLAPFrameKeepAlive {
			continue // keep alive heartbeat
		}
		if clientFrame.FrameType != wire.FLAPFrameData {
			return fmt.Errorf("unexpected clientFlap clientFrame type: %s", clientFrame.FrameType)
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
		case "toc_signon":
			sessBOS, err = rt.BOSProxy.Login(elems)
			if err != nil {
				return fmt.Errorf("init BOS failed: %w", err)
			}

			clientCh <- []byte("SIGN_ON:1")

			go rt.BOSProxy.ConsumeIncoming(ctx, sessBOS, clientCh)

		case "toc_send_im":
			if rt.BOSProxy.SendIM(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("send IM failed: %w", err)
			}
		case "toc_init_done":
			if err := rt.BOSProxy.ClientReady(ctx, sessBOS); err != nil {
				return fmt.Errorf("client ready notification failed: %w", err)
			}
		case "toc_add_buddy":
			if err := rt.BOSProxy.AddBuddy(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("add buddy failed: %w", err)
			}
		case "toc_remove_buddy":
			if err := rt.BOSProxy.RemoveBuddy(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("add buddy failed: %w", err)
			}
		case "toc_add_permit":
			if err := rt.BOSProxy.AddPermit(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("add buddy failed: %w", err)
			}
		case "toc_add_deny":
			if err := rt.BOSProxy.AddDeny(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("add buddy failed: %w", err)
			}
		case "toc_set_away":
			if err := rt.BOSProxy.SetAway(ctx, sessBOS, elems[1]); err != nil {
				return fmt.Errorf("set away failed: %w", err)
			}
		case "toc_set_caps":
			if err := rt.BOSProxy.SetCaps(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("set caps failed: %w", err)
			}
		case "toc_evil":
			response, err := rt.BOSProxy.Evil(ctx, sessBOS, elems)
			if err != nil {
				return fmt.Errorf("evil failed: %w", err)
			}
			clientCh <- []byte(response)
		case "toc_get_info":
			if err := clientFlap.SendDataFrame([]byte(fmt.Sprintf("GOTO_URL:profile:info?from=%s&user=%s", "mike", elems[1]))); err != nil {
				return fmt.Errorf("send toc_get_info failed: %w", err)
			}
		//case "toc_chat_join":
		//	exchange, err := strconv.Atoi(elems[1])
		//	if err != nil {
		//		return fmt.Errorf("parse exchange failed: %w", err)
		//	}
		//	bosCh <- wire.SNACMessage{
		//		Frame: wire.SNACFrame{
		//			FoodGroup: wire.OService,
		//			SubGroup:  wire.OServiceServiceRequest,
		//		},
		//		Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
		//			FoodGroup: wire.ChatNav,
		//		},
		//	}
		//	chatNavCh <- wire.SNACMessage{
		//		Frame: wire.SNACFrame{
		//			FoodGroup: wire.ChatNav,
		//			SubGroup:  wire.ChatNavCreateRoom,
		//		},
		//		Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		//			Exchange: uint16(exchange),
		//			Cookie:   "create",
		//			TLVBlock: wire.TLVBlock{
		//				TLVList: wire.TLVList{
		//					wire.NewTLVBE(wire.ChatRoomTLVRoomName, elems[2]),
		//				},
		//			},
		//		},
		//	}
		case "toc_chat_send":
			if err := rt.BOSProxy.SetCaps(ctx, sessChat, elems); err != nil {
				return fmt.Errorf("set caps failed: %w", err)
			}
		//case "toc_chat_accept":
		//	bosCh <- wire.SNACMessage{
		//		Frame: wire.SNACFrame{
		//			FoodGroup: wire.OService,
		//			SubGroup:  wire.OServiceServiceRequest,
		//		},
		//		Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
		//			FoodGroup: wire.ChatNav,
		//		},
		//	}
		//	chatNavCh <- wire.SNACMessage{
		//		Frame: wire.SNACFrame{
		//			FoodGroup: wire.ChatNav,
		//			SubGroup:  wire.ChatNavCreateRoom,
		//		},
		//		Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		//			Exchange: 4,
		//			Cookie:   "create",
		//			TLVBlock: wire.TLVBlock{
		//				TLVList: wire.TLVList{
		//					wire.NewTLVBE(wire.ChatRoomTLVRoomName, "haha"),
		//				},
		//			},
		//		},
		//	}
		//
		//	if err := clientFlap.SendDataFrame([]byte(fmt.Sprintf("CHAT_JOIN:%s:%s", "10", "haha"))); err != nil {
		//		return fmt.Errorf("send sign on data frame failed: %w", err)
		//	}
		//case "toc_chat_leave":
		//	if err := clientFlap.SendDataFrame([]byte(fmt.Sprintf("CHAT_LEFT:%s", "10"))); err != nil {
		//		return fmt.Errorf("send sign on data frame failed: %w", err)
		//	}
		case "toc_set_info":
			if err := rt.BOSProxy.SetInfo(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("set info failed: %w", err)
			}
		case "toc_set_dir":
			if err := rt.BOSProxy.SetDir(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("set info failed: %w", err)
			}
		case "toc_set_idle":
			if err := rt.BOSProxy.SetIdle(ctx, sessBOS, elems); err != nil {
				return fmt.Errorf("set info failed: %w", err)
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
	line = bytes.TrimSpace(line)

	if string(line) != "FLAPON" {
		return fmt.Errorf("unexpected line: %s", string(line))
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

func (rt Server) initChatNav(ctx context.Context, host string, cookie []byte, clientCh chan<- any, navCh <-chan wire.SNACMessage) error {
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	go func() {
		defer serverConn.Close()
		<-ctx.Done()
	}()

	rt.Logger.Info("connected to BOS server", "host", host)

	serverFlap := wire.NewFlapClient(0, serverConn, serverConn)

	if _, err := serverFlap.ReceiveSignonFrame(); err != nil {
		return err
	}

	tlv := []wire.TLV{
		wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
	}
	if err := serverFlap.SendSignonFrame(tlv); err != nil {
		return err
	}

	hostOnlineFrame := wire.SNACFrame{}
	hostOnlineSNAC := wire.SNAC_0x01_0x03_OServiceHostOnline{}
	if err := serverFlap.ReceiveSNAC(&hostOnlineFrame, &hostOnlineSNAC); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-navCh:
				if err := serverFlap.SendSNAC(msg.Frame, msg.Body); err != nil {
					rt.Logger.Error("send snac failed", "err", err)
					return
				}
			}
		}
	}()
	go func() {
		for {
			flap, err := serverFlap.ReceiveFLAP()
			if err != nil {
				if err != io.EOF {
					rt.Logger.Error("receive signon frame failed", "err", err)
				}
				return
			}
			clientCh <- flap
		}
	}()
	return nil
}

func (rt Server) initChatRoom(ctx context.Context, host string, cookie []byte, clientCh chan<- any, chatCh <-chan wire.SNACMessage) error {
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	go func() {
		defer serverConn.Close()
		<-ctx.Done()
	}()

	rt.Logger.Info("connected to BOS server", "host", host)

	serverFlap := wire.NewFlapClient(0, serverConn, serverConn)

	if _, err := serverFlap.ReceiveSignonFrame(); err != nil {
		return err
	}

	tlv := []wire.TLV{
		wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
	}
	if err := serverFlap.SendSignonFrame(tlv); err != nil {
		return err
	}

	hostOnlineFrame := wire.SNACFrame{}
	hostOnlineSNAC := wire.SNAC_0x01_0x03_OServiceHostOnline{}
	if err := serverFlap.ReceiveSNAC(&hostOnlineFrame, &hostOnlineSNAC); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-chatCh:
				if err := serverFlap.SendSNAC(msg.Frame, msg.Body); err != nil {
					rt.Logger.Error("send snac failed", "err", err)
					return
				}
			}
		}
	}()
	go func() {
		for {
			flap, err := serverFlap.ReceiveFLAP()
			if err != nil {
				if err != io.EOF {
					rt.Logger.Error("receive signon frame failed", "err", err)
				}
				return
			}
			clientCh <- flap
		}
	}()
	return nil
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
