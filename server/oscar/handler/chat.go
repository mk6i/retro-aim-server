package handler

import (
	"context"
	"github.com/mk6i/retro-aim-server/server/oscar"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type ChatService interface {
	ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x05_ChatChannelMsgToHost) (*wire.SNACMessage, error)
}

func NewChatHandler(logger *slog.Logger, chatService ChatService) ChatHandler {
	return ChatHandler{
		ChatService: chatService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type ChatHandler struct {
	ChatService
	middleware.RouteLogger
}

func (rt ChatHandler) ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.ChatService.ChannelMsgToHost(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	if outSNAC == nil {
		return nil
	}
	rt.Logger.InfoContext(ctx, "user sent a chat message")
	rt.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
