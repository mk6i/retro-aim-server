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

func routeChatNav(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
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
		return SendAndReceiveCreateRoom(flap, snac, r, w, sequence)
	case ChatNavNavInfo:
		panic("not implemented")
	}
	return nil
}

func SendAndReceiveNextChatRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveNextChatRights read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: CHAT_NAV,
		subGroup:  ChatNavNavInfo,
	}
	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x02,
				val:   uint8(1),
			},
			{
				tType: 0x03,
				val:   []byte{},
			},
		},
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacCreateRoom struct {
	exchange       uint16
	cookie         []byte
	instanceNumber uint16
	detailLevel    uint8
	TLVPayload
}

func (s *snacCreateRoom) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.exchange); err != nil {
		return err
	}
	var l uint8
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	s.cookie = make([]byte, l)
	if _, err := r.Read(s.cookie); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.instanceNumber); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.detailLevel); err != nil {
		return err
	}

	var tlvCount uint16
	if err := binary.Read(r, binary.BigEndian, &tlvCount); err != nil {
		return err
	}

	return s.TLVPayload.read(r, map[uint16]reflect.Kind{
		0x00d3: reflect.String, // name
		0x00d6: reflect.String, // charset
		0x00d7: reflect.String, // lang
	})
}

func (s *snacCreateRoom) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.exchange); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(len(s.cookie))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.cookie); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.instanceNumber); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.detailLevel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.TLVs))); err != nil {
		return err
	}
	return s.TLVPayload.write(w)
}

func SendAndReceiveCreateRoom(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveCreateRoom read SNAC frame: %+v\n", snac)

	snacPayloadIn := &snacCreateRoom{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	name, _ := snacPayloadIn.getString(0x00d3)
	//charset, _ := snacPayloadIn.getString(0x00d6)
	//lang, _ := snacPayloadIn.getString(0x00d7)

	snacPayloadIn.TLVPayload = TLVPayload{
		[]*TLV{
			//{
			//	tType: 0x00d3,
			//	val:   name,
			//},
			//{
			//	tType: 0x00d6,
			//	val:   charset,
			//},
			//{
			//	tType: 0x00d7,
			//	val:   lang,
			//},
			{
				tType: 0x006a,
				val:   name,
			},
			{
				tType: 0x00c9,
				val:   uint16(0),
			},
			{
				tType: 0x00ca,
				val:   uint32(time.Now().Unix()),
			},
			{
				tType: 0x00d1,
				val:   uint16(100),
			},
			{
				tType: 0x00d2,
				val:   uint16(100),
			},
			{
				tType: 0x00d3,
				val:   name,
			},
			{
				tType: 0x00d5,
				val:   uint8(1),
			},
		},
	}

	snacPayloadIn.detailLevel = 0x02

	buf := &bytes.Buffer{}
	if err := snacPayloadIn.write(buf); err != nil {
		return err
	}

	snacOut := &TLVPayload{
		[]*TLV{
			{
				tType: 0x0004,
				val:   buf.Bytes(),
			},
		},
	}

	snacFrameOut := snacFrame{
		foodGroup: CHAT_NAV,
		subGroup:  ChatNavCreateRoom,
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacOut, sequence, w)
}
