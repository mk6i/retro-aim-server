package handler

import (
	"context"
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
)

func TestChatNavService_CreateRoomHandler(t *testing.T) {
	userSess := newTestSession("user-screen-name", sessOptCannedID)

	chatRegistry := state.NewChatRegistry()

	sessionManager := newMockSessionManager(t)
	sessionManager.EXPECT().AddSession(userSess.ID(), userSess.ScreenName()).
		Return(&state.Session{})

	newChatRoom := func() state.ChatRoom {
		return state.ChatRoom{
			Cookie:     "dummy-cookie",
			CreateTime: time.UnixMilli(0),
		}
	}
	newChatSessMgr := func() SessionManager {
		return sessionManager
	}

	inFrame := oscar.SNACFrame{
		RequestID: 1234,
	}
	inBody := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange:       1,
		Cookie:         "create", // actual canned value sent by AIM client
		InstanceNumber: 2,
		DetailLevel:    3,
		TLVBlock: oscar.TLVBlock{
			TLVList: oscar.TLVList{
				oscar.NewTLV(oscar.ChatTLVRoomName, "the-chat-room-name"),
			},
		},
	}

	svc := NewChatNavService(nil, chatRegistry, newChatRoom, newChatSessMgr)
	outputSNAC, err := svc.CreateRoomHandler(context.Background(), userSess, inFrame, inBody)
	assert.NoError(t, err)

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
	assert.Equal(t, expectChatRoom, chatRoom)

	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(
						oscar.ChatNavRequestRoomInfo,
						oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
							Exchange:       chatRoom.Exchange,
							Cookie:         chatRoom.Cookie,
							InstanceNumber: chatRoom.InstanceNumber,
							DetailLevel:    chatRoom.DetailLevel,
							TLVBlock: oscar.TLVBlock{
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

func TestChatNavService_RequestRoomInfoHandler(t *testing.T) {
	tests := []struct {
		name     string
		chatRoom state.ChatRoom
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		want      oscar.SNACMessage
		wantErr   error
	}{
		{
			name: "request room info",
			chatRoom: state.ChatRoom{
				Cookie:         "the-chat-id",
				DetailLevel:    2,
				Exchange:       4,
				InstanceNumber: 8,
			},
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Cookie: []byte(`the-chat-id`),
				},
			},
			want: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavNavInfo,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x04, oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
								Cookie:         "the-chat-id",
								DetailLevel:    2,
								Exchange:       4,
								InstanceNumber: 8,
								TLVBlock: oscar.TLVBlock{
									TLVList: state.ChatRoom{Cookie: "the-chat-id"}.TLVList(),
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
			got, err := svc.RequestRoomInfoHandler(nil, tt.inputSNAC.Frame,
				tt.inputSNAC.Body.(oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo))
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChatNavService_RequestChatRightsHandler(t *testing.T) {
	svc := NewChatNavService(nil, nil, nil, nil)

	have := svc.RequestChatRightsHandler(nil, oscar.SNACFrame{RequestID: 1234})

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
			RequestID: 1234,
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

	assert.Equal(t, want, have)
}
