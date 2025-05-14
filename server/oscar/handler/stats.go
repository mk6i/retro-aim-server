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

type StatsService interface {
	ReportEvents(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0B_0x03_StatsReportEvents) wire.SNACMessage
}

func NewStatsHandler(logger *slog.Logger, statsService StatsService) StatsHandler {
	return StatsHandler{
		StatsService: statsService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type StatsHandler struct {
	StatsService
	middleware.RouteLogger
}

func (h StatsHandler) ReportEvents(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0B_0x03_StatsReportEvents{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}

	outSNAC := h.StatsService.ReportEvents(ctx, inFrame, inBody)
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)

	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
