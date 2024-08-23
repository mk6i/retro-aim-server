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

type ICBMService interface {
	ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x06_ICBMChannelMsgToHost) (*wire.SNACMessage, error)
	ClientEvent(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x14_ICBMClientEvent) error
	EvilRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x04_0x08_ICBMEvilRequest) (wire.SNACMessage, error)
	ParameterQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	ClientErr(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x04_0x0B_ICBMClientErr) error
}

func NewICBMHandler(logger *slog.Logger, icbmService ICBMService) ICBMHandler {
	return ICBMHandler{
		ICBMService: icbmService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type ICBMHandler struct {
	ICBMService
	middleware.RouteLogger
}

func (h ICBMHandler) AddParameters(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x04_0x02_ICBMAddParameters{}
	h.LogRequest(ctx, inFrame, inBody)
	return wire.UnmarshalBE(&inBody, r)
}

func (h ICBMHandler) ParameterQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := h.ICBMService.ParameterQuery(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, outSNAC, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h ICBMHandler) ChannelMsgToHost(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.ICBMService.ChannelMsgToHost(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.Logger.InfoContext(ctx, "user sent an IM", slog.String("recipient", inBody.ScreenName))
	if outSNAC == nil {
		return nil
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h ICBMHandler) EvilRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x04_0x08_ICBMEvilRequest{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.ICBMService.EvilRequest(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h ICBMHandler) ClientErr(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x04_0x0B_ICBMClientErr{}
	h.LogRequest(ctx, inFrame, inBody)
	err := wire.UnmarshalBE(&inBody, r)
	if err != nil {
		return err
	}
	return h.ICBMService.ClientErr(ctx, sess, inFrame, inBody)
}

func (h ICBMHandler) ClientEvent(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x04_0x14_ICBMClientEvent{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, inBody)
	return h.ICBMService.ClientEvent(ctx, sess, inFrame, inBody)
}
