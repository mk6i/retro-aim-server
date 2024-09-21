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

type ODirService interface {
	InfoQuery(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0F_0x02_InfoQuery) (wire.SNACMessage, error)
	KeywordListQuery(context.Context, wire.SNACFrame) (wire.SNACMessage, error)
}

func NewODirHandler(logger *slog.Logger, oDirService ODirService) ODirHandler {
	return ODirHandler{
		ODirService: oDirService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type ODirHandler struct {
	ODirService
	middleware.RouteLogger
}

func (h ODirHandler) InfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0F_0x02_InfoQuery{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.ODirService.InfoQuery(ctx, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, outSNAC, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (h ODirHandler) KeywordListQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	outSNAC, err := h.ODirService.KeywordListQuery(ctx, inFrame)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
