package server

import (
	"context"
	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
)

func NewAlertRouter(logger *slog.Logger) AlertRouter {
	return AlertRouter{
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type AlertRouter struct {
	RouteLogger
}

func (rt *AlertRouter) RouteAlert(ctx context.Context, inFrame oscar.SNACFrame) error {
	switch inFrame.SubGroup {
	case oscar.AlertNotifyCapabilities:
		fallthrough
	case oscar.AlertNotifyDisplayCapabilities:
		// just read the request to placate the client. no need to send a
		// response.
		rt.logRequest(ctx, inFrame, nil)
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}
