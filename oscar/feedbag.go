package oscar

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
)

const (
	FeedbagErr                      uint16 = 0x0001
	FeedbagRightsQuery                     = 0x0002
	FeedbagQuery                           = 0x0004
	FeedbagQueryIfModified                 = 0x0005
	FeedbagUse                             = 0x0007
	FeedbagInsertItem                      = 0x0008
	FeedbagUpdateItem                      = 0x0009
	FeedbagDeleteItem                      = 0x000A
	FeedbagInsertClass                     = 0x000B
	FeedbagUpdateClass                     = 0x000C
	FeedbagDeleteClass                     = 0x000D
	FeedbagStatus                          = 0x000E
	FeedbagReplyNotModified                = 0x000F
	FeedbagDeleteUser                      = 0x0010
	FeedbagStartCluster                    = 0x0011
	FeedbagEndCluster                      = 0x0012
	FeedbagAuthorizeBuddy                  = 0x0013
	FeedbagPreAuthorizeBuddy               = 0x0014
	FeedbagPreAuthorizedBuddy              = 0x0015
	FeedbagRemoveMe                        = 0x0016
	FeedbagRemoveMe2                       = 0x0017
	FeedbagRequestAuthorizeToHost          = 0x0018
	FeedbagRequestAuthorizeToClient        = 0x0019
	FeedbagRespondAuthorizeToHost          = 0x001A
	FeedbagRespondAuthorizeToClient        = 0x001B
	FeedbagBuddyAdded                      = 0x001C
	FeedbagRequestAuthorizeToBadog         = 0x001D
	FeedbagRespondAuthorizeToBadog         = 0x001E
	FeedbagBuddyAddedToBadog               = 0x001F
	FeedbagTestSnac                        = 0x0021
	FeedbagForwardMsg                      = 0x0022
	FeedbagIsAuthRequiredQuery             = 0x0023
	FeedbagIsAuthRequiredReply             = 0x0024
	FeedbagRecentBuddyUpdate               = 0x0025
)

func routeFeedbag(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	switch snac.subGroup {
	case FeedbagErr:
		panic("not implemented")
	case FeedbagRightsQuery:
		return SendAndReceiveFeedbagRightsQuery(flap, snac, r, w, sequence)
	case FeedbagQuery:
		return ReceiveAndSendFeedbagQuery(flap, snac, r, w, sequence)
	case FeedbagQueryIfModified:
		panic("not implemented")
	case FeedbagUse:
		panic("not implemented")
	case FeedbagInsertItem:
		panic("not implemented")
	case FeedbagUpdateItem:
		panic("not implemented")
	case FeedbagDeleteItem:
		panic("not implemented")
	case FeedbagInsertClass:
		panic("not implemented")
	case FeedbagUpdateClass:
		panic("not implemented")
	case FeedbagDeleteClass:
		panic("not implemented")
	case FeedbagStatus:
		panic("not implemented")
	case FeedbagReplyNotModified:
		panic("not implemented")
	case FeedbagDeleteUser:
		panic("not implemented")
	case FeedbagStartCluster:
		panic("not implemented")
	case FeedbagEndCluster:
		panic("not implemented")
	case FeedbagAuthorizeBuddy:
		panic("not implemented")
	case FeedbagPreAuthorizeBuddy:
		panic("not implemented")
	case FeedbagPreAuthorizedBuddy:
		panic("not implemented")
	case FeedbagRemoveMe:
		panic("not implemented")
	case FeedbagRemoveMe2:
		panic("not implemented")
	case FeedbagRequestAuthorizeToHost:
		panic("not implemented")
	case FeedbagRequestAuthorizeToClient:
		panic("not implemented")
	case FeedbagRespondAuthorizeToHost:
		panic("not implemented")
	case FeedbagRespondAuthorizeToClient:
		panic("not implemented")
	case FeedbagBuddyAdded:
		panic("not implemented")
	case FeedbagRequestAuthorizeToBadog:
		panic("not implemented")
	case FeedbagRespondAuthorizeToBadog:
		panic("not implemented")
	case FeedbagBuddyAddedToBadog:
		panic("not implemented")
	case FeedbagTestSnac:
		panic("not implemented")
	case FeedbagForwardMsg:
		panic("not implemented")
	case FeedbagIsAuthRequiredQuery:
		panic("not implemented")
	case FeedbagIsAuthRequiredReply:
		panic("not implemented")
	case FeedbagRecentBuddyUpdate:
		panic("not implemented")
	}

	return nil
}

type payloadFeedbagRightsQuery struct {
	TLVPayload
}

func (s *payloadFeedbagRightsQuery) read(r io.Reader) error {
	return s.TLVPayload.read(r, map[uint16]reflect.Kind{
		0x0B: reflect.Uint16,
	})
}

func SendAndReceiveFeedbagRightsQuery(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC frame: %+v\n", snac)

	snacPayloadIn := &payloadFeedbagRightsQuery{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x03,
	}
	snacPayloadOut := &snacBUCPLoginRequest{
		TLVPayload: TLVPayload{
			TLVs: []*TLV{
				{
					tType: 0x03,
					val:   uint16(200),
				},
				{
					tType: 0x04,
					val:   uint16(200),
				},
				{
					tType: 0x05,
					val:   uint16(200),
				},
				{
					tType: 0x06,
					val:   uint16(200),
				},
				{
					tType: 0x07,
					val:   uint16(200),
				},
				{
					tType: 0x08,
					val:   uint16(200),
				},
				{
					tType: 0x09,
					val:   uint16(200),
				},
				{
					tType: 0x0A,
					val:   uint16(200),
				},
				{
					tType: 0x0C,
					val:   uint16(200),
				},
				{
					tType: 0x0D,
					val:   uint16(200),
				},
				{
					tType: 0x0E,
					val:   uint16(100),
				},
			},
		},
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type feedbagItem struct {
	name    string
	groupID uint16
	itemID  uint16
	classID uint16
	tlvs    []*TLV
}

func (f *feedbagItem) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.name))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.name)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.groupID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.itemID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.classID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.tlvs))); err != nil {
		return err
	}
	for _, tlv := range f.tlvs {
		if err := tlv.write(w); err != nil {
			return err
		}
	}
	return nil
}

type snacFeedbagQuery struct {
	version    uint8
	items      []*feedbagItem
	lastUpdate uint32
}

func (s *snacFeedbagQuery) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.items))); err != nil {
		return err
	}
	for _, t := range s.items {
		if err := t.write(w); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.BigEndian, s.lastUpdate)
}

func ReceiveAndSendFeedbagQuery(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	fmt.Printf("receiveAndSendFeedbagQuery read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x06,
	}
	snacPayloadOut := &snacFeedbagQuery{
		version: 0,
		items: []*feedbagItem{
			{
				groupID: 0,
				itemID:  0,
				classID: 0,
				name:    "",
				tlvs: []*TLV{
					{
						tType: 0x00C8,
						val:   []uint16{321, 10},
					},
				},
			},
			{
				groupID: 0,
				itemID:  1805,
				classID: 3,
				name:    "spimmer123",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  4046,
				classID: 0x14,
				name:    "5",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  12108,
				classID: 4,
				name:    "",
				tlvs: []*TLV{
					{
						tType: 202,
						val:   uint8(0x04),
					},
					{
						tType: 203,
						val:   uint32(0xffffffff),
					},
					{
						tType: 204,
						val:   uint32(1),
					},
				},
			},
			{
				groupID: 0x0A,
				itemID:  0,
				classID: 1,
				name:    "Friends",
				tlvs: []*TLV{
					{
						tType: 200,
						val:   []uint16{110, 147},
					},
				},
			},
			{
				groupID: 0x0A,
				itemID:  110,
				classID: 0,
				name:    "ChattingChuck",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0x0A,
				itemID:  147,
				classID: 0,
				name:    "example@example.com",
				tlvs:    []*TLV{},
			},
			{
				groupID: 0,
				itemID:  0,
				classID: 1,
				name:    "Empty Group",
				tlvs: []*TLV{
					{
						tType: 200,
						val:   []uint16{},
					},
				},
			},
		},
		lastUpdate: uint32(time.Now().Unix()),
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}

//func ReceiveInsertItem(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
//	fmt.Printf("ReceiveInsertItem read SNAC frame: %+v\n", snac)
//
//	snacPayload := &snacFrame{}
//	if err := snacPayload.read(r); err != nil {
//		return err
//	}
//
//	// read out remainder
//
//	fmt.Printf("ReceiveInsertItem read SNAC: %+v\n", snacPayload)
//
//	return nil
//}
