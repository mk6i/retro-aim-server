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
func NewBuddyService(
	messageRelayer MessageRelayer,
	feedbagManager FeedbagManager,
	legacyBuddyListManager LegacyBuddyListManager,
) *BuddyService {
	return &BuddyService{
		feedbagManager:         feedbagManager,
		legacyBuddyListManager: legacyBuddyListManager,
		messageRelayer:         messageRelayer,
	}
}

// BuddyService provides functionality for the Buddy food group, which sends
// clients notifications about the state of users on their buddy list. The food
// group is used by old versions of AIM not currently supported by Retro Aim
// Server. BuddyService just exists to satisfy AIM 5.x's buddy rights requests.
// It may be expanded in the future to support older versions of AIM.
type BuddyService struct {
	feedbagManager         FeedbagManager
	legacyBuddyListManager LegacyBuddyListManager
	messageRelayer         MessageRelayer
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

func (s BuddyService) AddBuddies(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x04_BuddyAddBuddies) error {
	for _, entry := range inBody.Buddies {
		s.legacyBuddyListManager.AddBuddy(sess.IdentScreenName(), state.NewIdentScreenName(entry.ScreenName))
		if !sess.SignonComplete() {
			// client has not completed sign-on sequence, so any arrival
			// messages sent at this point would be ignored by the client.
			continue
		}
		buddy := s.messageRelayer.RetrieveByScreenName(state.NewIdentScreenName(entry.ScreenName))
		if buddy == nil || buddy.Invisible() {
			continue
		}
		// notify that buddy is online
		if err := s.UnicastBuddyArrived(ctx, buddy, sess); err != nil {
			return err
		}
	}
	return nil
}

func (s BuddyService) DelBuddies(_ context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x05_BuddyDelBuddies) {
	for _, entry := range inBody.Buddies {
		s.legacyBuddyListManager.DeleteBuddy(sess.IdentScreenName(), state.NewIdentScreenName(entry.ScreenName))
	}
}

// UnicastBuddyArrived sends the latest user info to a particular user.
// While updates are sent via the wire.BuddyArrived SNAC, the message is not
// only used to indicate the user coming online. It can also notify changes to
// buddy icons, warning levels, invisibility status, etc.
func (s BuddyService) UnicastBuddyArrived(ctx context.Context, from *state.Session, to *state.Session) error {
	userInfo := from.TLVUserInfo()
	icon, err := getBuddyIconRefFromFeedbag(from, s.feedbagManager)
	switch {
	case err != nil:
		return err
	case icon != nil:
		userInfo.Append(wire.NewTLV(wire.OServiceUserInfoBARTInfo, *icon))
	}
	s.messageRelayer.RelayToScreenName(ctx, to.IdentScreenName(), wire.SNACMessage{
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

// BroadcastBuddyArrived sends the latest user info to the user's adjacent users.
// While updates are sent via the wire.BuddyArrived SNAC, the message is not
// only used to indicate the user coming online. It can also notify changes to
// buddy icons, warning levels, invisibility status, etc.
func (s BuddyService) BroadcastBuddyArrived(ctx context.Context, sess *state.Session) error {
	// find users who have this user on their server-side buddy list
	recipients, err := s.feedbagManager.AdjacentUsers(sess.IdentScreenName())
	if err != nil {
		return err
	}

	// find users who have this user on their client-side buddy list
	legacyUsers := s.legacyBuddyListManager.WhoAddedUser(sess.IdentScreenName())
	recipients = append(recipients, legacyUsers...)

	userInfo := sess.TLVUserInfo()
	icon, err := getBuddyIconRefFromFeedbag(sess, s.feedbagManager)
	switch {
	case err != nil:
		return err
	case icon != nil:
		userInfo.Append(wire.NewTLV(wire.OServiceUserInfoBARTInfo, *icon))
	}

	s.messageRelayer.RelayToScreenNames(ctx, recipients, wire.SNACMessage{
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
	items, err := feedbagManager.Feedbag(sess.IdentScreenName())
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ClassID != wire.FeedbagClassIdBart {
			continue
		}
		bartType, err := extractBARTItemType(item)
		if err != nil {
			return nil, err
		}
		if bartType != wire.BARTTypesBuddyIcon {
			continue
		}
		b, hasBuf := item.Slice(wire.FeedbagAttributesBartInfo)
		if !hasBuf {
			return nil, errors.New("unable to extract icon payload")
		}
		bartInfo := wire.BARTInfo{}
		if err := wire.UnmarshalBE(&bartInfo, bytes.NewBuffer(b)); err != nil {
			return nil, err
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

func (s BuddyService) BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error {
	recipients, err := s.feedbagManager.AdjacentUsers(sess.IdentScreenName())
	if err != nil {
		return err
	}

	legacyUsers := s.legacyBuddyListManager.WhoAddedUser(sess.IdentScreenName())
	recipients = append(recipients, legacyUsers...)

	s.messageRelayer.RelayToScreenNames(ctx, recipients, wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDeparted,
		},
		Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: wire.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   string(sess.DisplayScreenName()),
				WarningLevel: sess.Warning(),
			},
		},
	})

	return nil
}

func (s BuddyService) UnicastBuddyDeparted(ctx context.Context, from *state.Session, to *state.Session) {
	s.messageRelayer.RelayToScreenName(ctx, to.IdentScreenName(), wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDeparted,
		},
		Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: wire.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   string(from.DisplayScreenName()),
				WarningLevel: from.Warning(),
			},
		},
	})
}
