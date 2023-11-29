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
	ClientVersionsHandler(ctx context.Context, frame oscar.SNACFrame, bodyIn oscar.SNAC_0x01_0x17_OServiceClientVersions) oscar.SNACMessage
	IdleNotificationHandler(ctx context.Context, sess *state.Session, bodyIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQueryHandler(ctx context.Context, frame oscar.SNACFrame) oscar.SNACMessage
	RateParamsSubAddHandler(context.Context, oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	SetUserInfoFieldsHandler(ctx context.Context, sess *state.Session, frame oscar.SNACFrame, bodyIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (oscar.SNACMessage, error)
	UserInfoQueryHandler(ctx context.Context, sess *state.Session, frame oscar.SNACFrame) oscar.SNACMessage
}

type OServiceBOSHandler interface {
	OServiceHandler
	WriteOServiceHostOnline() oscar.SNACMessage
	ServiceRequestHandler(ctx context.Context, sess *state.Session, frame oscar.SNACFrame, bodyIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (oscar.SNACMessage, error)
	ClientOnlineHandler(ctx context.Context, bodyIn oscar.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error
}

type OServiceChatHandler interface {
	OServiceHandler
	WriteOServiceHostOnline() oscar.SNACMessage
	ClientOnlineHandler(ctx context.Context, sess *state.Session, chatID string) error
}

type OServiceRouter struct {
	OServiceHandler
	RouteLogger
}

func (rt OServiceRouter) RouteOService(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.OServiceRateParamsQuery:
		outSNAC := rt.RateParamsQueryHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.OServiceRateParamsSubAdd:
		inBody := oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.RateParamsSubAddHandler(ctx, inBody)
		rt.logRequest(ctx, inFrame, inBody)
		return oscar.Unmarshal(&inBody, r)
	case oscar.OServiceUserInfoQuery:
		outSNAC := rt.UserInfoQueryHandler(ctx, sess, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.OServiceIdleNotification:
		inBody := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.logRequest(ctx, inFrame, inBody)
		return rt.IdleNotificationHandler(ctx, sess, inBody)
	case oscar.OServiceClientVersions:
		inBody := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC := rt.ClientVersionsHandler(ctx, inFrame, inBody)
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.OServiceSetUserInfoFields:
		inBody := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.SetUserInfoFieldsHandler(ctx, sess, inFrame, inBody)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
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

func (rt OServiceBOSRouter) RouteOService(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		inBody := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(ctx, sess, inFrame, inBody)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(inFrame, w, sequence)
		case err != nil:
			return err
		}
		rt.logRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	case oscar.OServiceClientOnline:
		inBody := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, inFrame, inBody)
		return rt.OServiceBOSHandler.ClientOnlineHandler(ctx, inBody, sess)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, inFrame, r, w, sequence)
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

func (rt OServiceChatRouter) RouteOService(ctx context.Context, sess *state.Session, chatID string, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.OServiceServiceRequest:
		return sendInvalidSNACErr(inFrame, w, sequence)
	case oscar.OServiceClientOnline:
		inBody := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inBody, r); err != nil {
			return err
		}
		rt.Logger.InfoContext(ctx, "user signed on")
		rt.logRequest(ctx, inFrame, inBody)
		return rt.OServiceChatHandler.ClientOnlineHandler(ctx, sess, chatID)
	default:
		return rt.OServiceRouter.RouteOService(ctx, sess, inFrame, r, w, sequence)
	}
}
