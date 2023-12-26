package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

// ChatServiceRouter is the interface that defines the entrypoint to the BOS service.
type ChatServiceRouter interface {
	// Route unmarshalls the SNAC frame header from the reader stream to
	// determine which food group to route to. The remainder of the reader
	// stream is passed on to the food group routers for the final SNAC body
	// extraction. Each response sent to the client via the writer stream
	// increments the sequence number.
	Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32, chatID string) error
}

// ChatService represents a service that implements a chat room session.
// Clients connect to this service upon creating a chat room or being invited
// to a chat room.
type ChatService struct {
	AuthHandler
	ChatServiceRouter
	Config
	OServiceChatRouter
}

// Start creates a TCP server that implements that chat flow.
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
			if err := rt.handleNewConnection(ctx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
			conn.Close()
		}()
	}
}

func (rt ChatService) handleNewConnection(ctx context.Context, rw io.ReadWriter) error {
	seq := uint32(100)

	flap, err := flapSignonHandshake(rw, &seq)
	if err != nil {
		return err
	}

	var ok bool
	buf, ok := flap.Slice(oscar.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	cookie := ChatCookie{}
	if err := oscar.Unmarshal(&cookie, bytes.NewBuffer(buf)); err != nil {
		return err
	}
	chatID := string(cookie.Cookie)

	chatSess, err := rt.RetrieveChatSession(chatID, cookie.SessID)
	if err != nil {
		return err
	}
	if chatSess == nil {
		return errors.New("session not found")
	}

	defer chatSess.Close()
	go func() {
		<-chatSess.Closed()
		if err := rt.SignoutChat(ctx, chatSess, chatID); err != nil {
			rt.Logger.ErrorContext(ctx, "unable to sign out user", "err", err.Error())
		}
	}()

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rw); err != nil {
		return err
	}

	fnClientReqHandler := func(ctx context.Context, r io.Reader, w io.Writer, seq *uint32) error {
		return rt.Route(ctx, chatSess, r, w, seq, chatID)
	}
	fnAlertHandler := func(ctx context.Context, msg oscar.SNACMessage, w io.Writer, seq *uint32) error {
		return sendSNAC(msg.Frame, msg.Body, seq, w)
	}
	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())
	dispatchIncomingMessages(ctx, chatSess, seq, rw, rt.Logger, fnClientReqHandler, fnAlertHandler)
	return nil
}
