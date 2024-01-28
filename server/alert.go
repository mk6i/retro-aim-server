package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

func NewAlertRouter(logger *slog.Logger) AlertRouter {
	return AlertRouter{
		routeLogger: routeLogger{
			Logger: logger,
		},
	}
}

type AlertRouter struct {
	routeLogger
}

func (rt AlertRouter) Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
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
