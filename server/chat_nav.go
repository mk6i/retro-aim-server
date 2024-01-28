package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

type ChatNavHandler interface {
	CreateRoomHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (oscar.SNACMessage, error)
	RequestChatRightsHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
	RequestRoomInfoHandler(ctx context.Context, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (oscar.SNACMessage, error)
}

func NewChatNavRouter(handler ChatNavHandler, logger *slog.Logger) ChatNavRouter {
	return ChatNavRouter{
		ChatNavHandler: handler,
		routeLogger: routeLogger{
			Logger: logger,
		},
	}
}

type ChatNavRouter struct {
	ChatNavHandler
	routeLogger
}

func (rt ChatNavRouter) Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.ChatNavRequestChatRights:
		outSNAC := rt.RequestChatRightsHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ChatNavRequestRoomInfo:
		inBody := oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.RequestRoomInfoHandler(ctx, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ChatNavCreateRoom:
		inBody := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.CreateRoomHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		roomName, _ := inBody.String(oscar.ChatTLVRoomName)
		rt.Logger.InfoContext(ctx, "user started a chat room", slog.String("roomName", roomName))
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ChatCookie struct {
	Cookie []byte `len_prefix:"uint16"`
	SessID string `len_prefix:"uint16"`
}
