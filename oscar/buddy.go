package oscar

import (
	"encoding/binary"
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

type snacBuddyRights struct {
	TLVRestBlock
}

func (s *snacBuddyRights) read(r io.Reader) error {
	return s.TLVRestBlock.read(r)
}

func SendAndReceiveBuddyRights(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveBuddyRights read SNAC frame: %+v\n", snac)

	snacPayloadIn := snacBuddyRights{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x03,
		subGroup:  0x03,
	}
	snacPayloadOut := snacBuddyRights{
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

type snacBuddyArrived struct {
	screenName   string
	warningLevel uint16
	TLVBlock
}

func (f snacBuddyArrived) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint8(len(f.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.warningLevel); err != nil {
		return err
	}
	return f.TLVBlock.write(w)
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
		snacOut: snacBuddyArrived{
			screenName:   sess.ScreenName,
			warningLevel: sess.GetWarning(),
			TLVBlock: TLVBlock{
				TLVList: sess.GetUserInfo(),
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
		snacOut: snacBuddyArrived{
			screenName:   sess.ScreenName,
			warningLevel: sess.GetWarning(),
		},
	})

	return nil
}
