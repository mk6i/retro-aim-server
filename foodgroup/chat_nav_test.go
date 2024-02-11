package foodgroup

import (
	"context"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestChatNavService_CreateRoom(t *testing.T) {
	bosSess := newTestSession("user-screen-name", sessOptCannedID)
	chatSess := &state.Session{}

	chatRegistry := state.NewChatRegistry()

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().AddSession(bosSess.ID(), bosSess.ScreenName()).
		Return(chatSess)

	newChatRoom := func() state.ChatRoom {
		return state.ChatRoom{
			Cookie:     "dummy-cookie",
			CreateTime: time.UnixMilli(0),
		}
	}
	newChatSessMgr := func() SessionManager {
		return sessionManager
	}

	inFrame := wire.SNACFrame{
		RequestID: 1234,
	}
	inBody := wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange:       1,
		Cookie:         "create", // actual canned value sent by AIM client
		InstanceNumber: 2,
		DetailLevel:    3,
		TLVBlock: wire.TLVBlock{
			TLVList: wire.TLVList{
				wire.NewTLV(wire.ChatTLVRoomName, "the-chat-room-name"),
			},
		},
	}

	svc := NewChatNavService(nil, chatRegistry, newChatRoom, newChatSessMgr)
	outputSNAC, err := svc.CreateRoom(context.Background(), bosSess, inFrame, inBody)
	assert.NoError(t, err)

	assert.Equal(t, chatSess.ChatRoomCookie(), newChatRoom().Cookie)

	expectChatRoom := state.ChatRoom{
		Cookie:         "dummy-cookie",
		CreateTime:     time.UnixMilli(0),
		DetailLevel:    3,
		Exchange:       1,
		InstanceNumber: 2,
		Name:           "the-chat-room-name",
	}
	chatRoom, _, err := chatRegistry.Retrieve("dummy-cookie")
	assert.NoError(t, err)

	// assert the user session is linked to the chat room
	assert.Equal(t, expectChatRoom, chatRoom)

	expectSNAC := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(
						wire.ChatNavRequestRoomInfo,
						wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
							Exchange:       chatRoom.Exchange,
							Cookie:         chatRoom.Cookie,
							InstanceNumber: chatRoom.InstanceNumber,
							DetailLevel:    chatRoom.DetailLevel,
							TLVBlock: wire.TLVBlock{
								TLVList: chatRoom.TLVList(),
							},
						},
					),
				},
			},
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}

func TestChatNavService_RequestRoomInfo(t *testing.T) {
	tests := []struct {
		name     string
		chatRoom state.ChatRoom
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		want      wire.SNACMessage
		wantErr   error
	}{
		{
			name: "request room info",
			chatRoom: state.ChatRoom{
				Cookie:         "the-chat-cookie",
				DetailLevel:    2,
				Exchange:       4,
				InstanceNumber: 8,
			},
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Cookie: "the-chat-cookie",
				},
			},
			want: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavNavInfo,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(0x04, wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
								Cookie:         "the-chat-cookie",
								DetailLevel:    2,
								Exchange:       4,
								InstanceNumber: 8,
								TLVBlock: wire.TLVBlock{
									TLVList: state.ChatRoom{Cookie: "the-chat-cookie"}.TLVList(),
								},
							}),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChatNavService(nil, state.NewChatRegistry(), nil, nil)
			svc.chatRegistry.Register(tt.chatRoom, nil)
			got, err := svc.RequestRoomInfo(nil, tt.inputSNAC.Frame,
				tt.inputSNAC.Body.(wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo))
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChatNavService_RequestChatRights(t *testing.T) {
	svc := NewChatNavService(nil, nil, nil, nil)

	have := svc.RequestChatRights(nil, wire.SNACFrame{RequestID: 1234})

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ChatNav,
			SubGroup:  wire.ChatNavNavInfo,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 4,
						TLVBlock: wire.TLVBlock{
							TLVList: wire.TLVList{
								wire.NewTLV(wire.ChatNavTLVClassPerms, uint16(0x0010)),
								wire.NewTLV(wire.ChatNavTLVFlags, uint16(15)),
								wire.NewTLV(wire.ChatNavTLVRoomName, "default exchange"),
								wire.NewTLV(wire.ChatNavTLVCreatePerms, uint8(2)),
								wire.NewTLV(wire.ChatNavTLVCharSet1, "us-ascii"),
								wire.NewTLV(wire.ChatNavTLVLang1, "en"),
								wire.NewTLV(wire.ChatNavTLVCharSet2, "us-ascii"),
								wire.NewTLV(wire.ChatNavTLVLang2, "en"),
							},
						},
					}),
				},
			},
		},
	}

	assert.Equal(t, want, have)
}
