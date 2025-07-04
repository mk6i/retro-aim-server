package oscar

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type OServiceService interface {
	ClientOnline(ctx context.Context, service uint16, bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline, sess *state.Session) error
	ClientVersions(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x17_OServiceClientVersions) wire.SNACMessage
	HostOnline(service uint16) wire.SNACMessage
	IdleNotification(ctx context.Context, sess *state.Session, bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame) wire.SNACMessage
	RateParamsSubAdd(context.Context, *state.Session, wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	ServiceRequest(ctx context.Context, service uint16, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x04_OServiceServiceRequest, connectHere string) (wire.SNACMessage, error)
	SetPrivacyFlags(ctx context.Context, bodyIn wire.SNAC_0x01_0x14_OServiceSetPrivacyFlags)
	SetUserInfoFields(ctx context.Context, sess *state.Session, frame wire.SNACFrame, bodyIn wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (wire.SNACMessage, error)
	UserInfoQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame) wire.SNACMessage
}

func NewOServiceHandler(logger *slog.Logger, oServiceService OServiceService) OServiceHandler {
	return OServiceHandler{
		OServiceService: oServiceService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type OServiceHandler struct {
	OServiceService
	middleware.RouteLogger
}

func (h OServiceHandler) RateParamsQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw ResponseWriter) error {
	outSNAC := h.OServiceService.RateParamsQuery(ctx, sess, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) RateParamsSubAdd(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.OServiceService.RateParamsSubAdd(ctx, sess, inBody)
	h.LogRequest(ctx, inFrame, inBody)
	return nil
}

func (h OServiceHandler) UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw ResponseWriter) error {
	outSNAC := h.OServiceService.UserInfoQuery(ctx, sess, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) IdleNotification(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x11_OServiceIdleNotification{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, inBody)
	return h.OServiceService.IdleNotification(ctx, sess, inBody)
}

func (h OServiceHandler) ClientVersions(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x17_OServiceClientVersions{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.OServiceService.ClientVersions(ctx, sess, inFrame, inBody)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) SetUserInfoFields(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.OServiceService.SetUserInfoFields(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) Noop(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw ResponseWriter) error {
	// no-op keep-alive
	h.LogRequest(ctx, inFrame, nil)
	return nil
}

func (h OServiceHandler) SetPrivacyFlags(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x14_OServiceSetPrivacyFlags{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.OServiceService.SetPrivacyFlags(ctx, inBody)
	h.LogRequest(ctx, inFrame, inBody)
	return nil
}

func (h OServiceHandler) ServiceRequest(ctx context.Context, service uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter, connectHere string) error {
	inBody := wire.SNAC_0x01_0x04_OServiceServiceRequest{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.OServiceService.ServiceRequest(ctx, service, sess, inFrame, inBody, connectHere)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h OServiceHandler) ClientOnline(ctx context.Context, service uint16, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ ResponseWriter) error {
	inBody := wire.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.Logger.InfoContext(ctx, "user signed on")
	h.LogRequest(ctx, inFrame, inBody)
	return h.OServiceService.ClientOnline(ctx, service, inBody, sess)
}
