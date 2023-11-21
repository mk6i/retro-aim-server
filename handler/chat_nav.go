package handler

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"log/slog"
	"time"
)

func NewChatRoom() state.ChatRoom {
	return state.ChatRoom{
		Cookie:     uuid.New().String(),
		CreateTime: time.Now(),
	}
}

func NewChatNavService(logger *slog.Logger, cr *state.ChatRegistry, newChatRoom func() state.ChatRoom, newChatSessMgr func() ChatSessionManager) *ChatNavService {
	return &ChatNavService{
		logger:         logger,
		chatRegistry:   cr,
		newChatRoom:    NewChatRoom,
		newChatSessMgr: newChatSessMgr,
	}
}

type ChatNavService struct {
	logger         *slog.Logger
	chatRegistry   *state.ChatRegistry
	newChatRoom    func() state.ChatRoom
	newChatSessMgr func() ChatSessionManager
}

func (s ChatNavService) RequestChatRightsHandler(context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		SnacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
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

func (s ChatNavService) CreateRoomHandler(ctx context.Context, sess *state.Session, snacPayloadIn oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (oscar.XMessage, error) {
	name, hasName := snacPayloadIn.GetString(oscar.ChatTLVRoomName)
	if !hasName {
		return oscar.XMessage{}, errors.New("unable to find chat name")
	}

	room := s.newChatRoom()
	room.DetailLevel = snacPayloadIn.DetailLevel
	room.Exchange = snacPayloadIn.Exchange
	room.InstanceNumber = snacPayloadIn.InstanceNumber
	room.Name = name

	chatSessMgr := s.newChatSessMgr()

	s.chatRegistry.Register(room, chatSessMgr)

	// add user to chat room
	chatSessMgr.NewSessionWithSN(sess.ID(), sess.ScreenName())

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		SnacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
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

func (s ChatNavService) RequestRoomInfoHandler(_ context.Context, snacPayloadIn oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (oscar.XMessage, error) {
	room, _, err := s.chatRegistry.Retrieve(string(snacPayloadIn.Cookie))
	if err != nil {
		return oscar.XMessage{}, err
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		SnacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
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
