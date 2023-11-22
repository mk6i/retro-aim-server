package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type ChatNavHandler interface {
	CreateRoomHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (oscar.SNACMessage, error)
	RequestChatRightsHandler(ctx context.Context) oscar.SNACMessage
	RequestRoomInfoHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (oscar.SNACMessage, error)
}

func NewChatNavRouter(handler ChatNavHandler, logger *slog.Logger) ChatNavRouter {
	return ChatNavRouter{
		ChatNavHandler: handler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type ChatNavRouter struct {
	ChatNavHandler
	RouteLogger
}

func (rt *ChatNavRouter) RouteChatNav(ctx context.Context, sess *state.Session, SNACFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ChatNavRequestChatRights:
		outSNAC := rt.RequestChatRightsHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame.RequestID, outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ChatNavRequestRoomInfo:
		inSNAC := oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.RequestRoomInfoHandler(ctx, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame.RequestID, outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.ChatNavCreateRoom:
		inSNAC := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.CreateRoomHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		roomName, _ := inSNAC.GetString(oscar.ChatTLVRoomName)
		rt.Logger.InfoContext(ctx, "user started a chat room", slog.String("roomName", roomName))
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame.RequestID, outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ChatCookie struct {
	Cookie []byte `len_prefix:"uint16"`
	SessID string `len_prefix:"uint16"`
}
