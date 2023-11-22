package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type BuddyHandler interface {
	RightsQueryHandler(ctx context.Context) oscar.SNACMessage
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

func (rt *BuddyRouter) RouteBuddy(ctx context.Context, SNACFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.BuddyRightsQuery:
		inSNAC := oscar.SNAC_0x03_0x02_BuddyRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
		return sendSNAC(SNACFrame, outSNAC.Frame, outSNAC.Body, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}
