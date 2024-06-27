package foodgroup

import (
	"context"
	"errors"
	"fmt"
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
		wire.NewTLV(wire.ChatRoomTLVNavCreatePerms, uint8(2)),
		wire.NewTLV(wire.ChatRoomTLVCharSet1, "us-ascii"),
		wire.NewTLV(wire.ChatRoomTLVLang1, "en"),
		wire.NewTLV(wire.ChatRoomTLVCharSet2, "us-ascii"),
		wire.NewTLV(wire.ChatRoomTLVLang2, "en"),
	},
}

// NewChatNavService creates a new instance of NewChatNavService.
func NewChatNavService(logger *slog.Logger, chatRoomManager ChatRoomRegistry, fnNewChatRoom func() state.ChatRoom) *ChatNavService {
	return &ChatNavService{
		logger:          logger,
		chatRoomManager: chatRoomManager,
		fnNewChatRoom:   fnNewChatRoom,
	}
}

// ChatNavService provides functionality for the ChatNav food group, which
// handles chat room creation and serving chat room metadata.
type ChatNavService struct {
	logger          *slog.Logger
	chatRoomManager ChatRoomRegistry
	fnNewChatRoom   func() state.ChatRoom
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
						Identifier: state.PrivateExchange,
						TLVBlock:   defaultExchangeCfg,
					}),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: state.PublicExchange,
						TLVBlock:   defaultExchangeCfg,
					}),
				},
			},
		},
	}
}

// CreateRoom creates and returns a chat room or returns an existing chat
// room. It returns SNAC wire.ChatNavNavInfo, which contains metadata for the
// chat room.
func (s ChatNavService) CreateRoom(_ context.Context, sess *state.Session, inFrame wire.SNACFrame, inBody wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate) (wire.SNACMessage, error) {
	if err := validateExchange(inBody.Exchange); err != nil {
		return wire.SNACMessage{}, err
	}
	if inBody.Cookie != "create" {
		s.logger.Info("got a non-create cookie", "value", inBody.Cookie)
	}

	name, hasName := inBody.String(wire.ChatRoomTLVRoomName)
	if !hasName {
		return wire.SNACMessage{}, errors.New("unable to find chat name in TLV payload")
	}

	// todo call ChatRoomByName and CreateChatRoom in a txn
	room, err := s.chatRoomManager.ChatRoomByName(inBody.Exchange, name)

	switch {
	case errors.Is(err, state.ErrChatRoomNotFound):
		if inBody.Exchange == state.PublicExchange {
			return wire.SNACMessage{}, fmt.Errorf("community chat rooms can only be created on exchange %d",
				state.PrivateExchange)
		}

		room = s.fnNewChatRoom()
		room.Creator = sess.IdentScreenName()
		room.DetailLevel = inBody.DetailLevel
		room.Exchange = inBody.Exchange
		room.InstanceNumber = inBody.InstanceNumber
		room.Name = name

		if err := s.chatRoomManager.CreateChatRoom(room); err != nil {
			return wire.SNACMessage{}, fmt.Errorf("unable to create chat room: %w", err)
		}
		break
	case err != nil:
		return wire.SNACMessage{}, fmt.Errorf("unable to retrieve chat room chat room %s on exchange %d: %w",
			name, inBody.Exchange, err)
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

// RequestRoomInfo returns wire.ChatNavNavInfo, which contains metadata for
// the chat room specified in the inFrame.hmacCookie.
func (s ChatNavService) RequestRoomInfo(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo) (wire.SNACMessage, error) {
	if err := validateExchange(inBody.Exchange); err != nil {
		return wire.SNACMessage{}, err
	}

	room, err := s.chatRoomManager.ChatRoomByCookie(inBody.Cookie)
	if err != nil {
		return wire.SNACMessage{}, fmt.Errorf("unable to find chat room: %w", err)
	}

	if room.Exchange != inBody.Exchange {
		return wire.SNACMessage{}, errors.New("chat room exchange does not match requested exchange")
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

func (s ChatNavService) ExchangeInfo(_ context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo) (wire.SNACMessage, error) {
	if err := validateExchange(inBody.Exchange); err != nil {
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
					wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: inBody.Exchange,
						TLVBlock:   defaultExchangeCfg,
					}),
				},
			},
		},
	}, nil
}

func validateExchange(exchange uint16) error {
	if !(exchange == state.PrivateExchange || exchange == state.PublicExchange) {
		return fmt.Errorf("only exchanges %d and %d are supported", state.PrivateExchange, state.PublicExchange)
	}
	return nil
}
