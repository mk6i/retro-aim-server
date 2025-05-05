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

func NewStatsHandler(logger *slog.Logger) StatsHandler {
	return StatsHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type StatsHandler struct {
	middleware.RouteLogger
}

func (h StatsHandler) ReportEvents(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0B_0x03_StatsReportEvents{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}

	snac := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Stats,
			SubGroup:  wire.StatsReportAck,
		},
		Body: wire.SNAC_0x0B_0x04_StatsReportAck{},
	}

	h.LogRequestAndResponse(ctx, inFrame, inBody, snac.Frame, snac.Body)

	return rw.SendSNAC(snac.Frame, snac.Body)
}
