package server

import (
	"io"

	"github.com/mkaminski/goaim/oscar"
)

type BuddyHandler interface {
	RightsQueryHandler() XMessage
}

func NewBuddyRouter() BuddyRouter {
	return BuddyRouter{
		BuddyHandler: BuddyService{},
	}
}

type BuddyRouter struct {
	BuddyHandler
}

func (rt *BuddyRouter) RouteBuddy(SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.BuddyRightsQuery:
		inSNAC := oscar.SNAC_0x03_0x02_BuddyRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler()
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type BuddyService struct {
}

func (s BuddyService) RightsQueryHandler() XMessage {
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

func BroadcastArrival(sess *Session, sm SessionManager, fm FeedbagManager) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
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

func BroadcastDeparture(sess *Session, sm SessionManager, fm *FeedbagStore) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
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

func UnicastArrival(srcScreenName, destScreenName string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}
	sm.SendToScreenName(destScreenName, XMessage{
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

func UnicastDeparture(srcScreenName, destScreenName string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(srcScreenName)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}

	sm.SendToScreenName(destScreenName, XMessage{
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
