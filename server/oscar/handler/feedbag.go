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

type FeedbagService interface {
	DeleteItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x13_0x0A_FeedbagDeleteItem) (wire.SNACMessage, error)
	Query(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame) (wire.SNACMessage, error)
	QueryIfModified(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x13_0x05_FeedbagQueryIfModified) (wire.SNACMessage, error)
	RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	StartCluster(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x13_0x11_FeedbagStartCluster)
	UpsertItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, items []wire.FeedbagItem) (wire.SNACMessage, error)
	Use(ctx context.Context, sess *state.Session) error
}

func NewFeedbagHandler(logger *slog.Logger, feedbagService FeedbagService) FeedbagHandler {
	return FeedbagHandler{
		FeedbagService: feedbagService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type FeedbagHandler struct {
	FeedbagService
	middleware.RouteLogger
}

func (h FeedbagHandler) RightsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x02_FeedbagRightsQuery{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC := h.FeedbagService.RightsQuery(ctx, inFrame)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) Query(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC, err := h.FeedbagService.Query(ctx, sess, inFrame)
	if err != nil {
		return err
	}
	h.LogRequest(ctx, inFrame, outSNAC)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) QueryIfModified(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x05_FeedbagQueryIfModified{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.FeedbagService.QueryIfModified(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) Use(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, _ io.Reader, _ oscar.ResponseWriter) error {
	h.LogRequest(ctx, inFrame, nil)
	return h.FeedbagService.Use(ctx, sess)
}

func (h FeedbagHandler) InsertItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x08_FeedbagInsertItem{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.FeedbagService.UpsertItem(ctx, sess, inFrame, inBody.Items)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) UpdateItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x09_FeedbagUpdateItem{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.FeedbagService.UpsertItem(ctx, sess, inFrame, inBody.Items)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) DeleteItem(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x0A_FeedbagDeleteItem{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.FeedbagService.DeleteItem(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h FeedbagHandler) StartCluster(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, _ oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x13_0x11_FeedbagStartCluster{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	h.FeedbagService.StartCluster(ctx, inFrame, inBody)
	h.LogRequest(ctx, inFrame, inBody)
	return nil
}

func (h FeedbagHandler) EndCluster(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, _ oscar.ResponseWriter) error {
	h.LogRequest(ctx, inFrame, nil)
	return nil
}
