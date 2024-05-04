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

func NewAlertHandler(logger *slog.Logger) AlertHandler {
	return AlertHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

// AlertHandler just reads the request to placate the client. No need to send a
// response.
type AlertHandler struct {
	middleware.RouteLogger
}

func (rt AlertHandler) NotifyCapabilities(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, _ oscar.ResponseWriter) error {
	rt.LogRequest(ctx, inFrame, nil)
	return nil
}

func (rt AlertHandler) NotifyDisplayCapabilities(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, _ oscar.ResponseWriter) error {
	rt.LogRequest(ctx, inFrame, nil)
	return nil
}
