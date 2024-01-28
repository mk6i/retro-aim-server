package server

import (
	"context"
	"io"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

// BOSRouter routes client connections to the OSCAR food group routers.
type BOSRouter struct {
	AlertRouter       Router
	BuddyRouter       Router
	ChatNavRouter     Router
	FeedbagRouter     Router
	ICBMRouter        Router
	LocateRouter      Router
	OServiceBOSRouter Router
}

// Route routes connections to the following food groups:
// - Alert
// - BUCP
// - Buddy
// - ChatNav
// - Feedbag
// - ICBM
// - Locate
// - OService
func (rt BOSRouter) Route(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch inFrame.FoodGroup {
	case oscar.OService:
		return rt.OServiceBOSRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.Locate:
		return rt.LocateRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.Buddy:
		return rt.BuddyRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.ICBM:
		return rt.ICBMRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.ChatNav:
		return rt.ChatNavRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.Feedbag:
		return rt.FeedbagRouter.Route(ctx, sess, inFrame, r, w, sequence)
	case oscar.Alert:
		return rt.AlertRouter.Route(ctx, sess, inFrame, r, w, sequence)
	default:
		return ErrUnsupportedSubGroup
	}
}
