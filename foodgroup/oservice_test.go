package foodgroup

import (
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
	"github.com/stretchr/testify/assert"
)

func TestOServiceServiceForBOS_ServiceRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// config is the application config
		cfg config.Config
		// chatRoom is the chat room the user connects to
		chatRoom *state.ChatRoom
		// userSession is the session of the user requesting the chat service
		// info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectOutput wire.SNACMessage
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name:        "request info for ICBM service, return invalid SNAC err",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.ICBM,
				},
			},
			expectErr: wire.ErrUnsupportedFoodGroup,
		},
		{
			name: "request info for connecting to chat room, return chat service and chat room metadata",
			cfg: config.Config{
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
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Chat,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         "the-chat-cookie",
								InstanceNumber: 16,
							}),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceResponse,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							wire.NewTLV(wire.OServiceTLVTagsLoginCookie, chatLoginCookie{
								Cookie: "the-chat-cookie",
								SessID: "user-session-id",
							}),
							wire.NewTLV(wire.OServiceTLVTagsGroupID, wire.Chat),
							wire.NewTLV(wire.OServiceTLVTagsSSLCertName, ""),
							wire.NewTLV(wire.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to non-existent chat room, return ErrChatRoomNotFound",
			cfg: config.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom:    nil,
			userSession: newTestSession("user_screen_name", sessOptCannedID),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.Chat,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(0x01, wire.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         "the-chat-cookie",
								InstanceNumber: 16,
							}),
						},
					},
				},
			},
			expectErr: state.ErrChatRoomNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			sessionManager := newMockSessionManager(t)
			chatRegistry := state.NewChatRegistry()
			chatSess := &state.Session{}
			if tc.chatRoom != nil {
				sessionManager.EXPECT().
					AddSession(tc.userSession.ID(), tc.userSession.ScreenName()).
					Return(chatSess).
					Maybe()
				chatRegistry.Register(*tc.chatRoom, sessionManager)
			}
			//
			// send input SNAC
			//
			svc := NewOServiceServiceForBOS(OServiceService{
				cfg: tc.cfg,
			}, chatRegistry)

			outputSNAC, err := svc.ServiceRequest(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x01_0x04_OServiceServiceRequest))
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			// assert the user session is linked to the chat room
			assert.Equal(t, chatSess.ChatRoomCookie(), tc.chatRoom.Cookie)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestSetUserInfoFields(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user whose info is being set
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// expectOutput is the SNAC reply sent from the server back to the
		// client
		expectOutput wire.SNACMessage
		// broadcastMessage is the arrival/departure message sent to buddies
		broadcastMessage []struct {
			recipients []string
			msg        wire.SNACMessage
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
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.OServiceUserInfoStatus, uint32(0x0000)),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("user_screen_name").TLVUserInfo(),
				},
			},
			broadcastMessage: []struct {
				recipients []string
				msg        wire.SNACMessage
			}{
				{
					recipients: []string{"friend1", "friend2"},
					msg: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyArrived,
						},
						Body: wire.SNAC_0x03_0x0B_BuddyArrived{
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
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLV(wire.OServiceUserInfoStatus, uint32(0x0100)),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceUserInfoUpdate,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: newTestSession("user_screen_name", sessOptInvisible).TLVUserInfo(),
				},
			},
			broadcastMessage: []struct {
				recipients []string
				msg        wire.SNACMessage
			}{
				{
					recipients: []string{"friend1", "friend2"},
					msg: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyDeparted,
						},
						Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
							TLVUserInfo: wire.TLVUserInfo{
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
					AdjacentUsers(user).
					Return(friends, nil).
					Maybe()
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, broadcastMsg := range tc.broadcastMessage {
				messageRelayer.EXPECT().RelayToScreenNames(mock.Anything, broadcastMsg.recipients, broadcastMsg.msg)
			}
			//
			// send input SNAC
			//
			svc := NewOServiceService(config.Config{}, messageRelayer, feedbagManager)
			outputSNAC, err := svc.SetUserInfoFields(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x01_0x1E_OServiceSetUserInfoFields))
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

func TestOServiceService_RateParamsQuery(t *testing.T) {
	svc := NewOServiceService(config.Config{}, nil, nil)

	have := svc.RateParamsQuery(nil, wire.SNACFrame{RequestID: 1234})
	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceRateParamsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x01_0x07_OServiceRateParamsReply{
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
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyRightsQuery,
						},
						{
							FoodGroup: wire.Chat,
							SubGroup:  wire.ChatChannelMsgToHost,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavRequestChatRights,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavRequestRoomInfo,
						},
						{
							FoodGroup: wire.ChatNav,
							SubGroup:  wire.ChatNavCreateRoom,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagRightsQuery,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagQuery,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagQueryIfModified,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagUse,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagInsertItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagUpdateItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagDeleteItem,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagStartCluster,
						},
						{
							FoodGroup: wire.Feedbag,
							SubGroup:  wire.FeedbagEndCluster,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMAddParameters,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMParameterQuery,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMChannelMsgToHost,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMEvilRequest,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMClientErr,
						},
						{
							FoodGroup: wire.ICBM,
							SubGroup:  wire.ICBMClientEvent,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateRightsQuery,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetDirInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateGetDirInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateSetKeywordInfo,
						},
						{
							FoodGroup: wire.Locate,
							SubGroup:  wire.LocateUserInfoQuery2,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceServiceRequest,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceClientOnline,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceRateParamsQuery,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceRateParamsSubAdd,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceUserInfoQuery,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceIdleNotification,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceClientVersions,
						},
						{
							FoodGroup: wire.OService,
							SubGroup:  wire.OServiceSetUserInfoFields,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, want, have)
}

func TestOServiceServiceForBOS_OServiceHostOnline(t *testing.T) {
	svc := NewOServiceServiceForBOS(*NewOServiceService(config.Config{}, nil, nil), nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.Alert,
				wire.Buddy,
				wire.ChatNav,
				wire.Feedbag,
				wire.ICBM,
				wire.Locate,
				wire.OService,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceServiceForChat_OServiceHostOnline(t *testing.T) {
	svc := NewOServiceServiceForChat(*NewOServiceService(config.Config{}, nil, nil), nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostOnline,
		},
		Body: wire.SNAC_0x01_0x03_OServiceHostOnline{
			FoodGroups: []uint16{
				wire.OService,
				wire.Chat,
			},
		},
	}

	have := svc.HostOnline()
	assert.Equal(t, want, have)
}

func TestOServiceService_ClientVersions(t *testing.T) {
	svc := NewOServiceService(config.Config{}, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceHostVersions,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: []uint16{5, 6, 7, 8},
		},
	}

	have := svc.ClientVersions(nil, wire.SNACFrame{
		RequestID: 1234,
	}, wire.SNAC_0x01_0x17_OServiceClientVersions{
		Versions: []uint16{5, 6, 7, 8},
	})

	assert.Equal(t, want, have)
}

func TestOServiceService_UserInfoQuery(t *testing.T) {
	svc := NewOServiceService(config.Config{}, nil, nil)
	sess := newTestSession("test-user")

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.OService,
			SubGroup:  wire.OServiceUserInfoUpdate,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.TLVUserInfo(),
		},
	}

	have := svc.UserInfoQuery(nil, sess, wire.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}

func TestOServiceService_IdleNotification(t *testing.T) {
	tests := []struct {
		name   string
		sess   *state.Session
		bodyIn wire.SNAC_0x01_0x11_OServiceIdleNotification
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientBuddies is a list of the recipient's buddies that get
		// updated warning level
		recipientBuddies []string
		broadcastMessage wire.SNACMessage
		wantErr          error
	}{
		{
			name: "set idle from active",
			sess: newTestSession("test-user"),
			bodyIn: wire.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 90,
			},
			recipientScreenName: "test-user",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyArrived,
				},
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("test-user", sessOptIdle(90*time.Second)).TLVUserInfo(),
				},
			},
		},
		{
			name: "set active from idle",
			sess: newTestSession("test-user", sessOptIdle(90*time.Second)),
			bodyIn: wire.SNAC_0x01_0x11_OServiceIdleNotification{
				IdleTime: 0,
			},
			recipientScreenName: "test-user",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyArrived,
				},
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("test-user").TLVUserInfo(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				AdjacentUsers(tt.recipientScreenName).
				Return(tt.recipientBuddies, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RelayToScreenNames(mock.Anything, tt.recipientBuddies, tt.broadcastMessage).
				Maybe()

			svc := NewOServiceService(config.Config{}, messageRelayer, feedbagManager)

			haveErr := svc.IdleNotification(nil, tt.sess, tt.bodyIn)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForBOS_ClientOnline(t *testing.T) {
	type buddiesLookupParams []struct {
		screenName string
		buddies    []string
	}

	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the session of the arriving user
		sess *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline
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
			bodyIn: wire.SNAC_0x01_0x02_OServiceClientOnline{},
			interestedUsersParams: interestedUsersParams{
				{
					screenName: "test-user",
					users:      []string{"buddy1", "buddy2", "buddy3", "buddy4"},
				},
			},
			broadcastToScreenNamesParams: broadcastToScreenNamesParams{
				{
					screenNames: []string{"buddy1", "buddy2", "buddy3", "buddy4"},
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyArrived,
						},
						Body: wire.SNAC_0x03_0x0B_BuddyArrived{
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
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyArrived,
						},
						Body: wire.SNAC_0x03_0x0B_BuddyArrived{
							TLVUserInfo: newTestSession("buddy1").TLVUserInfo(),
						},
					},
				},
				{
					screenName: "test-user",
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Buddy,
							SubGroup:  wire.BuddyArrived,
						},
						Body: wire.SNAC_0x03_0x0B_BuddyArrived{
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
					AdjacentUsers(params.screenName).
					Return(params.users, nil)
			}
			for _, params := range tt.broadcastToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			for _, params := range tt.buddyLookupParams {
				feedbagManager.EXPECT().
					Buddies(params.screenName).
					Return(params.buddies, nil)
			}
			for _, params := range tt.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tt.sendToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := NewOServiceServiceForBOS(OServiceService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}, nil)

			haveErr := svc.ClientOnline(nil, tt.bodyIn, tt.sess)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestOServiceServiceForChat_ClientOnline(t *testing.T) {
	chatter1 := newTestSession("chatter-1", sessOptChatRoomCookie("the-cookie"))
	chatter2 := newTestSession("chatter-2", sessOptChatRoomCookie("the-cookie"))
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
		message wire.SNACMessage
	}
	type sendToScreenNameParams []struct {
		screenName string
		message    wire.SNACMessage
	}

	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the user joining the chat room
		joiningChatter *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline
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
			bodyIn:         wire.SNAC_0x01_0x02_OServiceClientOnline{},
			broadcastExcept: broadcastExcept{
				{
					sess: chatter1,
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Chat,
							SubGroup:  wire.ChatUsersJoined,
						},
						Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
							Users: []wire.TLVUserInfo{
								chatter1.TLVUserInfo(),
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
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Chat,
							SubGroup:  wire.ChatRoomInfoUpdate,
						},
						Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
							Exchange:       chatRoom.Exchange,
							Cookie:         chatRoom.Cookie,
							InstanceNumber: chatRoom.InstanceNumber,
							DetailLevel:    chatRoom.DetailLevel,
							TLVBlock: wire.TLVBlock{
								TLVList: chatRoom.TLVList(),
							},
						},
					},
				},
				{
					screenName: chatter1.ScreenName(),
					message: wire.SNACMessage{
						Frame: wire.SNACFrame{
							FoodGroup: wire.Chat,
							SubGroup:  wire.ChatUsersJoined,
						},
						Body: wire.SNAC_0x0E_0x03_ChatUsersJoined{
							Users: []wire.TLVUserInfo{
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
					RelayToAllExcept(mock.Anything, params.sess, params.message).
					Maybe()
			}
			chatMessageRelayer.EXPECT().
				AllSessions().
				Return(tt.participantsParams).
				Maybe()
			for _, params := range tt.sendToScreenNameParams {
				chatMessageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message).
					Maybe()
			}

			chatRegistry := state.NewChatRegistry()
			chatRegistry.Register(chatRoom, chatMessageRelayer)

			svc := NewOServiceServiceForChat(OServiceService{
				feedbagManager: feedbagManager,
				messageRelayer: chatMessageRelayer,
			}, chatRegistry)

			haveErr := svc.ClientOnline(nil, tt.joiningChatter)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}
