package server

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type ChatService struct {
	AuthHandler
	ChatRouter
	Config
	OServiceChatRouter
	RouteLogger
}

func (rt ChatService) Start() {
	addr := Address("", rt.Config.ChatPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		rt.Logger.Error("unable to bind chat server address", "err", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	rt.Logger.Info("starting chat service", "addr", addr)

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
			rt.handleNewConnection(ctx, conn)
			conn.Close()
		}()
	}
}

func (rt ChatService) handleNewConnection(ctx context.Context, rw io.ReadWriter) {
	cookie, seq, err := rt.VerifyChatLogin(rw)
	if err != nil {
		rt.Logger.ErrorContext(ctx, "user disconnected with error", "err", err.Error())
		return
	}

	chatID := string(cookie.Cookie)

	chatSess, err := rt.RetrieveChatSession(ctx, chatID, cookie.SessID)
	if err != nil {
		rt.Logger.ErrorContext(ctx, "unable to find chat room", "err", err.Error())
		return
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		rt.SignoutChat(ctx, chatSess, chatID)
	}()

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rw); err != nil {
		rt.Logger.ErrorContext(ctx, "error WriteOServiceHostOnline")
		return
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return rt.route(ctx, chatSess, r, w, seq, chatID)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, seq *uint32) error {
		return sendSNAC(msg.Frame, msg.Body, seq, w)
	}
	dispatchIncomingMessages(ctx, chatSess, seq, rw, rt.Logger, fnClientReqHandler, fnAlertHandler)
}

func (rt ChatService) route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32, chatID string) error {
	inFrame := oscar.SNACFrame{}
	if err := oscar.Unmarshal(&inFrame, r); err != nil {
		return err
	}

	err := func() error {
		switch inFrame.FoodGroup {
		case oscar.OService:
			return rt.RouteOService(ctx, sess, chatID, inFrame, r, w, sequence)
		case oscar.Chat:
			return rt.RouteChat(ctx, sess, chatID, inFrame, r, w, sequence)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, inFrame, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(inFrame, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Config.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}
