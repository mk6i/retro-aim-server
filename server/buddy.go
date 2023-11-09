package server

import (
	"context"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type BuddyHandler interface {
	RightsQueryHandler(ctx context.Context) XMessage
}

func NewBuddyRouter(logger *slog.Logger) BuddyRouter {
	return BuddyRouter{
		BuddyHandler: BuddyService{},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type BuddyRouter struct {
	BuddyHandler
	RouteLogger
}

func (rt *BuddyRouter) RouteBuddy(ctx context.Context, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.BuddyRightsQuery:
		inSNAC := oscar.SNAC_0x03_0x02_BuddyRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.snacFrame, outSNAC.snacOut)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type BuddyService struct {
}

func (s BuddyService) RightsQueryHandler(context.Context) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyRightsReply,
		},
		snacOut: oscar.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x01, uint16(100)),
					oscar.NewTLV(0x02, uint16(100)),
					oscar.NewTLV(0x03, uint16(100)),
					oscar.NewTLV(0x04, uint16(100)),
				},
			},
		},
	}
}

func BroadcastArrival(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(ctx, screenNames, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyArrived,
		},
		snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
				TLVBlock: oscar.TLVBlock{
					TLVList: sess.GetUserInfo(),
				},
			},
		},
	})

	return nil
}

func BroadcastDeparture(ctx context.Context, sess *Session, sm SessionManager, fm FeedbagManager) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(ctx, screenNames, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyDeparted,
		},
		snacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
			},
		},
	})

	return nil
}

func UnicastArrival(ctx context.Context, srcScreenName, destScreenName string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}
	sm.SendToScreenName(ctx, destScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyArrived,
		},
		snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	})

	return nil
}

func UnicastDeparture(ctx context.Context, srcScreenName, destScreenName string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}

	sm.SendToScreenName(ctx, destScreenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyDeparted,
		},
		snacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
			},
		},
	})

	return nil
}
