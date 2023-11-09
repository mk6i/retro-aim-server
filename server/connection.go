package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
)

var (
	CapChat, _ = uuid.MustParse("748F2420-6287-11D1-8222-444553540000").MarshalBinary()
)

var (
	ErrUnsupportedFoodGroup = errors.New("unimplemented food group, your client version may be unsupported")
	ErrUnsupportedSubGroup  = errors.New("unimplemented subgroup, your client version may be unsupported")
)

type IncomingMessage struct {
	flap oscar.FlapFrame
	snac oscar.SnacFrame
	buf  io.Reader
}

type XMessage struct {
	snacFrame oscar.SnacFrame
	snacOut   any
}

func readIncomingRequests(ctx context.Context, logger *slog.Logger, rw io.Reader, msgCh chan IncomingMessage, errCh chan error) {
	defer close(msgCh)
	defer close(errCh)

	for {
		flap := oscar.FlapFrame{}
		if err := oscar.Unmarshal(&flap, rw); err != nil {
			errCh <- err
			return
		}

		switch flap.FrameType {
		case oscar.FlapFrameSignon:
			errCh <- errors.New("shouldn't get FlapFrameSignon")
			return
		case oscar.FlapFrameData:
			b := make([]byte, flap.PayloadLength)
			if _, err := rw.Read(b); err != nil {
				errCh <- err
				return
			}

			snac := oscar.SnacFrame{}
			buf := bytes.NewBuffer(b)
			if err := oscar.Unmarshal(&snac, buf); err != nil {
				errCh <- err
				return
			}

			msgCh <- IncomingMessage{
				flap: flap,
				snac: snac,
				buf:  buf,
			}
		case oscar.FlapFrameError:
			errCh <- fmt.Errorf("got FlapFrameError: %v", flap)
			return
		case oscar.FlapFrameSignoff:
			errCh <- ErrSignedOff
			return
		case oscar.FlapFrameKeepAlive:
			logger.DebugContext(ctx, "keepalive heartbeat")
		default:
			errCh <- fmt.Errorf("unknown frame type: %v", flap)
			return
		}
	}
}

func Signout(ctx context.Context, logger *slog.Logger, sess *Session, sm SessionManager, fm *FeedbagStore) {
	if err := BroadcastDeparture(ctx, sess, sm, fm); err != nil {
		logger.ErrorContext(ctx, "error notifying departure", "err", err.Error())
	}
	sm.Remove(sess)
}

type RouteSig func(ctx context.Context, rwc io.ReadWriter, u *uint32, snac oscar.SnacFrame, buf io.Reader) error

func ReadBos(ctx context.Context, cfg Config, sess *Session, seq uint32, rwc io.ReadWriter, logger *slog.Logger, fn RouteSig) {
	// buffered so that the go routine has room to exit
	msgCh := make(chan IncomingMessage, 1)
	errCh := make(chan error, 1)
	go readIncomingRequests(ctx, logger, rwc, msgCh, errCh)

	rl := RouteLogger{
		Logger: logger,
	}

	for {
		select {
		case m := <-msgCh:
			if err := fn(ctx, rwc, &seq, m.snac, m.buf); err != nil {
				if errors.Is(err, ErrUnsupportedSubGroup) || errors.Is(err, ErrUnsupportedFoodGroup) {
					if err1 := sendInvalidSNACErr(m.snac, rwc, &seq); err1 != nil {
						err = errors.Join(err1, err)
					}
					if cfg.FailFast {
						panic(err.Error())
					}
				}
				logRequestError(ctx, logger, m.snac, err)
				return
			}
		case m := <-sess.RecvMessage():
			if err := writeOutSNAC(oscar.SnacFrame{}, m.snacFrame, m.snacOut, &seq, rwc); err != nil {
				logRequestError(ctx, logger, m.snacFrame, err)
				return
			}
			rl.logRequest(ctx, m.snacFrame, m.snacOut)
		case <-sess.Closed():
			if err := gracefulDisconnect(seq, rwc); err != nil {
				logger.ErrorContext(ctx, "unable to gracefully disconnect user", "err", err)
			}
			return
		case err := <-errCh:
			switch {
			case errors.Is(io.EOF, err):
				fallthrough
			case errors.Is(ErrSignedOff, err):
				logger.InfoContext(ctx, "client signed off")
			default:
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return
		}
	}
}

func logRequestError(ctx context.Context, logger *slog.Logger, inFrame oscar.SnacFrame, err error) {
	logger.LogAttrs(ctx, slog.LevelError, "client disconnected with error",
		slog.Group("request",
			slog.String("food_group", oscar.FoodGroupStr(inFrame.FoodGroup)),
			slog.String("sub_group", oscar.SubGroupStr(inFrame.FoodGroup, inFrame.SubGroup)),
		),
		slog.String("err", err.Error()),
	)
}

func gracefulDisconnect(seq uint32, rwc io.ReadWriter) error {
	return oscar.Marshal(oscar.FlapFrame{
		StartMarker: 42,
		FrameType:   oscar.FlapFrameSignoff,
		Sequence:    uint16(seq),
	}, rwc)
}

func HandleChatConnection(ctx context.Context, cfg Config, cr *ChatRegistry, conn net.Conn, router ChatServiceRouter, logger *slog.Logger) {
	cookie, seq, err := VerifyChatLogin(conn)
	if err != nil {
		logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	room, err := cr.Retrieve(string(cookie.Cookie))
	if err != nil {
		logger.ErrorContext(ctx, "unable to find chat room", "err", err.Error())
		return
	}

	chatSess, found := room.Retrieve(cookie.SessID)
	if !found {
		logger.ErrorContext(ctx, "unable to find user for session", "sessID", cookie.SessID)
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		AlertUserLeft(ctx, chatSess, room)
		room.Remove(chatSess)
		cr.MaybeRemoveRoom(room.Cookie)
		conn.Close()
	}()

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName)

	if err := router.WriteOServiceHostOnline(conn, &seq); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
	}

	fn := func(ctx context.Context, rwc io.ReadWriter, u *uint32, snac oscar.SnacFrame, buf io.Reader) error {
		return router.Route(ctx, chatSess, rwc, u, snac, buf, room)
	}
	ReadBos(ctx, cfg, chatSess, seq, conn, logger, fn)
}

func HandleAuthConnection(cfg Config, sm *InMemorySessionManager, fm *FeedbagStore, conn net.Conn) {
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

func HandleBOSConnection(ctx context.Context, cfg Config, conn net.Conn, router BOSServiceRouter, logger *slog.Logger) {
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

	ctx = context.WithValue(ctx, "screenName", sess.ScreenName)

	if err := router.WriteOServiceHostOnline(conn, &seq); err != nil {
		logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
	}

	fn := func(ctx context.Context, rwc io.ReadWriter, u *uint32, snac oscar.SnacFrame, buf io.Reader) error {
		return router.Route(ctx, sess, rwc, u, snac, buf)
	}
	ReadBos(ctx, cfg, sess, seq, conn, logger, fn)
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
		go HandleChatConnection(ctx, cfg, cr, conn, router, logger)
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
		go HandleBOSConnection(ctx, cfg, conn, router, logger)
	}
}

func ListenBUCPLogin(cfg Config, err error, logger *slog.Logger, sm *InMemorySessionManager, fm *FeedbagStore) {
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
