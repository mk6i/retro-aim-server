package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"io"
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

func routeChatNav(sess *Session, cr *ChatRegistry, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case ChatNavRequestChatRights:
		return SendAndReceiveNextChatRights(snac, w, sequence)
	case ChatNavRequestRoomInfo:
		return SendAndReceiveRequestRoomInfo(cr, snac, r, w, sequence)
	case ChatNavCreateRoom:
		return SendAndReceiveCreateRoom(sess, cr, NewChatRoom, snac, r, w, sequence)
	default:
		return ErrUnimplementedSNAC
	}
}

type ChatCookie struct {
	Cookie []byte `len_prefix:"uint16"`
	SessID string `len_prefix:"uint16"`
}

func SendAndReceiveNextChatRights(snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveNextChatRights read SNAC frame: %+v\n", snac)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT_NAV,
		SubGroup:  ChatNavNavInfo,
	}

	snacPayloadOut := oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x02,
					Val:   uint8(10),
				},
				{
					TType: 0x03,
					Val: oscar.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 4,
						TLVBlock: oscar.TLVBlock{
							TLVList: oscar.TLVList{
								{
									TType: 0x0002,
									Val:   uint16(0x0010),
								},
								{
									TType: 0x00c9,
									Val:   uint16(15),
								},
								{
									TType: 0x00d3,
									Val:   "default Exchange",
								},
								{
									TType: 0x00d5,
									Val:   uint8(2),
								},
								{
									TType: 0xd6,
									Val:   "us-ascii",
								},
								{
									TType: 0xd7,
									Val:   "en",
								},
								{
									TType: 0xd8,
									Val:   "us-ascii",
								},
								{
									TType: 0xd9,
									Val:   "en",
								},
							},
						},
					},
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func NewChatRoom() ChatRoom {
	return ChatRoom{
		Cookie:         uuid.New().String(),
		CreateTime:     time.Now(),
		SessionManager: NewSessionManager(),
	}
}

type ChatRoomFactory func() ChatRoom

func SendAndReceiveCreateRoom(sess *Session, cr *ChatRegistry, newChatRoom ChatRoomFactory, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	name, hasName := snacPayloadIn.GetString(oscar.ChatTLVRoomName)
	if !hasName {
		return errors.New("unable to find chat name")
	}

	room := newChatRoom()
	room.DetailLevel = snacPayloadIn.DetailLevel
	room.Exchange = snacPayloadIn.Exchange
	room.InstanceNumber = snacPayloadIn.InstanceNumber
	room.Name = name
	cr.Register(room)

	// add user to chat room
	room.NewSessionWithSN(sess.ID, sess.ScreenName)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT_NAV,
		SubGroup:  ChatNavNavInfo,
	}
	snacPayloadOut := oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: oscar.ChatNavTLVRoomInfo,
					Val: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       snacPayloadIn.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: snacPayloadIn.InstanceNumber,
						DetailLevel:    snacPayloadIn.DetailLevel,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					},
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SendAndReceiveRequestRoomInfo(cr *ChatRegistry, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveRequestRoomInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	room, err := cr.Retrieve(string(snacPayloadIn.Cookie))
	if err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: CHAT_NAV,
		SubGroup:  ChatNavNavInfo,
	}

	snacPayloadOut := oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x04,
					Val: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       4,
						Cookie:         room.Cookie,
						InstanceNumber: 100,
						DetailLevel:    2,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					},
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
