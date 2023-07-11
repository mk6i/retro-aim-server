package oscar

import (
	"bytes"
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

const (
	FeedbagClassIdBuddy            uint16 = 0x0000
	FeedbagClassIdGroup                   = 0x0001
	FeedbagClassIdPermit                  = 0x0002
	FeedbagClassIdDeny                    = 0x0003
	FeedbagClassIdPdinfo                  = 0x0004
	FeedbagClassIdBuddyPrefs              = 0x0005
	FeedbagClassIdNonbuddy                = 0x0006
	FeedbagClassIdTpaProvider             = 0x0007
	FeedbagClassIdTpaSubscription         = 0x0008
	FeedbagClassIdClientPrefs             = 0x0009
	FeedbagClassIdStock                   = 0x000A
	FeedbagClassIdWeather                 = 0x000B
	FeedbagClassIdWatchList               = 0x000D
	FeedbagClassIdIgnoreList              = 0x000E
	FeedbagClassIdDateTime                = 0x000F
	FeedbagClassIdExternalUser            = 0x0010
	FeedbagClassIdRootCreator             = 0x0011
	FeedbagClassIdFish                    = 0x0012
	FeedbagClassIdImportTimestamp         = 0x0013
	FeedbagClassIdBart                    = 0x0014
	FeedbagClassIdRbOrder                 = 0x0015
	FeedbagClassIdPersonality             = 0x0016
	FeedbagClassIdAlProf                  = 0x0017
	FeedbagClassIdAlInfo                  = 0x0018
	FeedbagClassIdInteraction             = 0x0019
	FeedbagClassIdVanityInfo              = 0x001D
	FeedbagClassIdFavoriteLocation        = 0x001E
	FeedbagClassIdBartPdinfo              = 0x001F
	FeedbagClassIdCustomEmoticons         = 0x0024
	FeedbagClassIdMaxPredefined           = 0x0024
	FeedbagClassIdXIcqStatusNote          = 0x015C
	FeedbagClassIdMin                     = 0x0400
)

func routeFeedbag(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	switch snac.subGroup {
	case FeedbagErr:
		panic("not implemented")
	case FeedbagRightsQuery:
		return SendAndReceiveFeedbagRightsQuery(flap, snac, r, w, sequence)
	case FeedbagQuery:
		return ReceiveAndSendFeedbagQuery(flap, snac, r, w, sequence)
	case FeedbagQueryIfModified:
		return ReceiveAndSendFeedbagQueryIfModified(flap, snac, r, w, sequence)
	case FeedbagUse:
		return ReceiveUse(flap, snac, r, w, sequence)
	case FeedbagInsertItem:
		return ReceiveInsertItem(flap, snac, r, w, sequence)
	case FeedbagUpdateItem:
		return ReceiveUpdateItem(flap, snac, r, w, sequence)
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
		return ReceiveFeedbagStartCluster(flap, snac, r, w, sequence)
	case FeedbagEndCluster:
		return ReceiveFeedbagEndCluster(flap, snac, r, w, sequence)
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

func SendAndReceiveFeedbagRightsQuery(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
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
	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x03,
				val:   uint16(200),
			},
			{
				tType: 0x04,
				val: []uint16{
					0x3D,
					0x3D,
					0x64,
					0x64,
					0x01,
					0x01,
					0x32,
					0x00,
					0x00,
					0x03,
					0x00,
					0x00,
					0x00,
					0x80,
					0xFF,
					0x14,
					0xC8,
					0x01,
					0x00,
					0x01,
					0x00,
				},
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
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type feedbagItem struct {
	name    string
	groupID uint16
	itemID  uint16
	classID uint16
	TLVPayload
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
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.TLVPayload.TLVs))); err != nil {
		return err
	}
	if err := f.TLVPayload.write(w); err != nil {
		return err
	}
	return nil
}

func (f *feedbagItem) read(r io.Reader) error {
	var l uint16
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	f.name = string(buf)
	if err := binary.Read(r, binary.BigEndian, &f.groupID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.itemID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.classID); err != nil {
		return err
	}
	return f.TLVPayload.read(r, map[uint16]reflect.Kind{
		FeedbagClassIdBuddy:            reflect.Slice,
		FeedbagClassIdGroup:            reflect.Slice,
		FeedbagClassIdPermit:           reflect.Slice,
		FeedbagClassIdDeny:             reflect.Slice,
		FeedbagClassIdPdinfo:           reflect.Slice,
		FeedbagClassIdBuddyPrefs:       reflect.Slice,
		FeedbagClassIdNonbuddy:         reflect.Slice,
		FeedbagClassIdTpaProvider:      reflect.Slice,
		FeedbagClassIdTpaSubscription:  reflect.Slice,
		FeedbagClassIdClientPrefs:      reflect.Slice,
		FeedbagClassIdStock:            reflect.Slice,
		FeedbagClassIdWeather:          reflect.Slice,
		FeedbagClassIdWatchList:        reflect.Slice,
		FeedbagClassIdIgnoreList:       reflect.Slice,
		FeedbagClassIdDateTime:         reflect.Slice,
		FeedbagClassIdExternalUser:     reflect.Slice,
		FeedbagClassIdRootCreator:      reflect.Slice,
		FeedbagClassIdFish:             reflect.Slice,
		FeedbagClassIdImportTimestamp:  reflect.Slice,
		FeedbagClassIdBart:             reflect.Slice,
		FeedbagClassIdRbOrder:          reflect.Slice,
		FeedbagClassIdPersonality:      reflect.Slice,
		FeedbagClassIdAlProf:           reflect.Slice,
		FeedbagClassIdAlInfo:           reflect.Slice,
		FeedbagClassIdInteraction:      reflect.Slice,
		FeedbagClassIdVanityInfo:       reflect.Slice,
		FeedbagClassIdFavoriteLocation: reflect.Slice,
		FeedbagClassIdBartPdinfo:       reflect.Slice,
		FeedbagClassIdXIcqStatusNote:   reflect.Slice,
		FeedbagClassIdMin:              reflect.Slice,
	})
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

func ReceiveAndSendFeedbagQuery(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("receiveAndSendFeedbagQuery read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x06,
	}
	snacPayloadOut := &snacFeedbagQuery{
		version:    0,
		items:      []*feedbagItem{},
		lastUpdate: uint32(time.Now().Unix()),
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacQueryIfModified struct {
	lastUpdate uint32
	count      uint8
}

func (s *snacQueryIfModified) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.lastUpdate); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, s.count)
}

func (s *snacQueryIfModified) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.lastUpdate); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &s.count)
}

func ReceiveAndSendFeedbagQueryIfModified(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("ReceiveAndSendFeedbagQueryIfModified read SNAC frame: %+v\n", snac)

	snacPayload := &snacQueryIfModified{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveAndSendFeedbagQueryIfModified read SNAC: %+v\n", snacPayload)

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x0F,
	}
	snacPayloadOut := &snacQueryIfModified{
		lastUpdate: snacPayload.lastUpdate,
		count:      snacPayload.count,
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacFeedbagStatusReply struct {
	results []uint16
}

func (s *snacFeedbagStatusReply) write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s.results)
}

func ReceiveInsertItem(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("ReceiveInsertItem read SNAC frame: %+v\n", snac)

	snacPayloadOut := &snacFeedbagStatusReply{}

	for {
		item := &feedbagItem{}
		if err := item.read(r); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		snacPayloadOut.results = append(snacPayloadOut.results, 0x0000) // success by default
		fmt.Printf("ReceiveInsertItem read SNAC feedbag item: %+v\n", item)
	}

	snacFrameOut := snacFrame{
		foodGroup: FEEDBAG,
		subGroup:  FeedbagStatus,
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveUpdateItem(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("ReceiveUpdateItem read SNAC frame: %+v\n", snac)

	b := make([]byte, flap.payloadLength-10)
	if _, err := r.Read(b); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)

	var items []*feedbagItem

	for buf.Len() > 0 {
		item := &feedbagItem{}
		if err := item.read(buf); err != nil {
			return err
		}
		fmt.Printf("\titem: %+v\n", item)
		items = append(items, item)
	}

	snacPayloadOut := &snacFeedbagStatusReply{}

	for _, item := range items {
		snacPayloadOut.results = append(snacPayloadOut.results, 0x0000) // success by default
		fmt.Printf("ReceiveUpdateItem read SNAC feedbag item: %+v\n", item)
	}

	snacFrameOut := snacFrame{
		foodGroup: FEEDBAG,
		subGroup:  FeedbagStatus,
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveFeedbagStartCluster(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("ReceiveFeedbagStartCluster read SNAC frame: %+v\n", snac)

	tlv := &TLVPayload{}
	if err := tlv.read(r, map[uint16]reflect.Kind{}); err != nil {
		return err
	}

	return nil
}

func ReceiveFeedbagEndCluster(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("receiveFeedbagEndCluster read SNAC frame: %+v\n", snac)
	return nil
}

func ReceiveUse(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint16) error {
	fmt.Printf("ReceiveUse read SNAC frame: %+v\n", snac)
	return nil
}
