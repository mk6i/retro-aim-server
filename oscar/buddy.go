package oscar

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
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

func routeBuddy(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {

	switch snac.subGroup {
	case BuddyErr:
		panic("not implemented")
	case BuddyRightsQuery:
		return SendAndReceiveBuddyRights(flap, snac, r, w, sequence)
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
	TLVPayload
}

func (s *snacBuddyRights) read(r io.Reader) error {
	return s.TLVPayload.read(r, map[uint16]reflect.Kind{
		0x05: reflect.Uint16,
	})
}

func SendAndReceiveBuddyRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveBuddyRights read SNAC frame: %+v\n", snac)

	snacPayloadIn := &snacBuddyRights{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveBuddyRights read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x03,
		subGroup:  0x03,
	}
	snacPayloadOut := &snacBuddyRights{
		TLVPayload: TLVPayload{
			TLVs: []*TLV{
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

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacBuddyArrived struct {
	screenName   string
	warningLevel uint16
	TLVPayload
}

func (f *snacBuddyArrived) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint8(len(f.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.warningLevel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.TLVs))); err != nil {
		return err
	}
	return f.TLVPayload.write(w)
}

func SetBuddyArrived(w io.Writer, sequence *uint32, screenName string) error {
	flap := &flapFrame{
		startMarker: 42,
		frameType:   2,
	}

	snacFrameOut := snacFrame{
		foodGroup: BUDDY,
		subGroup:  BuddyArrived,
		requestID: 12425,
	}
	snacPayloadOut := &snacBuddyArrived{
		screenName:   screenName,
		warningLevel: 0,
		TLVPayload: TLVPayload{
			TLVs: []*TLV{
				{
					tType: 0x01,
					val:   uint16(0x0004),
				},
				{
					tType: 0x06,
					val:   uint16(0x0000),
				},
			},
		},
	}

	return writeOutSNAC(nil, flap, snacFrameOut, snacPayloadOut, sequence, w)
}
