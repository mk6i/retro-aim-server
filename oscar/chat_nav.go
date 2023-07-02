package oscar

import (
	"fmt"
	"io"
)

const (
	ChatNavErr                 uint16 = 0x0001
	ChatNavRequestChatRights          = 0x0002
	ChatNavRequestExchangeInfo        = 0x0003
	ChatNavRequestRoomInfo            = 0x0004
	ChatNavRequestMoreRoomInfo        = 0x0005
	ChatNavRequestOccupantList        = 0x0006
	ChatNavSearchForRoom              = 0x0007
	ChatNavCreateRoom                 = 0x0008
	ChatNavNavInfo                    = 0x0009
)

func routeChatNav(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	switch snac.subGroup {
	case ChatNavErr:
		panic("not implemented")
	case ChatNavRequestChatRights:
		return SendAndReceiveNextChatRights(flap, snac, r, w, sequence)
	case ChatNavRequestExchangeInfo:
		panic("not implemented")
	case ChatNavRequestRoomInfo:
		panic("not implemented")
	case ChatNavRequestMoreRoomInfo:
		panic("not implemented")
	case ChatNavRequestOccupantList:
		panic("not implemented")
	case ChatNavSearchForRoom:
		panic("not implemented")
	case ChatNavCreateRoom:
		panic("not implemented")
	case ChatNavNavInfo:
		panic("not implemented")
	}
	return nil
}

func SendAndReceiveNextChatRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("sendAndReceiveNextChatRights read SNAC frame: %+v\n", snac)

	snacPayload := &snacFrame{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveNextChatRights read SNAC payload: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: 0x0D,
		subGroup:  0x09,
	}
	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{},
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}
