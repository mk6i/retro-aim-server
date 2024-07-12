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

type LocateService interface {
	RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	SetDirInfo(ctx context.Context, frame wire.SNACFrame) wire.SNACMessage
	SetInfo(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x02_0x04_LocateSetInfo) error
	SetKeywordInfo(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x02_0x05_LocateUserInfoQuery) (wire.SNACMessage, error)
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
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, inBody)
	return h.LocateService.SetInfo(ctx, sess, inBody)
}

func (h LocateHandler) SetDirInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x09_LocateSetDirInfo{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.LocateService.SetDirInfo(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) GetDirInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x0B_LocateGetDirInfo{}
	h.LogRequest(ctx, inFrame, inBody)
	return wire.UnmarshalBE(&inBody, r)
}

func (h LocateHandler) SetKeywordInfo(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.LocateService.SetKeywordInfo(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) UserInfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x05_LocateUserInfoQuery{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.LocateService.UserInfoQuery(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h LocateHandler) UserInfoQuery2(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x02_0x15_LocateUserInfoQuery2{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	// SNAC functionality for LocateUserInfoQuery and LocateUserInfoQuery2 is
	// identical except for the Type field data type (uint16 vs uint32).
	wrappedBody := wire.SNAC_0x02_0x05_LocateUserInfoQuery{
		Type:       uint16(inBody.Type2),
		ScreenName: inBody.ScreenName,
	}
	outSNAC, err := h.LocateService.UserInfoQuery(ctx, sess, inFrame, wrappedBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
