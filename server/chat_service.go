package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
)

// ChatService represents a service that implements a chat room session.
// Clients connect to this service upon creating a chat room or being invited
// to a chat room.
type ChatService struct {
	AuthHandler
	OServiceChatHandler
	Router
	config.Config
	Logger *slog.Logger
}

// Start creates a TCP server that implements that chat flow.
func (rt ChatService) Start() {
	addr := config.Address("", rt.Config.ChatPort)
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
		}()
	}
}

func (rt ChatService) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {
	seq := uint32(100)

	flap, err := flapSignonHandshake(rwc, &seq)
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

	defer func() {
		chatSess.Close()
		rwc.Close()
		if err := rt.SignoutChat(ctx, chatSess, chatID); err != nil {
			rt.Logger.ErrorContext(ctx, "unable to sign out user", "err", err.Error())
		}
	}()

	msg := rt.WriteOServiceHostOnline()
	if err := sendSNAC(msg.Frame, msg.Body, &seq, rwc); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "screenName", chatSess.ScreenName())
	return dispatchIncomingMessages(ctx, chatSess, seq, rwc, rt.Logger, rt.Router, sendSNAC, rt.Config)
}
