package handler

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/server/oscar"

	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type ChatNavService interface {
	CreateRoom(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (wire.SNACMessage, error)
	ExchangeInfo(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo) (wire.SNACMessage, error)
	RequestChatRights(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	RequestRoomInfo(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (wire.SNACMessage, error)
}

func NewChatNavHandler(chatNavService ChatNavService, logger *slog.Logger) ChatNavHandler {
	return ChatNavHandler{
		ChatNavService: chatNavService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type ChatNavHandler struct {
	ChatNavService
	middleware.RouteLogger
}

func (rt ChatNavHandler) RequestChatRights(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := rt.ChatNavService.RequestChatRights(ctx, inFrame)
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt ChatNavHandler) RequestExchangeInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.ChatNavService.ExchangeInfo(ctx, inFrame, inBody)
	if err != nil {
		return err
	}
	rt.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt ChatNavHandler) RequestRoomInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.ChatNavService.RequestRoomInfo(ctx, inFrame, inBody)
	if err != nil {
		return err
	}
	rt.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt ChatNavHandler) CreateRoom(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.ChatNavService.CreateRoom(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	roomName, _ := inBody.String(wire.ChatRoomTLVRoomName)
	rt.Logger.InfoContext(ctx, "user started a chat room", slog.String("roomName", roomName))
	rt.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
