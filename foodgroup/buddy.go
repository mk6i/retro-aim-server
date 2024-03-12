package foodgroup

import (
	"bytes"
	"context"
	"errors"
	"strconv"

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

// broadcastArrival sends the latest user info to the user's adjacent users.
// While updates are sent via the wire.BuddyArrived SNAC, the message is not
// only used to indicate the user coming online. It can also notify changes to
// buddy icons, warning levels, invisibility status, etc.
func broadcastArrival(ctx context.Context, sess *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	screenNames, err := feedbagManager.AdjacentUsers(sess.ScreenName())
	if err != nil {
		return err
	}

	userInfo := sess.TLVUserInfo()
	icon, err := getBuddyIconRefFromFeedbag(sess, feedbagManager)
	switch {
	case err != nil:
		return err
	case icon != nil:
		userInfo.Append(wire.NewTLV(wire.OServiceUserInfoBARTInfo, *icon))
	}

	messageRelayer.RelayToScreenNames(ctx, screenNames, wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyArrived,
		},
		Body: wire.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: userInfo,
		},
	})

	return nil
}

// getBuddyIconRefFromFeedbag retrieves a reference to the user's buddy icon
// from their feedbag. If it exists, the buddy icon is the feedbag item of
// class wire.FeedbagClassIdBart with BART type wire.BARTTypesBuddyIcon.
func getBuddyIconRefFromFeedbag(sess *state.Session, feedbagManager FeedbagManager) (*wire.BARTID, error) {
	items, err := feedbagManager.Feedbag(sess.ScreenName())
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ClassID != wire.FeedbagClassIdBart {
			continue
		}
		b, hasBuf := item.Slice(wire.FeedbagAttributesBartInfo)
		if !hasBuf {
			return nil, errors.New("unable to extract icon payload")
		}
		bartInfo := wire.BARTInfo{}
		if err := wire.Unmarshal(&bartInfo, bytes.NewBuffer(b)); err != nil {
			return nil, err
		}
		bartType, err := extractBARTItemType(item)
		if err != nil {
			return nil, err
		}
		if bartType != wire.BARTTypesBuddyIcon {
			continue
		}
		if bartInfo.HasClearIconHash() {
			bartInfo.Flags = wire.BARTFlagsCustom | wire.BARTFlagsData
		}
		return &wire.BARTID{
			Type: bartType,
			BARTInfo: wire.BARTInfo{
				Flags: bartInfo.Flags,
				Hash:  bartInfo.Hash,
			},
		}, nil
	}

	return nil, nil
}

// extractBARTItemType gets the BART type for item, which is stored in the
// "name" field.
func extractBARTItemType(item wire.FeedbagItem) (uint16, error) {
	var bartType uint16
	// Feedbag items of type wire.FeedbagClassIdBart store the BART type in the
	// name field.
	if bt, err := strconv.ParseUint(item.Name, 10, 16); err != nil {
		return 0, err
	} else {
		bartType = uint16(bt)
	}
	return bartType, nil
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

// unicastArrival sends the latest user info to a particular user.
// While updates are sent via the wire.BuddyArrived SNAC, the message is not
// only used to indicate the user coming online. It can also notify changes to
// buddy icons, warning levels, invisibility status, etc.
func unicastArrival(ctx context.Context, from *state.Session, to *state.Session, messageRelayer MessageRelayer, feedbagManager FeedbagManager) error {
	userInfo := from.TLVUserInfo()
	icon, err := getBuddyIconRefFromFeedbag(from, feedbagManager)
	switch {
	case err != nil:
		return err
	case icon != nil:
		userInfo.Append(wire.NewTLV(wire.OServiceUserInfoBARTInfo, *icon))
	}
	messageRelayer.RelayToScreenName(ctx, to.ScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyArrived,
		},
		Body: wire.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: userInfo,
		},
	})
	return nil
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
