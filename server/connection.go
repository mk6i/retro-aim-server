package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

var (
	ErrUnsupportedSubGroup = errors.New("unimplemented subgroup, your client version may be unsupported")
)

type (
	incomingMessage struct {
		flap    oscar.FlapFrame
		payload *bytes.Buffer
	}
	alertHandler     func(ctx context.Context, msg oscar.XMessage, w io.Writer, u *uint32) error
	clientReqHandler func(ctx context.Context, r io.Reader, w io.Writer, u *uint32) error
)

func consumeFLAPFrames(r io.Reader, msgCh chan incomingMessage, errCh chan error) {
	defer close(msgCh)
	defer close(errCh)

	for {
		in := incomingMessage{}
		if err := oscar.Unmarshal(&in.flap, r); err != nil {
			errCh <- err
			return
		}

		if in.flap.FrameType == oscar.FlapFrameData {
			buf := make([]byte, in.flap.PayloadLength)
			if _, err := r.Read(buf); err != nil {
				errCh <- err
				return
			}
			in.payload = bytes.NewBuffer(buf)
		}

		msgCh <- in
	}
}

func dispatchIncomingMessages(ctx context.Context, sess *state.Session, seq uint32, rw io.ReadWriter, logger *slog.Logger, fn clientReqHandler, alertHandler alertHandler) {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(rw, msgCh, readErrCh)

	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	for {
		select {
		case m := <-msgCh:
			switch m.flap.FrameType {
			case oscar.FlapFrameData:
				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := fn(ctx, m.payload, rw, &seq); err != nil {
					return
				}
			case oscar.FlapFrameSignon:
				logger.ErrorContext(ctx, "shouldn't get FlapFrameSignon", "flap", m.flap)
			case oscar.FlapFrameError:
				logger.ErrorContext(ctx, "got FlapFrameError", "flap", m.flap)
				return
			case oscar.FlapFrameSignoff:
				logger.InfoContext(ctx, "got FlapFrameSignoff", "flap", m.flap)
				return
			case oscar.FlapFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				logger.ErrorContext(ctx, "got unknown FLAP frame type", "flap", m.flap)
				return
			}
		case m := <-sess.RecvMessage():
			// forward a notification sent from another client to this client
			if err := alertHandler(ctx, m, rw, &seq); err != nil {
				logRequestError(ctx, logger, m.SnacFrame, err)
				return
			}
			logRequest(ctx, logger, m.SnacFrame, m.SnacOut)
		case <-sess.Closed():
			// gracefully disconnect so that the client does not try to
			// reconnect when the connection closes.
			flap := oscar.FlapFrame{
				StartMarker:   42,
				FrameType:     oscar.FlapFrameSignoff,
				Sequence:      uint16(seq),
				PayloadLength: uint16(0),
			}
			if err := oscar.Marshal(flap, rw); err != nil {
				logger.ErrorContext(ctx, "unable to gracefully disconnect user", "err", err)
			}
			return
		case err := <-readErrCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return
		}
	}
}

func handleChatConnection(ctx context.Context, rw io.ReadWriter, serviceManager ChatServiceManager, logger *slog.Logger) {
	cookie, seq, err := serviceManager.VerifyChatLogin(rw)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	chatID := string(cookie.Cookie)

	chatSess, err := serviceManager.RetrieveChatSession(ctx, chatID, cookie.SessID)
	if err != nil {
		logger.ErrorContext(ctx, "unable to find chat room", "err", err.Error())
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		serviceManager.SignoutChat(ctx, chatSess, chatID)
	}()

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())

	msg := serviceManager.WriteOServiceHostOnline()
	if err := writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, &seq, rw); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
		return
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return serviceManager.Route(ctx, chatSess, r, w, seq, chatID)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, seq *uint32) error {
		return writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, seq, w)
	}
	dispatchIncomingMessages(ctx, chatSess, seq, rw, logger, fnClientReqHandler, fnAlertHandler)
}

func handleAuthConnection(authHandler AuthHandler, conn net.Conn) {
	defer conn.Close()
	seq := uint32(100)
	_, err := authHandler.SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, conn); err != nil {
		log.Println(err)
		return
	}
	b := make([]byte, flap.PayloadLength)
	if _, err := conn.Read(b); err != nil {
		log.Println(err)
		return
	}
	snac := oscar.SnacFrame{}
	buf := bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		log.Println(err)
		return
	}

	snacPayloadIn := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, buf); err != nil {
		log.Println(err)
		return
	}

	msg, err := authHandler.ReceiveAndSendAuthChallenge(snacPayloadIn, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}
	if err := writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, &seq, conn); err != nil {
		log.Println(err)
		return
	}

	flap = oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, conn); err != nil {
		log.Println(err)
		return
	}
	snac = oscar.SnacFrame{}
	b = make([]byte, flap.PayloadLength)
	if _, err := conn.Read(b); err != nil {
		log.Println(err)
		return
	}
	buf = bytes.NewBuffer(b)
	if err := oscar.Unmarshal(&snac, buf); err != nil {
		log.Println(err)
		return
	}

	snacPayloadIn2 := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn2, buf); err != nil {
		log.Println(err)
		return
	}

	msg, err = authHandler.ReceiveAndSendBUCPLoginRequest(snacPayloadIn2, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}
	if err := writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, &seq, conn); err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(ctx context.Context, conn net.Conn, serviceManager BOSServiceManager, logger *slog.Logger) {
	// todo why is conn net.Conn but handleChat is rw?
	sess, seq, err := serviceManager.VerifyLogin(conn)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	defer sess.Close()
	defer conn.Close()

	go func() {
		<-sess.Closed()
		if err := serviceManager.Signout(ctx, sess); err != nil {
			logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
		}
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName())

	msg := serviceManager.WriteOServiceHostOnline()
	if err := writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, &seq, conn); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
		return
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return serviceManager.Route(ctx, sess, r, w, seq)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, seq *uint32) error {
		return writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, seq, w)
	}
	dispatchIncomingMessages(ctx, sess, seq, conn, logger, fnClientReqHandler, fnAlertHandler)
}

func ListenChat(cfg Config, router ChatServiceManager, logger *slog.Logger) {
	addr := Address("", cfg.ChatPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind chat server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting service", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		logger.DebugContext(ctx, "accepted connection")
		go func() {
			handleChatConnection(ctx, conn, router, logger)
			conn.Close()
		}()
	}
}

func ListenBOS(cfg Config, router BOSServiceManager, logger *slog.Logger) {
	addr := Address("", cfg.BOSPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind BOS server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting service", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "ip", conn.RemoteAddr().String())
		logger.DebugContext(ctx, "accepted connection")
		go handleBOSConnection(ctx, conn, router, logger)
	}
}

func ListenBUCPLogin(cfg Config, err error, logger *slog.Logger, authHandler AuthHandler) {
	addr := Address("", cfg.OSCARPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("unable to bind OSCAR server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("starting OSCAR server", "addr", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleAuthConnection(authHandler, conn)
	}
}
