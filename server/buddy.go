package server

import (
	"context"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"

	"github.com/mkaminski/goaim/oscar"
)

type BuddyHandler interface {
	RightsQueryHandler(ctx context.Context) oscar.XMessage
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
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type BuddyService struct {
}

func (s BuddyService) RightsQueryHandler(context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyRightsReply,
		},
		SnacOut: oscar.SNAC_0x03_0x03_BuddyRightsReply{
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

func BroadcastArrival(ctx context.Context, sess *user.Session, sm SessionManager, fm FeedbagManager) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(ctx, screenNames, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyArrived,
		},
		SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName(),
				WarningLevel: sess.Warning(),
				TLVBlock: oscar.TLVBlock{
					TLVList: sess.UserInfo(),
				},
			},
		},
	})

	return nil
}

func BroadcastDeparture(ctx context.Context, sess *user.Session, sm SessionManager, fm FeedbagManager) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(ctx, screenNames, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyDeparted,
		},
		SnacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName(),
				WarningLevel: sess.Warning(),
			},
		},
	})

	return nil
}

func UnicastArrival(ctx context.Context, srcScreenName, destScreenName string, sm SessionManager) {
	sess := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case sess == nil:
		fallthrough
	case sess.Invisible(): // don't tell user this buddy is online
		return
	}
	sm.SendToScreenName(ctx, destScreenName, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyArrived,
		},
		SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	})
}

func UnicastDeparture(ctx context.Context, srcScreenName, destScreenName string, sm SessionManager) {
	sess := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case sess == nil:
		fallthrough
	case sess.Invisible(): // don't tell user this buddy is online
		return
	}

	sm.SendToScreenName(ctx, destScreenName, oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.BUDDY,
			SubGroup:  oscar.BuddyDeparted,
		},
		SnacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   sess.ScreenName(),
				WarningLevel: sess.Warning(),
			},
		},
	})
}
