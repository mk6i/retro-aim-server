package server

import (
	"context"
	"errors"
	"io"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

// BOSRootRouter routes client connections to the OSCAR food group routers.
type BOSRootRouter struct {
	AlertRouter
	BuddyRouter
	ChatNavRouter
	config.Config
	FeedbagRouter
	ICBMRouter
	LocateRouter
	OServiceBOSRouter
	RouteLogger
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
func (rt BOSRootRouter) Route(ctx context.Context, sess *state.Session, r io.Reader, w io.Writer, sequence *uint32) error {
	inFrame := oscar.SNACFrame{}
	if err := oscar.Unmarshal(&inFrame, r); err != nil {
		return err
	}

	err := func() error {
		switch inFrame.FoodGroup {
		case oscar.OService:
			return rt.RouteOService(ctx, sess, inFrame, r, w, sequence)
		case oscar.Locate:
			return rt.RouteLocate(ctx, sess, inFrame, r, w, sequence)
		case oscar.Buddy:
			return rt.RouteBuddy(ctx, inFrame, r, w, sequence)
		case oscar.ICBM:
			return rt.RouteICBM(ctx, sess, inFrame, r, w, sequence)
		case oscar.ChatNav:
			return rt.RouteChatNav(ctx, sess, inFrame, r, w, sequence)
		case oscar.Feedbag:
			return rt.RouteFeedbag(ctx, sess, inFrame, r, w, sequence)
		case oscar.BUCP:
			return ErrUnsupportedSubGroup
		case oscar.Alert:
			return rt.RouteAlert(ctx, inFrame)
		default:
			return ErrUnsupportedSubGroup
		}
	}()

	if err != nil {
		rt.logRequestError(ctx, inFrame, err)
		if errors.Is(err, ErrUnsupportedSubGroup) {
			if err1 := sendInvalidSNACErr(inFrame, w, sequence); err1 != nil {
				err = errors.Join(err1, err)
			}
			if rt.Config.FailFast {
				panic(err.Error())
			}
			return nil
		}
	}

	return err
}
