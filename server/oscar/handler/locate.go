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

type LocateService interface {
	RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	SetDirInfo(ctx context.Context, frame wire.SNACFrame) wire.SNACMessage
	SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfo(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	UserInfoQuery2(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x15_LocateUserInfoQuery2) (wire.SNACMessage, error)
}

func NewLocateHandler(locateService LocateService, logger *slog.Logger) LocateHandler {
	return LocateHandler{
		LocateService: locateService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type LocateHandler struct {
	LocateService
	middleware.RouteLogger
}

func (h LocateHandler) RightsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := h.LocateService.RightsQuery(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) SetInfo(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x04_LocateSetInfo{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, inBody)
	return h.LocateService.SetInfo(ctx, sess, inBody)
}

func (h LocateHandler) SetDirInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x09_LocateSetDirInfo{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.LocateService.SetDirInfo(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) GetDirInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x0B_LocateGetDirInfo{}
	h.LogRequest(ctx, inFrame, inBody)
	return wire.Unmarshal(&inBody, r)
}

func (h LocateHandler) SetKeywordInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.LocateService.SetKeywordInfo(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) UserInfoQuery2(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x15_LocateUserInfoQuery2{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.LocateService.UserInfoQuery2(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
