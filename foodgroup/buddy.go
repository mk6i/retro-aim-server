package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewBuddyService creates a new instance of BuddyService.
func NewBuddyService() *BuddyService {
	return &BuddyService{}
}

// BuddyService provides functionality for the Buddy food group, which sends
// clients notifications about the state of users on their buddy list. The food
// group is used by old versions of AIM not currently supported by Retro Aim
// Server. BuddyService just exists to satisfy AIM 5.x's buddy rights requests.
// It may be expanded in the future to support older versions of AIM.
type BuddyService struct {
}

// RightsQuery returns buddy list service parameters.
func (s BuddyService) RightsQuery(_ context.Context, frameIn wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
			RequestID: frameIn.RequestID,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
				},
			},
		},
	}
}

func broadcastArrival(ctx context.Context, sess *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	screenNames, err := feedbagManager.AdjacentUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	messageRelayer.RelayToScreenNames(ctx, screenNames, wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyArrived,
		},
		Body: wire.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	})

	return nil
}

func broadcastDeparture(ctx context.Context, sess *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	screenNames, err := feedbagManager.AdjacentUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	messageRelayer.RelayToScreenNames(ctx, screenNames, wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDeparted,
		},
		Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: wire.TLVUserInfo{
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
	messageRelayer.RelayToScreenName(ctx, to.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyArrived,
		},
		Body: wire.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: from.TLVUserInfo(),
		},
	})
}

func unicastDeparture(ctx context.Context, from *state.Session, to *state.Session, messageRelayer MessageRelayer) {
	messageRelayer.RelayToScreenName(ctx, to.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDeparted,
		},
		Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: wire.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   from.ScreenName(),
				WarningLevel: from.Warning(),
			},
		},
	})
}
