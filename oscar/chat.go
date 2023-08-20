package oscar

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
)

const (
	ChatErr                uint16 = 0x0001
	ChatRoomInfoUpdate            = 0x0002
	ChatUsersJoined               = 0x0003
	ChatUsersLeft                 = 0x0004
	ChatChannelMsgTohost          = 0x0005
	ChatChannelMsgToclient        = 0x0006
	ChatEvilRequest               = 0x0007
	ChatEvilReply                 = 0x0008
	ChatClientErr                 = 0x0009
	ChatPauseRoomReq              = 0x000A
	ChatPauseRoomAck              = 0x000B
	ChatResumeRoom                = 0x000C
	ChatShowMyRow                 = 0x000D
	ChatShowRowByUsername         = 0x000E
	ChatShowRowByNumber           = 0x000F
	ChatShowRowByName             = 0x0010
	ChatRowInfo                   = 0x0011
	ChatListRows                  = 0x0012
	ChatRowListInfo               = 0x0013
	ChatMoreRows                  = 0x0014
	ChatMoveToRow                 = 0x0015
	ChatToggleChat                = 0x0016
	ChatSendQuestion              = 0x0017
	ChatSendComment               = 0x0018
	ChatTallyVote                 = 0x0019
	ChatAcceptBid                 = 0x001A
	ChatSendInvite                = 0x001B
	ChatDeclineInvite             = 0x001C
	ChatAcceptInvite              = 0x001D
	ChatNotifyMessage             = 0x001E
	ChatGotoRow                   = 0x001F
	ChatStageUserJoin             = 0x0020
	ChatStageUserLeft             = 0x0021
	ChatUnnamedSnac22             = 0x0022
	ChatClose                     = 0x0023
	ChatUserBan                   = 0x0024
	ChatUserUnban                 = 0x0025
	ChatJoined                    = 0x0026
	ChatUnnamedSnac27             = 0x0027
	ChatUnnamedSnac28             = 0x0028
	ChatUnnamedSnac29             = 0x0029
	ChatRoomInfoOwner             = 0x0030
)

func routeChat(cr *ChatRegistry, sess *Session, sm *SessionManager, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case ChatErr:
		panic("not implemented")
	case ChatRoomInfoUpdate:
		panic("not implemented")
	case ChatUsersJoined:
		panic("not implemented")
	case ChatUsersLeft:
		panic("not implemented")
	case ChatChannelMsgTohost:
		return SendAndReceiveChatChannelMsgTohost(sess, sm, flap, snac, r, w, sequence)
	case ChatChannelMsgToclient:
		panic("not implemented")
	case ChatEvilRequest:
		panic("not implemented")
	case ChatEvilReply:
		panic("not implemented")
	case ChatClientErr:
		panic("not implemented")
	case ChatPauseRoomReq:
		panic("not implemented")
	case ChatPauseRoomAck:
		panic("not implemented")
	case ChatResumeRoom:
		panic("not implemented")
	case ChatShowMyRow:
		panic("not implemented")
	case ChatShowRowByUsername:
		panic("not implemented")
	case ChatShowRowByNumber:
		panic("not implemented")
	case ChatShowRowByName:
		panic("not implemented")
	case ChatRowInfo:
		panic("not implemented")
	case ChatListRows:
		panic("not implemented")
	case ChatRowListInfo:
		panic("not implemented")
	case ChatMoreRows:
		panic("not implemented")
	case ChatMoveToRow:
		panic("not implemented")
	case ChatToggleChat:
		panic("not implemented")
	case ChatSendQuestion:
		panic("not implemented")
	case ChatSendComment:
		panic("not implemented")
	case ChatTallyVote:
		panic("not implemented")
	case ChatAcceptBid:
		panic("not implemented")
	case ChatSendInvite:
		panic("not implemented")
	case ChatDeclineInvite:
		panic("not implemented")
	case ChatAcceptInvite:
		panic("not implemented")
	case ChatNotifyMessage:
		panic("not implemented")
	case ChatGotoRow:
		panic("not implemented")
	case ChatStageUserJoin:
		panic("not implemented")
	case ChatStageUserLeft:
		panic("not implemented")
	case ChatUnnamedSnac22:
		panic("not implemented")
	case ChatClose:
		panic("not implemented")
	case ChatUserBan:
		panic("not implemented")
	case ChatUserUnban:
		panic("not implemented")
	case ChatJoined:
		panic("not implemented")
	case ChatUnnamedSnac27:
		panic("not implemented")
	case ChatUnnamedSnac28:
		panic("not implemented")
	case ChatUnnamedSnac29:
		panic("not implemented")
	case ChatRoomInfoOwner:
		panic("not implemented")
	}
	return nil
}

type snacChatMessage struct {
	cookie  uint64
	channel uint16
	TLVPayload
}

func (s *snacChatMessage) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.cookie); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &s.channel); err != nil {
		return err
	}
	return s.TLVPayload.read(r, map[uint16]reflect.Kind{
		0x01: reflect.Slice,
		0x06: reflect.Slice,
		0x05: reflect.Slice,
	})
}

func (s *snacChatMessage) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.cookie); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.channel); err != nil {
		return err
	}
	return s.TLVPayload.write(w)
}

type senders []*snacSenderInfo

func (s senders) write(w io.Writer) error {
	for _, sender := range s {
		if err := sender.write(w); err != nil {
			return err
		}
	}
	return nil
}

type snacSenderInfo struct {
	screenName   string
	warningLevel uint16
	TLVPayload
}

func (f *snacSenderInfo) write(w io.Writer) error {
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

func SendAndReceiveChatChannelMsgTohost(sess *Session, sm *SessionManager, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveChatChannelMsgTohost read SNAC frame: %+v\n", snac)

	snacPayloadIn := &snacChatMessage{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: CHAT,
		subGroup:  ChatChannelMsgToclient,
	}

	snacPayloadOut := &snacChatMessage{
		cookie:  snacPayloadIn.cookie,
		channel: snacPayloadIn.channel,
	}

	if tlv, ok := snacPayloadIn.getTLV(0x01); ok {
		snacPayloadOut.addTLV(tlv)
	}
	if tlv, ok := snacPayloadIn.getTLV(0x05); ok {
		snacPayloadOut.addTLV(tlv)
	}
	if tlv, ok := snacPayloadIn.getTLV(0x01); ok {
		snacPayloadOut.addTLV(tlv)
	}

	snacPayloadOut.addTLV(&TLV{
		tType: 0x03,
		val: &snacSenderInfo{
			screenName:   sess.ScreenName,
			warningLevel: sess.GetWarning(),
			TLVPayload: TLVPayload{
				TLVs: sess.GetUserInfo(),
			},
		},
	})

	sm.Broadcast(&XMessage{
		flap:      flap,
		snacFrame: snacFrameOut,
		snacOut:   snacPayloadOut,
	})

	return nil
}

func AlertChatArrival(sess *Session, sm *SessionManager) error {

	s := senders{}

	sessions := sm.All()

	for _, uSess := range sessions {
		s = append(s, &snacSenderInfo{
			screenName:   uSess.ScreenName,
			warningLevel: uSess.GetWarning(),
			TLVPayload: TLVPayload{
				TLVs: uSess.GetUserInfo(),
			},
		})
	}

	sess.SendMessage(&XMessage{
		flap: &flapFrame{
			startMarker: 42,
			frameType:   2,
		},
		snacFrame: snacFrame{
			foodGroup: CHAT,
			subGroup:  ChatUsersJoined,
		},
		snacOut: s,
	})

	return nil
}

func AlertChatRoomInfoUpdate(sess *Session, sm *SessionManager) error {
	sess.SendMessage(&XMessage{
		flap: &flapFrame{
			startMarker: 42,
			frameType:   2,
		},
		snacFrame: snacFrame{
			foodGroup: CHAT,
			subGroup:  ChatRoomInfoUpdate,
		},
		snacOut: &snacCreateRoom{
			exchange:       4,
			cookie:         []byte(cannedUUID.String()),
			instanceNumber: 100,
			detailLevel:    2,
			TLVPayload: TLVPayload{
				TLVs: []*TLV{
					{
						tType: 0x006a,
						val:   cannedName,
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
						val:   cannedName,
					},
					{
						tType: 0x00d5,
						val:   uint8(2),
					},
				},
			},
		},
	})
	return nil
}
