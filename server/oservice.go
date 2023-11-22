package server

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

type OServiceHandler interface {
	ClientVersionsHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.XMessage
	IdleNotificationHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQueryHandler(ctx context.Context) oscar.XMessage
	RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	SetUserInfoFieldsHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.XMessage, error)
	UserInfoQueryHandler(ctx context.Context, sess *state.Session) oscar.XMessage
}

type OServiceBOSHandler interface {
	OServiceHandler
	WriteOServiceHostOnline() oscar.XMessage
	ServiceRequestHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.XMessage, error)
	ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error
}

type OServiceChatHandler interface {
	OServiceHandler
	WriteOServiceHostOnline() oscar.XMessage
	ClientOnlineHandler(ctx context.Context, snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session, chatID string) error
}

type OServiceRouter struct {
	OServiceHandler
	RouteLogger
}

func (rt OServiceRouter) RouteOService(ctx context.Context, sess *state.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceRateParamsQuery:
		outSNAC := rt.RateParamsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return sendSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceRateParamsSubAdd:
		inSNAC := oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.RateParamsSubAddHandler(ctx, inSNAC)
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.OServiceUserInfoQuery:
		outSNAC := rt.UserInfoQueryHandler(ctx, sess)
		rt.logRequestAndResponse(ctx, SNACFrame, nil, outSNAC.SnacFrame, outSNAC.SnacOut)
		return sendSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceIdleNotification:
		inSNAC := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.IdleNotificationHandler(ctx, sess, inSNAC)
	case oscar.OServiceClientVersions:
		inSNAC := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.ClientVersionsHandler(ctx, inSNAC)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return sendSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceSetUserInfoFields:
		inSNAC := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.SetUserInfoFieldsHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return sendSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

func NewOServiceRouterForBOS(logger *slog.Logger, oserviceHandler OServiceHandler, oserviceBOSHandler OServiceBOSHandler) OServiceBOSRouter {
	return OServiceBOSRouter{
		OServiceRouter: OServiceRouter{
			OServiceHandler: oserviceHandler,
			RouteLogger: RouteLogger{
				Logger: logger,
			},
		},
		OServiceBOSHandler: oserviceBOSHandler,
	}
}

type OServiceBOSRouter struct {
	OServiceRouter
	OServiceBOSHandler
}

func (rt OServiceBOSRouter) RouteOService(ctx context.Context, sess *state.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		inSNAC := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(ctx, sess, inSNAC)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(SNACFrame, w, sequence)
		case err != nil:
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return sendSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.OServiceBOSHandler.ClientOnlineHandler(ctx, inSNAC, sess)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, SNACFrame, r, w, sequence)
	}
}

func NewOServiceRouterForChat(logger *slog.Logger, oserviceHandler OServiceHandler, chatHandler OServiceChatHandler) OServiceChatRouter {
	return OServiceChatRouter{
		OServiceRouter: OServiceRouter{
			OServiceHandler: oserviceHandler,
			RouteLogger: RouteLogger{
				Logger: logger,
			},
		},
		OServiceChatHandler: chatHandler,
	}
}

type OServiceChatRouter struct {
	OServiceRouter
	OServiceChatHandler
}

func (rt OServiceChatRouter) RouteOService(ctx context.Context, sess *state.Session, chatID string, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		return sendInvalidSNACErr(SNACFrame, w, sequence)
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return rt.OServiceChatHandler.ClientOnlineHandler(ctx, inSNAC, sess, chatID)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, SNACFrame, r, w, sequence)
	}
}
