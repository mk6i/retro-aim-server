package handler

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
)

func NewBuddyService() *BuddyService {
	return &BuddyService{}
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

func broadcastArrival(ctx context.Context, sess *server.Session, sm server.SessionManager, fm server.FeedbagManager) error {
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

func broadcastDeparture(ctx context.Context, sess *server.Session, sm server.SessionManager, fm server.FeedbagManager) error {
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

func unicastArrival(ctx context.Context, srcScreenName, destScreenName string, sm server.SessionManager) {
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

func unicastDeparture(ctx context.Context, srcScreenName, destScreenName string, sm server.SessionManager) {
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
