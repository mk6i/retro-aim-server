package handler

import (
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
)

func TestReceiveAndSendServiceRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// config is the application config
		cfg server.Config
		// chatRoom is the chat room the user connects to
		chatRoom *state.ChatRoom
		// userSession is the session of the user requesting the chat service
		// info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectOutput oscar.SNACMessage
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name:        "request info for ICBM service, return invalid SNAC err",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: oscar.ICBM,
				},
			},
			expectErr: server.ErrUnsupportedSubGroup,
		},
		{
			name: "request info for connecting to chat room, return chat service and chat room metadata",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom: &state.ChatRoom{
				CreateTime:     time.UnixMilli(0),
				DetailLevel:    4,
				Exchange:       8,
				Cookie:         "the-chat-cookie",
				InstanceNumber: 16,
				Name:           "my new chat",
			},
			userSession: newTestSession("user_screen_name", sessOptCannedID),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: oscar.Chat,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, oscar.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         []byte("the-chat-cookie"),
								InstanceNumber: 16,
							}),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, server.ChatCookie{
								Cookie: []byte("the-chat-cookie"),
								SessID: "user-sess-id",
							}),
							oscar.NewTLV(oscar.OServiceTLVTagsGroupID, oscar.Chat),
							oscar.NewTLV(oscar.OServiceTLVTagsSSLCertName, ""),
							oscar.NewTLV(oscar.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to non-existent chat room, return SNAC error",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom:    nil,
			userSession: newTestSession("user_screen_name", sessOptCannedID),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: oscar.Chat,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, oscar.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         []byte("the-chat-cookie"),
								InstanceNumber: 16,
							}),
						},
					},
				},
			},
			expectErr: server.ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			sessionManager := newMockSessionManager(t)
			chatRegistry := state.NewChatRegistry()
			if tc.chatRoom != nil {
				sessionManager.EXPECT().
					NewSessionWithSN(tc.userSession.ID(), tc.userSession.ScreenName()).
					Return(&state.Session{}).
					Maybe()
				chatRegistry.Register(*tc.chatRoom, sessionManager)
			}
			//
			// send input SNAC
			//
			svc := OServiceServiceForBOS{
				OServiceService: OServiceService{
					cfg: tc.cfg,
				},
				chatRegistry: chatRegistry,
			}

			outputSNAC, err := svc.ServiceRequestHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x01_0x04_OServiceServiceRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestSetUserInfoFieldsHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user whose info is being set
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// expectOutput is the SNAC reply sent from the server back to the
		// client
		expectOutput oscar.SNACMessage
		// broadcastMessage is the arrival/departure message sent to buddies
		broadcastMessage []struct {
			recipients []string
			msg        oscar.SNACMessage
		}
		// interestedUserLookups contains all the users who have this user on
		// their buddy list
		interestedUserLookups map[string][]string
		// expectErr is the expected error returned
		expectErr error
	}{
		{
			name:        "set user status to visible",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.OServiceUserInfoStatus, uint32(0x0000)),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("user_screen_name").TLVUserInfo(),
				},
			},
			broadcastMessage: []struct {
				recipients []string
				msg        oscar.SNACMessage
			}{
				{
					recipients: []string{"friend1", "friend2"},
					msg: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyArrived,
						},
						Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
							TLVUserInfo: newTestSession("user_screen_name").TLVUserInfo(),
						},
					},
				},
			},
			interestedUserLookups: map[string][]string{
				"user_screen_name": {"friend1", "friend2"},
			},
		},
		{
			name:        "set user status to invisible",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.OServiceUserInfoStatus, uint32(0x0100)),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("user_screen_name", sessOptInvisible).TLVUserInfo(),
				},
			},
			broadcastMessage: []struct {
				recipients []string
				msg        oscar.SNACMessage
			}{
				{
					recipients: []string{"friend1", "friend2"},
					msg: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyDeparted,
						},
						Body: oscar.SNAC_0x03_0x0C_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "user_screen_name",
								WarningLevel: 0,
							},
						},
					},
				},
			},
			interestedUserLookups: map[string][]string{
				"user_screen_name": {"friend1", "friend2"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			feedbagManager := newMockFeedbagManager(t)
			for user, friends := range tc.interestedUserLookups {
				feedbagManager.EXPECT().
					InterestedUsers(user).
					Return(friends, nil).
					Maybe()
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, broadcastMsg := range tc.broadcastMessage {
				messageRelayer.EXPECT().BroadcastToScreenNames(mock.Anything, broadcastMsg.recipients, broadcastMsg.msg)
			}
			//
			// send input SNAC
			//
			svc := OServiceService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}
			outputSNAC, err := svc.SetUserInfoFieldsHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestOServiceService_RateParamsQueryHandler(t *testing.T) {

	svc := OServiceService{}

	have := svc.RateParamsQueryHandler(nil, oscar.SNACFrame{RequestID: 1234})
	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceRateParamsReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
			RateClasses: []struct {
				ID              uint16
				WindowSize      uint32
				ClearLevel      uint32
				AlertLevel      uint32
				LimitLevel      uint32
				DisconnectLevel uint32
				CurrentLevel    uint32
				MaxLevel        uint32
				LastTime        uint32
				CurrentState    uint8
			}{
				{
					ID:              0x0001,
					WindowSize:      0x00000050,
					ClearLevel:      0x000009C4,
					AlertLevel:      0x000007D0,
					LimitLevel:      0x000005DC,
					DisconnectLevel: 0x00000320,
					CurrentLevel:    0x00000D69,
					MaxLevel:        0x00001770,
					LastTime:        0x00000000,
					CurrentState:    0x00,
				},
			},
			RateGroups: []struct {
				ID    uint16
				Pairs []struct {
					FoodGroup uint16
					SubGroup  uint16
				} `count_prefix:"uint16"`
			}{
				{
					ID: 1,
					Pairs: []struct {
						FoodGroup uint16
						SubGroup  uint16
					}{
						{
							FoodGroup: oscar.ICBM,
							SubGroup:  oscar.ICBMChannelMsgToHost,
						},
						{
							FoodGroup: oscar.Chat,
							SubGroup:  oscar.ChatChannelMsgToHost,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, want, have)
}
