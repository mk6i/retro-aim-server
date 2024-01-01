package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
)

type BuddyHandler interface {
	RightsQueryHandler(ctx context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage
}

func NewBuddyRouter(logger *slog.Logger, buddyHandler BuddyHandler) BuddyRouter {
	return BuddyRouter{
		BuddyHandler: buddyHandler,
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type BuddyRouter struct {
	BuddyHandler
	RouteLogger
}

func (rt *BuddyRouter) RouteBuddy(ctx context.Context, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.SubGroup {
	case oscar.BuddyRightsQuery:
		inSNAC := oscar.SNAC_0x03_0x02_BuddyRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx, inFrame)
		rt.logRequestAndResponse(ctx, inFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
