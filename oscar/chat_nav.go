package oscar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
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

func routeChatNav(sess *Session, cr *ChatRegistry, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case ChatNavErr:
		panic("not implemented")
	case ChatNavRequestChatRights:
		return SendAndReceiveNextChatRights(flap, snac, r, w, sequence)
	case ChatNavRequestExchangeInfo:
		panic("not implemented")
	case ChatNavRequestRoomInfo:
		return SendAndReceiveChatNav(flap, snac, r, w, sequence)
	case ChatNavRequestMoreRoomInfo:
		panic("not implemented")
	case ChatNavRequestOccupantList:
		panic("not implemented")
	case ChatNavSearchForRoom:
		panic("not implemented")
	case ChatNavCreateRoom:
		return SendAndReceiveCreateRoom(sess, cr, flap, snac, r, w, sequence)
	case ChatNavNavInfo:
		panic("not implemented")
	}
	return nil
}

type ChatCookie struct {
	Cookie []byte
	SessID string
}

func (s *ChatCookie) read(r io.Reader) error {
	var l uint16
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	s.Cookie = make([]byte, l)
	if _, err := r.Read(s.Cookie); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	s.SessID = string(buf)
	return nil
}

func (s *ChatCookie) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.Cookie))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.Cookie); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.SessID))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s.SessID)); err != nil {
		return err
	}
	return nil
}

type exchangeInfo struct {
	identifier uint16
	TLVPayload
}

func (s *exchangeInfo) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.identifier); err != nil {
		return err
	}
	//if err := binary.Write(w, binary.BigEndian, uint8(len(s.TLVs))); err != nil {
	//	return err
	//}
	return s.TLVPayload.write(w)
}

func SendAndReceiveNextChatRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveNextChatRights read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: CHAT_NAV,
		subGroup:  ChatNavNavInfo,
	}

	xchange := TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x000a,
				val:   uint16(0x0114),
			},
			//{
			//	tType: 0x000d,
			//	val:   nil,
			//},
			{
				tType: 0x0004,
				val:   uint8(15),
			},
			{
				tType: 0x0002,
				val:   uint16(0x0010),
			},
			{
				tType: 0x00c9,
				val:   uint16(15),
			},
			{
				tType: 0x00ca,
				val:   uint32(time.Now().Unix()),
			},
			//{
			//	tType: 0x00d0,
			//	val:   nil,
			//},
			{
				tType: 0x00d1,
				val:   uint16(1024),
			},
			{
				tType: 0x00d2,
				val:   uint16(100),
			},
			{
				tType: 0x00d3,
				val:   "hello",
			},
			{
				tType: 0x00d4,
				val:   "http://www.google.com",
			},
			{
				tType: 0x00d5,
				val:   uint8(2),
			},
			{
				tType: 0xd6,
				val:   "us-ascii",
			},
			{
				tType: 0xd7,
				val:   "en",
			},
			{
				tType: 0xd8,
				val:   "us-ascii",
			},
			{
				tType: 0xd9,
				val:   "en",
			},
			{
				tType: 0x00da,
				val:   uint16(0),
			},
		},
	}

	roomBuf := &bytes.Buffer{}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(4)); err != nil {
		return err
	}
	if err := xchange.write(roomBuf); err != nil {
		return err
	}

	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x02,
				val:   uint8(10),
			},
			{
				tType: 0x03,
				val:   roomBuf.Bytes(),
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

func SendAndReceiveCreateRoom(sess *Session, cr *ChatRegistry, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveCreateRoom read SNAC frame: %+v\n", snac)

	snacPayloadIn := &snacCreateRoom{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	name, _ := snacPayloadIn.getString(0x00d3)
	//charset, _ := snacPayloadIn.getString(0x00d6)
	//lang, _ := snacPayloadIn.getString(0x00d7)

	snacFrameOut := snacFrame{
		foodGroup: CHAT_NAV,
		subGroup:  ChatNavNavInfo,
	}

	xchange := TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x006a,
				val:   name,
			},
			{
				tType: 0x00c9,
				val:   uint16(1), // tweak this
			},
			{
				tType: 0x00ca,
				val:   uint32(time.Now().Unix()),
			},
			{
				tType: 0x00d1,
				val:   uint16(1024),
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
				val:   uint8(2),
			},
		},
	}

	uu, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	sm := NewSessionManager()
	sm.NewSessionWithSN(sess.ID, sess.ScreenName)

	chatID := uu.String()
	cr.Register(chatID, sm)

	roomBuf := &bytes.Buffer{}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(4)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint8(len(chatID))); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, []byte(chatID)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(100)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint8(2)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(len(xchange.TLVs))); err != nil {
		return err
	}
	if err := xchange.write(roomBuf); err != nil {
		return err
	}

	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x04,
				val:   roomBuf.Bytes(),
			},
		},
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

type roomInfoOService struct {
	exchange       uint16
	cookie         []byte
	instanceNumber uint16
}

func (s *roomInfoOService) read(r io.Reader) error {
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
	return binary.Read(r, binary.BigEndian, &s.instanceNumber)
}

type roomInfo struct {
	exchange       uint16
	cookie         []byte
	instanceNumber uint16
	detailLevel    uint8
}

func (s *roomInfo) read(r io.Reader) error {
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
	return binary.Read(r, binary.BigEndian, &s.detailLevel)
}

func SendAndReceiveChatNav(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveChatNav read SNAC frame: %+v\n", snac)

	snacPayloadIn := &roomInfo{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	//name, _ := snacPayloadIn.getString(0x00d3)
	//charset, _ := snacPayloadIn.getString(0x00d6)
	//lang, _ := snacPayloadIn.getString(0x00d7)

	snacFrameOut := snacFrame{
		foodGroup: CHAT_NAV,
		subGroup:  ChatNavNavInfo,
	}

	xchange := TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x006a,
				val:   "hahahnewroom!",
			},
			{
				tType: 0x00c9,
				val:   uint16(1), // tweak this
			},
			{
				tType: 0x00ca,
				val:   uint32(time.Now().Unix()),
			},
			{
				tType: 0x00d1,
				val:   uint16(1024),
			},
			{
				tType: 0x00d2,
				val:   uint16(100),
			},
			{
				tType: 0x00d3,
				val:   "hello",
			},
			{
				tType: 0x00d5,
				val:   uint8(2),
			},
		},
	}

	uu, err := uuid.NewUUID()
	if err != nil {
		return nil
	}
	roomBuf := &bytes.Buffer{}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(4)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint8(len(uu.String()))); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, []byte(uu.String())); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(100)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint8(2)); err != nil {
		return err
	}
	if err := binary.Write(roomBuf, binary.BigEndian, uint16(len(xchange.TLVs))); err != nil {
		return err
	}
	if err := xchange.write(roomBuf); err != nil {
		return err
	}

	snacPayloadOut := &TLVPayload{
		TLVs: []*TLV{
			{
				tType: 0x04,
				val:   roomBuf.Bytes(),
			},
		},
	}

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}
