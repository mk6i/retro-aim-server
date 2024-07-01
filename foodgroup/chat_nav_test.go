package foodgroup

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestChatNavService_CreateRoom(t *testing.T) {
	basicChatRoom := state.ChatRoom{
		Cookie:         "dummy-cookie",
		CreateTime:     time.UnixMilli(0),
		Creator:        state.NewIdentScreenName("the-screen-name"),
		DetailLevel:    3,
		Exchange:       4,
		InstanceNumber: 2,
		Name:           "the-chat-room-name",
	}
	publicChatRoom := state.ChatRoom{
		Cookie:         "dummy-cookie",
		CreateTime:     time.UnixMilli(0),
		Creator:        state.NewIdentScreenName("the-screen-name"),
		DetailLevel:    3,
		Exchange:       5,
		InstanceNumber: 2,
		Name:           "the-public-chat-room-name",
	}

	tests := []struct {
		name          string
		chatRoom      state.ChatRoom
		sess          *state.Session
		inputSNAC     wire.SNACMessage
		want          wire.SNACMessage
		mockParams    mockParams
		wantErr       error
		fnNewChatRoom func() state.ChatRoom
	}{
		{
			name:     "create private room that already exists",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       basicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, basicChatRoom.Name),
						},
					},
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
							wire.NewTLV(
								wire.ChatNavRequestRoomInfo,
								wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
									Exchange:       basicChatRoom.Exchange,
									Cookie:         basicChatRoom.Cookie,
									InstanceNumber: basicChatRoom.InstanceNumber,
									DetailLevel:    basicChatRoom.DetailLevel,
									TLVBlock: wire.TLVBlock{
										TLVList: basicChatRoom.TLVList(),
									},
								},
							),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: basicChatRoom.Exchange,
							name:     basicChatRoom.Name,
							room:     basicChatRoom,
						},
					},
				},
			},
		},
		{
			name:     "create private room that doesn't already exist",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       basicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, basicChatRoom.Name),
						},
					},
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
							wire.NewTLV(
								wire.ChatNavRequestRoomInfo,
								wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
									Exchange:       basicChatRoom.Exchange,
									Cookie:         basicChatRoom.Cookie,
									InstanceNumber: basicChatRoom.InstanceNumber,
									DetailLevel:    basicChatRoom.DetailLevel,
									TLVBlock: wire.TLVBlock{
										TLVList: basicChatRoom.TLVList(),
									},
								},
							),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: basicChatRoom.Exchange,
							name:     basicChatRoom.Name,
							err:      state.ErrChatRoomNotFound,
						},
					},
					createChatRoomParams: createChatRoomParams{
						{
							room: basicChatRoom,
						},
					},
				},
			},
			fnNewChatRoom: func() state.ChatRoom {
				return basicChatRoom
			},
		},
		{
			name:     "create public room that already exists",
			chatRoom: publicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       publicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: publicChatRoom.InstanceNumber,
					DetailLevel:    publicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, publicChatRoom.Name),
						},
					},
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
							wire.NewTLV(
								wire.ChatNavRequestRoomInfo,
								wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
									Exchange:       publicChatRoom.Exchange,
									Cookie:         publicChatRoom.Cookie,
									InstanceNumber: publicChatRoom.InstanceNumber,
									DetailLevel:    publicChatRoom.DetailLevel,
									TLVBlock: wire.TLVBlock{
										TLVList: publicChatRoom.TLVList(),
									},
								},
							),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: publicChatRoom.Exchange,
							name:     publicChatRoom.Name,
							room:     publicChatRoom,
						},
					},
				},
			},
		},
		{
			name:     "create public room that does not exist",
			chatRoom: publicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       publicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: publicChatRoom.InstanceNumber,
					DetailLevel:    publicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, publicChatRoom.Name),
						},
					},
				},
			},
			want: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNoMatch,
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: publicChatRoom.Exchange,
							name:     publicChatRoom.Name,
							err:      state.ErrChatRoomNotFound,
						},
					},
				},
			},
			fnNewChatRoom: func() state.ChatRoom {
				return basicChatRoom
			},
		},
		{
			name:     "create room invalid exchange number",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       1337,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, basicChatRoom.Name),
						},
					},
				},
			},
			want: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
		{
			name:     "incoming create room missing name tlv",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       basicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{}, // intentionally empty for test
					},
				},
			},
			want:    wire.SNACMessage{},
			wantErr: errChatNavRoomNameMissing,
		},
		{
			name:     "create private room failed",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       basicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, basicChatRoom.Name),
						},
					},
				},
			},
			want:    wire.SNACMessage{},
			wantErr: errChatNavRoomCreateFailed,
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: basicChatRoom.Exchange,
							name:     basicChatRoom.Name,
							err:      state.ErrChatRoomNotFound,
						},
					},
					createChatRoomParams: createChatRoomParams{
						{
							room: basicChatRoom,
							err:  errors.New("fake database error"),
						},
					},
				},
			},
			fnNewChatRoom: func() state.ChatRoom {
				return basicChatRoom
			},
		},
		{
			name:     "create private room failed unknown error",
			chatRoom: basicChatRoom,
			sess:     newTestSession("the-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange:       basicChatRoom.Exchange,
					Cookie:         "create", // actual canned value sent by AIM client
					InstanceNumber: basicChatRoom.InstanceNumber,
					DetailLevel:    basicChatRoom.DetailLevel,
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.ChatRoomTLVRoomName, basicChatRoom.Name),
						},
					},
				},
			},
			want:    wire.SNACMessage{},
			wantErr: errChatNavRetrieveFailed,
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByNameParams: chatRoomByNameParams{
						{
							exchange: basicChatRoom.Exchange,
							name:     basicChatRoom.Name,
							err:      errors.New("fake database error"),
						},
					},
				},
			},
			fnNewChatRoom: func() state.ChatRoom {
				return basicChatRoom
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRoomRegistry := newMockChatRoomRegistry(t)
			for _, params := range tt.mockParams.chatRoomByNameParams {
				chatRoomRegistry.EXPECT().
					ChatRoomByName(params.exchange, params.name).
					Return(params.room, params.err)
			}
			for _, params := range tt.mockParams.createChatRoomParams {
				chatRoomRegistry.EXPECT().
					CreateChatRoom(params.room).
					Return(params.err)
			}

			svc := NewChatNavService(slog.Default(), chatRoomRegistry, tt.fnNewChatRoom)
			outputSNAC, err := svc.CreateRoom(context.Background(), tt.sess, tt.inputSNAC.Frame, tt.inputSNAC.Body.(wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate))
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, outputSNAC)
		})
	}
}

func TestChatNavService_RequestRoomInfo(t *testing.T) {
	tests := []struct {
		name       string
		inputSNAC  wire.SNACMessage
		want       wire.SNACMessage
		mockParams mockParams
		wantErr    error
	}{
		{
			name: "request existing private room info successfully",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: state.PrivateExchange,
					Cookie:   "the-chat-cookie",
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
								Exchange:       state.PrivateExchange,
								InstanceNumber: 8,
								TLVBlock: wire.TLVBlock{
									TLVList: state.ChatRoom{Cookie: "the-chat-cookie"}.TLVList(),
								},
							}),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: "the-chat-cookie",
							room: state.ChatRoom{
								Cookie:         "the-chat-cookie",
								DetailLevel:    2,
								Exchange:       state.PrivateExchange,
								InstanceNumber: 8,
							},
						},
					},
				},
			},
		},
		{
			name: "request existing public room info successfully",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: state.PublicExchange,
					Cookie:   "the-chat-cookie",
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
								Exchange:       state.PublicExchange,
								InstanceNumber: 8,
								TLVBlock: wire.TLVBlock{
									TLVList: state.ChatRoom{Cookie: "the-chat-cookie"}.TLVList(),
								},
							}),
						},
					},
				},
			},
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: "the-chat-cookie",
							room: state.ChatRoom{
								Cookie:         "the-chat-cookie",
								DetailLevel:    2,
								Exchange:       state.PublicExchange,
								InstanceNumber: 8,
							},
						},
					},
				},
			},
		},
		{
			name: "request room info with invalid exchange number",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: 1337,
					Cookie:   "the-chat-cookie",
				},
			},
			want: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
		{
			name: "request room info on nonexistent room",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: state.PrivateExchange,
					Cookie:   "the-chat-cookie",
				},
			},
			want:    wire.SNACMessage{},
			wantErr: errChatNavMismatchedExchange,
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: "the-chat-cookie",
							room: state.ChatRoom{
								Cookie:         "the-chat-cookie",
								DetailLevel:    2,
								Exchange:       state.PublicExchange,
								InstanceNumber: 8,
							},
						},
					},
				},
			},
		},
		{
			name: "request room info with mismatched exchange",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: state.PrivateExchange,
					Cookie:   "the-chat-cookie",
				},
			},
			want:    wire.SNACMessage{},
			wantErr: state.ErrChatRoomNotFound,
			mockParams: mockParams{
				chatRoomRegistryParams: chatRoomRegistryParams{
					chatRoomByCookieParams: chatRoomByCookieParams{
						{
							cookie: "the-chat-cookie",
							room: state.ChatRoom{
								Cookie:         "the-chat-cookie",
								DetailLevel:    2,
								Exchange:       state.PrivateExchange,
								InstanceNumber: 8,
							},
							err: state.ErrChatRoomNotFound,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRoomRegistry := newMockChatRoomRegistry(t)
			for _, params := range tt.mockParams.chatRoomByCookieParams {
				chatRoomRegistry.EXPECT().
					ChatRoomByCookie(params.cookie).
					Return(params.room, params.err)
			}

			svc := NewChatNavService(slog.Default(), chatRoomRegistry, nil)
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
	svc := NewChatNavService(nil, nil, nil)

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
						},
					}),
					wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
						Identifier: 5,
						TLVBlock: wire.TLVBlock{
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
						},
					}),
				},
			},
		},
	}

	assert.Equal(t, want, have)
}

func TestChatNavService_ExchangeInfo(t *testing.T) {
	tests := []struct {
		name       string
		inputSNAC  wire.SNACMessage
		want       wire.SNACMessage
		mockParams mockParams
		wantErr    error
	}{
		{
			name: "request private exchange info",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo{
					Exchange: state.PrivateExchange,
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
							wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
							wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
								Identifier: state.PrivateExchange,
								TLVBlock: wire.TLVBlock{
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
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "request public exchange info",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo{
					Exchange: state.PublicExchange,
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
							wire.NewTLV(wire.ChatNavTLVMaxConcurrentRooms, uint8(10)),
							wire.NewTLV(wire.ChatNavTLVExchangeInfo, wire.SNAC_0x0D_0x09_TLVExchangeInfo{
								Identifier: state.PublicExchange,
								TLVBlock: wire.TLVBlock{
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
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "request invalid exchange info",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo{
					Exchange: 6,
				},
			},
			want: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChatNavService(slog.Default(), nil, nil)
			outputSNAC, err := svc.ExchangeInfo(context.Background(), tt.inputSNAC.Frame,
				tt.inputSNAC.Body.(wire.SNAC_0x0D_0x03_ChatNavRequestExchangeInfo))
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.want, outputSNAC)
		})
	}
}
