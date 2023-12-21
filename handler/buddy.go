package handler

import (
	"context"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

func NewBuddyService() *BuddyService {
	return &BuddyService{}
}

type BuddyService struct {
}

func (s BuddyService) RightsQueryHandler(_ context.Context, frameIn oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyRightsReply,
			RequestID: frameIn.RequestID,
		},
		Body: oscar.SNAC_0x03_0x03_BuddyRightsReply{
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

func broadcastArrival(ctx context.Context, sess *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	screenNames, err := feedbagManager.InterestedUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	messageRelayer.RelayToScreenNames(ctx, screenNames, oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyArrived,
		},
		Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	})

	return nil
}

func broadcastDeparture(ctx context.Context, sess *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	screenNames, err := feedbagManager.InterestedUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	messageRelayer.RelayToScreenNames(ctx, screenNames, oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyDeparted,
		},
		Body: oscar.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   sess.ScreenName(),
				WarningLevel: sess.Warning(),
			},
		},
	})

	return nil
}

func unicastArrival(ctx context.Context, from *state.Session, to *state.Session, messageRelayer MessageRelayer) {
	messageRelayer.RelayToScreenName(ctx, to.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyArrived,
		},
		Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: from.TLVUserInfo(),
		},
	})
}

func unicastDeparture(ctx context.Context, from *state.Session, to *state.Session, messageRelayer MessageRelayer) {
	messageRelayer.RelayToScreenName(ctx, to.ScreenName(), oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyDeparted,
		},
		Body: oscar.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   from.ScreenName(),
				WarningLevel: from.Warning(),
			},
		},
	})
}
