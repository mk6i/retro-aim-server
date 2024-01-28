package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

type ChatHandler interface {
	ChannelMsgToHostHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*oscar.SNACMessage, error)
}

func NewChatRouter(logger *slog.Logger, chatHandler ChatHandler) ChatRouter {
	return ChatRouter{
		ChatHandler: chatHandler,
		routeLogger: routeLogger{
			Logger: logger,
		},
	}
}

type ChatRouter struct {
	ChatHandler
	routeLogger
}

func (rt ChatRouter) Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.ChatChannelMsgToHost:
		inBody := oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.ChannelMsgToHostHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		if outSNAC == nil {
			return nil
		}
		rt.Logger.InfoContext(ctx, "user sent a chat message")
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
