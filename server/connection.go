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

func dispatchIncomingMessages(ctx context.Context, sess *Session, seq uint32, rw io.ReadWriter, logger *slog.Logger, fn clientReqHandler, alertHandler alertHandler) {
	// buffered so that the go routine has room to exit
	msgCh := make(chan incomingMessage, 1)
	readErrCh := make(chan error, 1)
	go consumeFLAPFrames(rw, msgCh, readErrCh)

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
			// handle a read error
			switch {
			case errors.Is(io.EOF, err):
				fallthrough
			default:
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return
		}
	}
}

func HandleChatConnection(ctx context.Context, cr *ChatRegistry, rw io.ReadWriter, router ChatServiceRouter, logger *slog.Logger) {
	cookie, seq, err := VerifyChatLogin(rw)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	room, chatSessMgr, err := cr.Retrieve(string(cookie.Cookie))
	if err != nil {
		logger.ErrorContext(ctx, "unable to find chat room", "err", err.Error())
		return
	}

	chatSess, found := chatSessMgr.Retrieve(cookie.SessID)
	if !found {
		logger.ErrorContext(ctx, "unable to find user for session", "sessID", cookie.SessID)
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		AlertUserLeft(ctx, chatSess, chatSessMgr)
		chatSessMgr.Remove(chatSess)
		cr.MaybeRemoveRoom(room.Cookie)
	}()

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())

	if err := router.WriteOServiceHostOnline(rw, &seq); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return router.Route(ctx, chatSess, r, w, seq, chatSessMgr, room)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, seq *uint32) error {
		return writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, seq, w)
	}
	dispatchIncomingMessages(ctx, chatSess, seq, rw, logger, fnClientReqHandler, fnAlertHandler)
}

func HandleAuthConnection(cfg Config, sm *InMemorySessionManager, fm *SQLiteFeedbagStore, conn net.Conn) {
	defer conn.Close()
	seq := uint32(100)
	_, err := SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	err = ReceiveAndSendAuthChallenge(cfg, fm, conn, conn, &seq, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}

	err = ReceiveAndSendBUCPLoginRequest(cfg, sm, fm, conn, conn, &seq, uuid.New)
	if err != nil {
		log.Println(err)
		return
	}
}

func HandleBOSConnection(ctx context.Context, conn net.Conn, router BOSServiceRouter, logger *slog.Logger) {
	sess, seq, err := router.VerifyLogin(conn)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	defer sess.Close()
	defer conn.Close()

	go func() {
		<-sess.Closed()
		router.Signout(ctx, logger, sess)
	}()

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName())

	if err := router.WriteOServiceHostOnline(conn, &seq); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return router.Route(ctx, sess, r, w, seq)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.XMessage, w io.Writer, seq *uint32) error {
		return writeOutSNAC(oscar.SnacFrame{}, msg.SnacFrame, msg.SnacOut, seq, w)
	}
	dispatchIncomingMessages(ctx, sess, seq, conn, logger, fnClientReqHandler, fnAlertHandler)
}

func ListenChat(cfg Config, router ChatServiceRouter, cr *ChatRegistry, logger *slog.Logger) {
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
			HandleChatConnection(ctx, cr, conn, router, logger)
			conn.Close()
		}()
	}
}

func ListenBOS(cfg Config, router BOSServiceRouter, logger *slog.Logger) {
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
		go HandleBOSConnection(ctx, conn, router, logger)
	}
}

func ListenBUCPLogin(cfg Config, err error, logger *slog.Logger, sm *InMemorySessionManager, fm *SQLiteFeedbagStore) {
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

		go HandleAuthConnection(cfg, sm, fm, conn)
	}
}
