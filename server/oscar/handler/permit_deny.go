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

type PermitDenyService interface {
	RightsQuery(_ context.Context, frame wire.SNACFrame) wire.SNACMessage
}

func NewPermitDenyHandler(logger *slog.Logger, permitDenyService PermitDenyService) PermitDenyHandler {
	return PermitDenyHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
		PermitDenyService: permitDenyService,
	}
}

type PermitDenyHandler struct {
	PermitDenyService
	middleware.RouteLogger
}

func (rt PermitDenyHandler) RightsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := rt.PermitDenyService.RightsQuery(ctx, inFrame)
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt PermitDenyHandler) AddPermListEntries(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	rt.Logger.Debug("got a request for AddPermListEntries, but not doing anything about it right now")
	return nil
}
func (rt PermitDenyHandler) SetGroupPermitMask(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	rt.Logger.Debug("got a request for SetGroupPermitMask, but not doing anything about it right now")
	return nil
}
