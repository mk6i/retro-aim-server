package server

import (
	"errors"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"time"
)

type ChatNavHandler interface {
	CreateRoomHandler(sess *Session, cr *ChatRegistry, newChatRoom ChatRoomFactory, snacPayloadIn oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (XMessage, error)
	RequestChatRightsHandler() XMessage
	RequestRoomInfoHandler(cr *ChatRegistry, snacPayloadIn oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (XMessage, error)
}

func NewChatNavRouter() ChatNavRouter {
	return ChatNavRouter{
		ChatNavHandler: ChatNavService{},
	}
}

type ChatNavRouter struct {
	ChatNavHandler
}

func (rt *ChatNavRouter) RouteChatNav(sess *Session, cr *ChatRegistry, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.ChatNavRequestChatRights:
		outSNAC := rt.RequestChatRightsHandler()
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.ChatNavRequestRoomInfo:
		inSNAC := oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.RequestRoomInfoHandler(cr, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.ChatNavCreateRoom:
		snacPayloadIn := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{}
		if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
			return err
		}
		outSNAC, err := rt.CreateRoomHandler(sess, cr, NewChatRoom, snacPayloadIn)
		if err != nil {
			return err
		}
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type ChatNavService struct {
}

type ChatCookie struct {
	Cookie []byte `len_prefix:"uint16"`
	SessID string `len_prefix:"uint16"`
}

func (s ChatNavService) RequestChatRightsHandler() XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		snacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x02, uint8(10)),
					oscar.NewTLV(0x03, oscar.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 4,
						TLVBlock: oscar.TLVBlock{
							TLVList: oscar.TLVList{
								oscar.NewTLV(0x0002, uint16(0x0010)),
								oscar.NewTLV(0x00c9, uint16(15)),
								oscar.NewTLV(0x00d3, "default Exchange"),
								oscar.NewTLV(0x00d5, uint8(2)),
								oscar.NewTLV(0xd6, "us-ascii"),
								oscar.NewTLV(0xd7, "en"),
								oscar.NewTLV(0xd8, "us-ascii"),
								oscar.NewTLV(0xd9, "en"),
							},
						},
					}),
				},
			},
		},
	}
}

func NewChatRoom() ChatRoom {
	return ChatRoom{
		Cookie:         uuid.New().String(),
		CreateTime:     time.Now(),
		SessionManager: NewSessionManager(),
	}
}

type ChatRoomFactory func() ChatRoom

func (s ChatNavService) CreateRoomHandler(sess *Session, cr *ChatRegistry, newChatRoom ChatRoomFactory, snacPayloadIn oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (XMessage, error) {
	name, hasName := snacPayloadIn.GetString(oscar.ChatTLVRoomName)
	if !hasName {
		return XMessage{}, errors.New("unable to find chat name")
	}

	room := newChatRoom()
	room.DetailLevel = snacPayloadIn.DetailLevel
	room.Exchange = snacPayloadIn.Exchange
	room.InstanceNumber = snacPayloadIn.InstanceNumber
	room.Name = name
	cr.Register(room)

	// add user to chat room
	room.NewSessionWithSN(sess.ID, sess.ScreenName)

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		snacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.ChatNavTLVRoomInfo, oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       snacPayloadIn.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: snacPayloadIn.InstanceNumber,
						DetailLevel:    snacPayloadIn.DetailLevel,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}

func (s ChatNavService) RequestRoomInfoHandler(cr *ChatRegistry, snacPayloadIn oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (XMessage, error) {
	room, err := cr.Retrieve(string(snacPayloadIn.Cookie))
	if err != nil {
		return XMessage{}, err
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		snacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x04, oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       4,
						Cookie:         room.Cookie,
						InstanceNumber: 100,
						DetailLevel:    2,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}
