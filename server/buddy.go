package server

import (
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
)

const (
	BuddyErr                 uint16 = 0x0001
	BuddyRightsQuery                = 0x0002
	BuddyAddBuddies                 = 0x0004
	BuddyDelBuddies                 = 0x0005
	BuddyWatcherListQuery           = 0x0006
	BuddyWatcherSubRequest          = 0x0008
	BuddyWatcherNotification        = 0x0009
	BuddyRejectNotification         = 0x000A
	BuddyArrived                    = 0x000B
	BuddyDeparted                   = 0x000C
	BuddyAddTempBuddies             = 0x000F
	BuddyDelTempBuddies             = 0x0010
)

func routeBuddy(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {

	switch snac.SubGroup {
	case BuddyErr:
		panic("not implemented")
	case BuddyRightsQuery:
		return SendAndReceiveBuddyRights(snac, r, w, sequence)
	case BuddyAddBuddies:
		panic("not implemented")
	case BuddyDelBuddies:
		panic("not implemented")
	case BuddyWatcherListQuery:
		panic("not implemented")
	case BuddyWatcherSubRequest:
		panic("not implemented")
	case BuddyWatcherNotification:
		panic("not implemented")
	case BuddyRejectNotification:
		panic("not implemented")
	case BuddyArrived:
		panic("not implemented")
	case BuddyDeparted:
		panic("not implemented")
	case BuddyAddTempBuddies:
		panic("not implemented")
	case BuddyDelTempBuddies:
		panic("not implemented")
	}
	return nil
}

func SendAndReceiveBuddyRights(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveBuddyRights read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x03_0x02_BuddyRightsQuery{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: 0x03,
		SubGroup:  0x03,
	}
	snacPayloadOut := oscar.SNAC_0x03_0x03_BuddyRightsReply{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x01,
					Val:   uint16(100),
				},
				{
					TType: 0x02,
					Val:   uint16(100),
				},
				{
					TType: 0x03,
					Val:   uint16(100),
				},
				{
					TType: 0x04,
					Val:   uint16(100),
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func NotifyArrival(sess *Session, sm *InMemorySessionManager, fm *FeedbagStore) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: BUDDY,
			SubGroup:  BuddyArrived,
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

func NotifyDeparture(sess *Session, sm *InMemorySessionManager, fm *FeedbagStore) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: BUDDY,
			SubGroup:  BuddyDeparted,
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
