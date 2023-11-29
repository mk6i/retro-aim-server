package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

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
			svc := NewOServiceServiceForBOS(OServiceService{
				cfg: tc.cfg,
			}, chatRegistry)

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
			svc := NewOServiceService(server.Config{}, messageRelayer, feedbagManager)
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
	svc := NewOServiceService(server.Config{}, nil, nil)

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

func TestOServiceServiceForBOS_WriteOServiceHostOnline(t *testing.T) {
	svc := NewOServiceServiceForBOS(*NewOServiceService(server.Config{}, nil, nil), nil)

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostOnline,
		},
		Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				oscar.Alert,
				oscar.Buddy,
				oscar.ChatNav,
				oscar.Feedbag,
				oscar.ICBM,
				oscar.Locate,
				oscar.OService,
			},
		},
	}

	have := svc.WriteOServiceHostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceServiceForChat_WriteOServiceHostOnline(t *testing.T) {
	svc := NewOServiceServiceForChat(*NewOServiceService(server.Config{}, nil, nil), nil)

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostOnline,
		},
		Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				oscar.OService,
				oscar.Chat,
			},
		},
	}

	have := svc.WriteOServiceHostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceService_ClientVersionsHandler(t *testing.T) {
	svc := NewOServiceService(server.Config{}, nil, nil)

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceHostVersions,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: []uint16{5, 6, 7, 8},
		},
	}

	have := svc.ClientVersionsHandler(nil, oscar.SNACFrame{
		RequestID: 1234,
	}, oscar.SNAC_0x01_0x17_OServiceClientVersions{
		Versions: []uint16{5, 6, 7, 8},
	})

	assert.Equal(t, want, have)
}

func TestOServiceService_UserInfoQueryHandler(t *testing.T) {
	svc := NewOServiceService(server.Config{}, nil, nil)
	sess := newTestSession("test-user")

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceUserInfoUpdate,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}

	have := svc.UserInfoQueryHandler(nil, sess, oscar.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}

func TestOServiceService_IdleNotificationHandler(t *testing.T) {
	tests := []struct {
		name   string
		sess   *state.Session
		bodyIn oscar.SNAC_0x01_0x11_OServiceIdleNotification
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientBuddies is a list of the recipient's buddies that get
		// updated warning level
		recipientBuddies []string
		broadcastMessage oscar.SNACMessage
		wantErr          error
	}{
		{
			name: "set idle from active",
			sess: newTestSession("test-user"),
			bodyIn: oscar.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 90,
			},
			recipientScreenName: "test-user",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyArrived,
				},
				Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("test-user", sessOptIdle(90*time.Second)).TLVUserInfo(),
				},
			},
		},
		{
			name: "set active from idle",
			sess: newTestSession("test-user", sessOptIdle(90*time.Second)),
			bodyIn: oscar.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 0,
			},
			recipientScreenName: "test-user",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyArrived,
				},
				Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("test-user").TLVUserInfo(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				InterestedUsers(tt.recipientScreenName).
				Return(tt.recipientBuddies, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				BroadcastToScreenNames(mock.Anything, tt.recipientBuddies, tt.broadcastMessage).
				Maybe()

			svc := NewOServiceService(server.Config{}, messageRelayer, feedbagManager)

			haveErr := svc.IdleNotificationHandler(nil, tt.sess, tt.bodyIn)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForBOS_ClientOnlineHandler(t *testing.T) {
	type buddiesLookupParams []struct {
		screenName string
		buddies    []string
	}
	type interestedUsersParams []struct {
		screenName string
		users      []string
	}
	type broadcastToScreenNamesParams []struct {
		screenNames []string
		message     oscar.SNACMessage
	}
	type retrieveByScreenNameParams []struct {
		screenName string
		sess       *state.Session
	}
	type sendToScreenNameParams []struct {
		screenName string
		message    oscar.SNACMessage
	}

	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the session of the arriving user
		sess *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn oscar.SNAC_0x01_0x02_OServiceClientOnline
		// buddyLookupParams contains params for looking up arriving user's
		// buddies
		buddyLookupParams buddiesLookupParams
		// interestedUsersParams contains params for looking up users who have
		// the arriving user on their buddy list
		interestedUsersParams interestedUsersParams
		// broadcastToScreenNamesParams contains params for sending
		// buddy online notification to users who have the arriving user on
		// their buddy list
		broadcastToScreenNamesParams broadcastToScreenNamesParams
		// retrieveByScreenNameParams contains params for looking up the
		// session for each of the arriving user's buddies
		retrieveByScreenNameParams retrieveByScreenNameParams
		// sendToScreenNameParams contains params for sending arrival
		// notifications for each of the arriving user's buddies to the
		// arriving user's client
		sendToScreenNameParams sendToScreenNameParams
		wantErr                error
	}{
		{
			name:   "notify arriving user's buddies of its arrival and populate the arriving user's buddy list",
			sess:   newTestSession("test-user"),
			bodyIn: oscar.SNAC_0x01_0x02_OServiceClientOnline{},
			interestedUsersParams: interestedUsersParams{
				{
					screenName: "test-user",
					users:      []string{"buddy1", "buddy2", "buddy3", "buddy4"},
				},
			},
			broadcastToScreenNamesParams: broadcastToScreenNamesParams{
				{
					screenNames: []string{"buddy1", "buddy2", "buddy3", "buddy4"},
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyArrived,
						},
						Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
							TLVUserInfo: newTestSession("test-user").TLVUserInfo(),
						},
					},
				},
			},
			buddyLookupParams: buddiesLookupParams{
				{
					screenName: "test-user",
					buddies:    []string{"buddy1", "buddy3"},
				},
			},
			retrieveByScreenNameParams: retrieveByScreenNameParams{
				{
					screenName: "buddy1",
					sess:       newTestSession("buddy1"),
				},
				{
					screenName: "buddy3",
					sess:       newTestSession("buddy3"),
				},
			},
			sendToScreenNameParams: sendToScreenNameParams{
				{
					screenName: "test-user",
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyArrived,
						},
						Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
							TLVUserInfo: newTestSession("buddy1").TLVUserInfo(),
						},
					},
				},
				{
					screenName: "test-user",
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Buddy,
							SubGroup:  oscar.BuddyArrived,
						},
						Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
							TLVUserInfo: newTestSession("buddy3").TLVUserInfo(),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.interestedUsersParams {
				feedbagManager.EXPECT().
					InterestedUsers(params.screenName).
					Return(params.users, nil).
					Maybe()
			}
			for _, params := range tt.broadcastToScreenNamesParams {
				messageRelayer.EXPECT().
					BroadcastToScreenNames(mock.Anything, params.screenNames, params.message).
					Maybe()
			}
			for _, params := range tt.buddyLookupParams {
				feedbagManager.EXPECT().
					Buddies(params.screenName).
					Return(params.buddies, nil).
					Maybe()
			}
			for _, params := range tt.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess).
					Maybe()
			}
			for _, params := range tt.sendToScreenNameParams {
				messageRelayer.EXPECT().
					SendToScreenName(mock.Anything, params.screenName, params.message).
					Maybe()
			}

			svc := NewOServiceServiceForBOS(OServiceService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}, nil)

			haveErr := svc.ClientOnlineHandler(nil, tt.bodyIn, tt.sess)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForChat_ClientOnlineHandler(t *testing.T) {
	chatter1 := newTestSession("chatter-1")
	chatter2 := newTestSession("chatter-2")
	chatRoom := state.ChatRoom{
		Cookie:         "the-cookie",
		DetailLevel:    1,
		Exchange:       2,
		InstanceNumber: 3,
		Name:           "the-chat-room",
	}

	type participantsParams []*state.Session
	type broadcastExcept []struct {
		sess    *state.Session
		message oscar.SNACMessage
	}
	type sendToScreenNameParams []struct {
		screenName string
		message    oscar.SNACMessage
	}

	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the user joining the chat room
		joiningChatter *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn oscar.SNAC_0x01_0x02_OServiceClientOnline
		// participantsParams contains all the chat room participants
		participantsParams participantsParams
		// broadcastExcept contains params for broadcasting chat arrival to all
		// chat participants except the user joining
		broadcastExcept broadcastExcept
		// sendToScreenNameParams contains params for sending chat room
		// metadata and chat participant list to joining user
		sendToScreenNameParams sendToScreenNameParams
		wantErr                error
	}{
		{
			name:           "upon joining, send chat room metadata and participant list to joining user; alert arrival to existing participants",
			joiningChatter: chatter1,
			bodyIn:         oscar.SNAC_0x01_0x02_OServiceClientOnline{},
			broadcastExcept: broadcastExcept{
				{
					sess: chatter1,
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Chat,
							SubGroup:  oscar.ChatUsersJoined,
						},
						Body: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
							Users: []oscar.TLVUserInfo{
								{
									ScreenName:   chatter1.ScreenName(),
									WarningLevel: chatter1.Warning(),
									TLVBlock: oscar.TLVBlock{
										TLVList: chatter1.UserInfo(),
									},
								},
							},
						},
					},
				},
			},
			participantsParams: participantsParams{
				chatter1,
				chatter2,
			},
			sendToScreenNameParams: sendToScreenNameParams{
				{
					screenName: chatter1.ScreenName(),
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Chat,
							SubGroup:  oscar.ChatRoomInfoUpdate,
						},
						Body: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
							Exchange:       chatRoom.Exchange,
							Cookie:         chatRoom.Cookie,
							InstanceNumber: chatRoom.InstanceNumber,
							DetailLevel:    chatRoom.DetailLevel,
							TLVBlock: oscar.TLVBlock{
								TLVList: chatRoom.TLVList(),
							},
						},
					},
				},
				{
					screenName: chatter1.ScreenName(),
					message: oscar.SNACMessage{
						Frame: oscar.SNACFrame{
							FoodGroup: oscar.Chat,
							SubGroup:  oscar.ChatUsersJoined,
						},
						Body: oscar.SNAC_0x0E_0x03_ChatUsersJoined{
							Users: []oscar.TLVUserInfo{
								chatter1.TLVUserInfo(),
								chatter2.TLVUserInfo(),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tt.broadcastExcept {
				chatMessageRelayer.EXPECT().
					BroadcastExcept(mock.Anything, params.sess, params.message).
					Maybe()
			}
			chatMessageRelayer.EXPECT().
				Participants().
				Return(tt.participantsParams).
				Maybe()
			for _, params := range tt.sendToScreenNameParams {
				chatMessageRelayer.EXPECT().
					SendToScreenName(mock.Anything, params.screenName, params.message).
					Maybe()
			}

			chatRegistry := state.NewChatRegistry()
			chatRegistry.Register(chatRoom, chatMessageRelayer)

			svc := NewOServiceServiceForChat(OServiceService{
				feedbagManager: feedbagManager,
				messageRelayer: chatMessageRelayer,
			}, chatRegistry)

			haveErr := svc.ClientOnlineHandler(nil, tt.joiningChatter, chatRoom.Cookie)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}
