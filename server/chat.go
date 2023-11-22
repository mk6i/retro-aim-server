package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type ChatHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, chatID string, snacPayloadIn oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.SNACMessage, error)
}

func NewChatRouter(logger *slog.Logger, chatHandler ChatHandler) ChatRouter {
	return ChatRouter{
		ChatHandler: chatHandler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ChatRouter struct {
	ChatHandler
	RouteLogger
}

func (rt *ChatRouter) RouteChat(ctx context.Context, sess *state.Session, chatID string, SNACFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ChatChannelMsgToHost:
		inSNAC := oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sess, chatID, inSNAC)
		if err != nil {
			return err
		}
		if outSNAC == nil {
			return nil
		}
		rt.Logger.InfoContext(ctx, "user sent a chat message")
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
