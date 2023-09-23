package oscar

import (
	"fmt"
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

func routeBuddy(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {

	switch snac.subGroup {
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

func SendAndReceiveBuddyRights(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveBuddyRights read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x03_0x02_BuddyRightsQuery{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x03,
		subGroup:  0x03,
	}
	snacPayloadOut := SNAC_0x03_0x03_BuddyRightsReply{
		TLVRestBlock: TLVRestBlock{
			TLVList: TLVList{
				{
					tType: 0x01,
					val:   uint16(100),
				},
				{
					tType: 0x02,
					val:   uint16(100),
				},
				{
					tType: 0x03,
					val:   uint16(100),
				},
				{
					tType: 0x04,
					val:   uint16(100),
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func NotifyArrival(sess *Session, sm *SessionManager, fm *FeedbagStore) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
		snacFrame: snacFrame{
			foodGroup: BUDDY,
			subGroup:  BuddyArrived,
		},
		snacOut: SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
				TLVBlock: TLVBlock{
					TLVList: sess.GetUserInfo(),
				},
			},
		},
	})

	return nil
}

func NotifyDeparture(sess *Session, sm *SessionManager, fm *FeedbagStore) error {
	screenNames, err := fm.InterestedUsers(sess.ScreenName)
	if err != nil {
		return err
	}

	sm.BroadcastToScreenNames(screenNames, XMessage{
		snacFrame: snacFrame{
			foodGroup: BUDDY,
			subGroup:  BuddyDeparted,
		},
		snacOut: SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
			},
		},
	})

	return nil
}
