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

type OServiceService interface {
	ClientVersions(ctx context.Context, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x17_OServiceClientVersions) wire.SNACMessage
	IdleNotification(ctx context.Context, sess *state.Session, bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQuery(ctx context.Context, frame wire.SNACFrame) wire.SNACMessage
	RateParamsSubAdd(context.Context, wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	SetUserInfoFields(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (wire.SNACMessage, error)
	UserInfoQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame) wire.SNACMessage
}

type OServiceBOSService interface {
	OServiceService
	HostOnline() wire.SNACMessage
	ServiceRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest) (wire.SNACMessage, error)
	ClientOnline(ctx context.Context, bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error
}

type OServiceChatService interface {
	OServiceService
	HostOnline() wire.SNACMessage
	ClientOnline(ctx context.Context, sess *state.Session) error
}

type OServiceChatNavService interface {
	OServiceService
	HostOnline() wire.SNACMessage
}

type OServiceAlertService interface {
	OServiceService
	HostOnline() wire.SNACMessage
}

type OServiceHandler struct {
	OServiceService
	middleware.RouteLogger
}

func (h OServiceHandler) RateParamsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := h.OServiceService.RateParamsQuery(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) RateParamsSubAdd(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	h.OServiceService.RateParamsSubAdd(ctx, inBody)
	h.LogRequest(ctx, inFrame, inBody)
	return wire.Unmarshal(&inBody, r)
}

func (h OServiceHandler) UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := h.OServiceService.UserInfoQuery(ctx, sess, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) IdleNotification(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x11_OServiceIdleNotification{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, inBody)
	return h.OServiceService.IdleNotification(ctx, sess, inBody)
}

func (h OServiceHandler) ClientVersions(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x17_OServiceClientVersions{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.OServiceService.ClientVersions(ctx, inFrame, inBody)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) SetUserInfoFields(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.OServiceService.SetUserInfoFields(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) Noop(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	// no-op keep-alive
	h.LogRequest(ctx, inFrame, nil)
	return nil
}

func NewOServiceHandlerForBOS(logger *slog.Logger, oServiceService OServiceService, oServiceBOSService OServiceBOSService) OServiceBOSHandler {
	return OServiceBOSHandler{
		OServiceHandler: OServiceHandler{
			OServiceService: oServiceService,
			RouteLogger: middleware.RouteLogger{
				Logger: logger,
			},
		},
		OServiceBOSService: oServiceBOSService,
	}
}

type OServiceBOSHandler struct {
	OServiceHandler
	OServiceBOSService
}

func (s OServiceBOSHandler) ServiceRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x04_OServiceServiceRequest{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := s.OServiceBOSService.ServiceRequest(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	s.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (s OServiceBOSHandler) ClientOnline(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	s.Logger.InfoContext(ctx, "user signed on")
	s.LogRequest(ctx, inFrame, inBody)
	return s.OServiceBOSService.ClientOnline(ctx, inBody, sess)
}

func NewOServiceHandlerForChat(logger *slog.Logger, oServiceService OServiceService, oServiceChatService OServiceChatService) OServiceChatHandler {
	return OServiceChatHandler{
		OServiceHandler: OServiceHandler{
			OServiceService: oServiceService,
			RouteLogger: middleware.RouteLogger{
				Logger: logger,
			},
		},
		OServiceChatService: oServiceChatService,
	}
}

type OServiceChatHandler struct {
	OServiceHandler
	OServiceChatService
}

func (s OServiceChatHandler) ClientOnline(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	s.Logger.InfoContext(ctx, "user signed on")
	s.LogRequest(ctx, inFrame, inBody)
	return s.OServiceChatService.ClientOnline(ctx, sess)
}

func NewOServiceHandlerForChatNav(logger *slog.Logger, oServiceService OServiceService, oServiceChatNavService OServiceChatNavService) OServiceChatNavHandler {
	return OServiceChatNavHandler{
		OServiceHandler: OServiceHandler{
			OServiceService: oServiceService,
			RouteLogger: middleware.RouteLogger{
				Logger: logger,
			},
		},
		OServiceChatNavService: oServiceChatNavService,
	}
}

type OServiceChatNavHandler struct {
	OServiceHandler
	OServiceChatNavService
}

func (s OServiceChatNavHandler) ClientOnline(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	s.Logger.InfoContext(ctx, "user signed on")
	s.LogRequest(ctx, inFrame, inBody)
	return nil
}

func NewOServiceHandlerForAlert(logger *slog.Logger, oServiceService OServiceService, oServiceAlertService OServiceAlertService) OServiceAlertHandler {
	return OServiceAlertHandler{
		OServiceHandler: OServiceHandler{
			OServiceService: oServiceService,
			RouteLogger: middleware.RouteLogger{
				Logger: logger,
			},
		},
		OServiceAlertService: oServiceAlertService,
	}
}

type OServiceAlertHandler struct {
	OServiceHandler
	OServiceAlertService
}

func (s OServiceAlertHandler) ClientOnline(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	s.Logger.InfoContext(ctx, "user signed on")
	s.LogRequest(ctx, inFrame, inBody)
	return nil
}
