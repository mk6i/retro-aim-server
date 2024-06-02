package foodgroup

import (
	"context"
	"errors"

	"log/slog"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

var defaultExchangeCfg = wire.TLVBlock{
	TLVList: wire.TLVList{
		wire.NewTLV(wire.ChatRoomTLVMaxConcurrentRooms, uint8(10)),
		wire.NewTLV(wire.ChatRoomTLVClassPerms, uint16(0x0010)),
		wire.NewTLV(wire.ChatRoomTLVMaxNameLen, uint16(100)),
		wire.NewTLV(wire.ChatRoomTLVFlags, uint16(15)),
		wire.NewTLV(wire.ChatRoomTLVRoomName, "default exchange"),
		wire.NewTLV(wire.ChatRoomTLVNavCreatePerms, uint8(2)),
		wire.NewTLV(wire.ChatRoomTLVCharSet1, "us-ascii"),
		wire.NewTLV(wire.ChatRoomTLVLang1, "en"),
		wire.NewTLV(wire.ChatRoomTLVCharSet2, "us-ascii"),
		wire.NewTLV(wire.ChatRoomTLVLang2, "en"),
	},
}

// NewChatNavService creates a new instance of NewChatNavService.
func NewChatNavService(logger *slog.Logger, chatRegistry *state.ChatRegistry, newChatRoom func() state.ChatRoom, newChatSessMgr func() SessionManager) *ChatNavService {
	return &ChatNavService{
		logger:         logger,
		chatRegistry:   chatRegistry,
		newChatRoom:    newChatRoom,
		newChatSessMgr: newChatSessMgr,
	}
}

// ChatNavService provides functionality for the ChatNav food group, which
// handles chat room creation and serving chat room metadata.
type ChatNavService struct {
	logger         *slog.Logger
	chatRegistry   *state.ChatRegistry
	newChatRoom    func() state.ChatRoom
	newChatSessMgr func() SessionManager
}

// RequestChatRights returns SNAC wire.ChatNavNavInfo, which contains chat
// navigation service parameters and limits.
func (s ChatNavService) RequestChatRights(_ context.Context, inFrame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 4,
						TLVBlock:   defaultExchangeCfg,
					}),
				},
			},
		},
	}
}

// CreateRoom creates a chat room with the current user as the first
// participant. It returns SNAC wire.ChatNavNavInfo, which contains metadata
// for the chat room.
func (s ChatNavService) CreateRoom(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (wire.SNACMessage, error) {
	name, hasName := inBody.String(wire.ChatRoomTLVRoomName)
	if !hasName {
		return wire.SNACMessage{}, errors.New("unable to find chat name")
	}

	room := s.newChatRoom()
	room.DetailLevel = inBody.DetailLevel
	room.Exchange = inBody.Exchange
	room.InstanceNumber = inBody.InstanceNumber
	room.Name = name

	chatSessMgr := s.newChatSessMgr()

	s.chatRegistry.Register(room, chatSessMgr)

	// add user to chat room
	chatSess := chatSessMgr.AddSession(sess.ScreenName())
	chatSess.SetChatRoomCookie(room.Cookie)

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ChatNavTLVRoomInfo, wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       inBody.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: inBody.InstanceNumber,
						DetailLevel:    inBody.DetailLevel,
						TLVBlock: wire.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}

// RequestRoomInfo returns wire.ChatNavNavInfo, which contains metadata for
// the chat room specified in the inFrame.hmacCookie.
func (s ChatNavService) RequestRoomInfo(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (wire.SNACMessage, error) {
	room, _, err := s.chatRegistry.Retrieve(inBody.Cookie)
	if err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ChatNavTLVRoomInfo, wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       room.Exchange,
						Cookie:         room.Cookie,
						InstanceNumber: room.InstanceNumber,
						DetailLevel:    room.DetailLevel,
						TLVBlock: wire.TLVBlock{
							TLVList: room.TLVList(),
						},
					}),
				},
			},
		},
	}, nil
}

func (s ChatNavService) ExchangeInfo(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: inFrame.RequestID,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: inBody.Exchange,
						TLVBlock:   defaultExchangeCfg,
					}),
				},
			},
		},
	}
}
