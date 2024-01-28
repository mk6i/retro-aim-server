package handler

import (
	"context"
	"errors"

	"log/slog"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
)

// NewChatNavService creates a new instance of NewChatNavService.
func NewChatNavService(logger *slog.Logger, chatRegistry *state.ChatRegistry, newChatRoom func() state.ChatRoom, newChatSessMgr func() SessionManager) *ChatNavService {
	return &ChatNavService{
		logger:         logger,
		chatRegistry:   chatRegistry,
		newChatRoom:    newChatRoom,
		newChatSessMgr: newChatSessMgr,
	}
}

// ChatNavService provides handlers for the ChatNav food group.
type ChatNavService struct {
	logger         *slog.Logger
	chatRegistry   *state.ChatRegistry
	newChatRoom    func() state.ChatRoom
	newChatSessMgr func() SessionManager
}

// RequestChatRightsHandler returns SNAC oscar.ChatNavNavInfo, which contains
// chat navigation service parameters and limits.
func (s ChatNavService) RequestChatRightsHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.ChatNavTLVMaxConcurrentRooms, uint8(10)),
					oscar.NewTLV(oscar.ChatNavTLVExchangeInfo, oscar.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 4,
						TLVBlock: oscar.TLVBlock{
							TLVList: oscar.TLVList{
								oscar.NewTLV(oscar.ChatNavTLVClassPerms, uint16(0x0010)),
								oscar.NewTLV(oscar.ChatNavTLVFlags, uint16(15)),
								oscar.NewTLV(oscar.ChatNavTLVRoomName, "default exchange"),
								oscar.NewTLV(oscar.ChatNavTLVCreatePerms, uint8(2)),
								oscar.NewTLV(oscar.ChatNavTLVCharSet1, "us-ascii"),
								oscar.NewTLV(oscar.ChatNavTLVLang1, "en"),
								oscar.NewTLV(oscar.ChatNavTLVCharSet2, "us-ascii"),
								oscar.NewTLV(oscar.ChatNavTLVLang2, "en"),
							},
						},
					}),
				},
			},
		},
	}
}

// CreateRoomHandler creates a chat room with the current user as the first
// participant. It returns SNAC oscar.ChatNavNavInfo, which contains metadata
// for the chat room.
func (s ChatNavService) CreateRoomHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (oscar.SNACMessage, error) {
	name, hasName := inBody.String(oscar.ChatTLVRoomName)
	if !hasName {
		return oscar.SNACMessage{}, errors.New("unable to find chat name")
	}

	room := s.newChatRoom()
	room.DetailLevel = inBody.DetailLevel
	room.Exchange = inBody.Exchange
	room.InstanceNumber = inBody.InstanceNumber
	room.Name = name

	chatSessMgr := s.newChatSessMgr()

	s.chatRegistry.Register(room, chatSessMgr)

	// add user to chat room
	chatSess := chatSessMgr.AddSession(sess.ID(), sess.ScreenName())
	chatSess.SetChatID(room.Cookie)

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.ChatNavRequestRoomInfo, oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       inBody.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: inBody.InstanceNumber,
						DetailLevel:    inBody.DetailLevel,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}

// RequestRoomInfoHandler returns oscar.ChatNavNavInfo, which contains metadata
// for the chat room specified in the inFrame.Cookie.
func (s ChatNavService) RequestRoomInfoHandler(_ context.Context, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (oscar.SNACMessage, error) {
	room, _, err := s.chatRegistry.Retrieve(string(inBody.Cookie))
	if err != nil {
		return oscar.SNACMessage{}, err
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.ChatNavRequestRoomInfo, oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       room.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: room.InstanceNumber,
						DetailLevel:    room.DetailLevel,
						TLVBlock: oscar.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}
