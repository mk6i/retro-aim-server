package foodgroup

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestBuddyService_RightsQuery(t *testing.T) {
	svc := NewBuddyService(nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
				},
			},
		},
	}
	have := svc.RightsQuery(nil, wire.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}

func TestBuddyService_AddBuddies(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// sess is the client session
		sess *state.Session
		// bodyIn is the input SNAC
		bodyIn wire.SNAC_0x03_0x04_BuddyAddBuddies
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantErr is the expected error
		wantErr error
	}{
		{
			name: "add 2 buddies, sign-on complete",
			sess: newTestSession("user_screen_name", sessOptSignonComplete),
			bodyIn: wire.SNAC_0x03_0x04_BuddyAddBuddies{
				Buddies: []struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: state.NewIdentScreenName("buddy_1_online"),
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: state.NewIdentScreenName("buddy_2_offline"),
							sess:       nil,
						},
					},
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_1_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: state.NewIdentScreenName("buddy_1_online"),
						},
					},
				},
			},
		},
		{
			name: "add 2 buddies, sign-on not complete",
			sess: newTestSession("user_screen_name"),
			bodyIn: wire.SNAC_0x03_0x04_BuddyAddBuddies{
				Buddies: []struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tt.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}

			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.addBuddyParams {
				legacyBuddyListManager.EXPECT().
					AddBuddy(params.userScreenName, params.buddyScreenName)
			}

			svc := NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager)

			haveErr := svc.AddBuddies(nil, tt.sess, tt.bodyIn)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}

func TestBuddyService_DelBuddies(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// sess is the client session
		sess *state.Session
		// bodyIn is the input SNAC
		bodyIn wire.SNAC_0x03_0x05_BuddyDelBuddies
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// wantErr is the expected error
		wantErr error
	}{
		{
			name: "delete 2 buddies",
			sess: newTestSession("user_screen_name", sessOptSignonComplete),
			bodyIn: wire.SNAC_0x03_0x05_BuddyDelBuddies{
				Buddies: []struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{
					{
						ScreenName: "buddy_1_online",
					},
					{
						ScreenName: "buddy_2_offline",
					},
				},
			},
			mockParams: mockParams{
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					deleteBuddyParams: deleteBuddyParams{
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							userScreenName:  state.NewIdentScreenName("user_screen_name"),
							buddyScreenName: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tt.mockParams.deleteBuddyParams {
				legacyBuddyListManager.EXPECT().
					DeleteBuddy(params.userScreenName, params.buddyScreenName)
			}

			svc := NewBuddyService(nil, nil, legacyBuddyListManager)

			svc.DelBuddies(nil, tt.sess, tt.bodyIn)
		})
	}
}

func TestBuddyService_BroadcastBuddyArrived(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// sourceSession is the session of the user
		userSession *state.Session
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "send buddy arrival notification to users who have user_screen_name on their server-side or client-side buddy list",
			userSession: newTestSession("user_screen_name"),
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					adjacentUsersParams: adjacentUsersParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							users:      []state.IdentScreenName{state.NewIdentScreenName("friend1")},
						},
					},
					feedbagParams: feedbagParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							results:    []wire.FeedbagItem{},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []state.IdentScreenName{
								state.NewIdentScreenName("friend1"),
								state.NewIdentScreenName("friend2"),
							},
							message: wire.SNACMessage{
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
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					whoAddedUserParams: whoAddedUserParams{
						{
							userScreenName: state.NewIdentScreenName("user_screen_name"),
							result:         []state.IdentScreenName{state.NewIdentScreenName("friend2")},
						},
					},
				},
			},
		},
		{
			name:        "send buddy arrival notification containing buddy icon to user who has user_screen_name on their server-side buddy list",
			userSession: newTestSession("user_screen_name"),
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					adjacentUsersParams: adjacentUsersParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							users:      []state.IdentScreenName{state.NewIdentScreenName("friend1")},
						},
					},
					feedbagParams: feedbagParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							results: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBuddy,
									Name:    "friend10",
								},
								{
									ClassID: wire.FeedbagClassIdBart,
									Name:    strconv.Itoa(int(wire.BARTTypesBadgeUrl)),
								},
								{
									ClassID: wire.FeedbagClassIdBart,
									Name:    strconv.Itoa(int(wire.BARTTypesBuddyIcon)),
									TLVLBlock: wire.TLVLBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
												Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											}),
										},
									},
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []state.IdentScreenName{state.NewIdentScreenName("friend1")},
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: userInfoWithBARTIcon(
										newTestSession("user_screen_name"),
										wire.BARTID{
											Type: wire.BARTTypesBuddyIcon,
											BARTInfo: wire.BARTInfo{
												Flags: wire.BARTFlagsKnown,
												Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											},
										},
									),
								},
							},
						},
					},
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					whoAddedUserParams: whoAddedUserParams{
						{
							userScreenName: state.NewIdentScreenName("user_screen_name"),
							result:         []state.IdentScreenName{},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.adjacentUsersParams {
				feedbagManager.EXPECT().
					AdjacentUsers(params.screenName).
					Return(params.users, params.err)
			}
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tc.mockParams.whoAddedUserParams {
				legacyBuddyListManager.EXPECT().
					WhoAddedUser(params.userScreenName).
					Return(params.result)
			}

			svc := NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager)

			err := svc.BroadcastBuddyArrived(nil, tc.userSession)
			assert.NoError(t, err)
		})
	}
}

func TestBuddyService_BroadcastDeparture(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// sourceSession is the session of the user
		userSession *state.Session
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "send buddy departed notification to users who have user_screen_name on their server-side or client-side buddy list",
			userSession: newTestSession("user_screen_name"),
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					adjacentUsersParams: adjacentUsersParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							users:      []state.IdentScreenName{state.NewIdentScreenName("friend1")},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []state.IdentScreenName{
								state.NewIdentScreenName("friend1"),
								state.NewIdentScreenName("friend2"),
							},
							message: wire.SNACMessage{
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
				},
				legacyBuddyListManagerParams: legacyBuddyListManagerParams{
					whoAddedUserParams: whoAddedUserParams{
						{
							userScreenName: state.NewIdentScreenName("user_screen_name"),
							result:         []state.IdentScreenName{state.NewIdentScreenName("friend2")},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.adjacentUsersParams {
				feedbagManager.EXPECT().
					AdjacentUsers(params.screenName).
					Return(params.users, params.err)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tc.mockParams.whoAddedUserParams {
				legacyBuddyListManager.EXPECT().
					WhoAddedUser(params.userScreenName).
					Return(params.result)
			}

			svc := NewBuddyService(messageRelayer, feedbagManager, legacyBuddyListManager)

			err := svc.BroadcastBuddyDeparted(nil, tc.userSession)
			assert.NoError(t, err)
		})
	}
}

func TestBuddyService_UnicastBuddyDeparted(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// sourceSession is the session of the user
		sourceSession *state.Session
		// destSession is the session of the user receiving the notification
		destSession *state.Session
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:          "send buddy departed notification to user",
			sourceSession: newTestSession("src_screen_name"),
			destSession:   newTestSession("dest_screen_name"),
			mockParams: mockParams{
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("dest_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyDeparted,
								},
								Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
									TLVUserInfo: wire.TLVUserInfo{
										ScreenName:   "src_screen_name",
										WarningLevel: 0,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := NewBuddyService(messageRelayer, nil, nil)

			svc.UnicastBuddyDeparted(nil, tc.sourceSession, tc.destSession)
		})
	}
}

func TestBuddyService_UnicastBuddyArrived(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// sourceSession is the session of the user
		sourceSession *state.Session
		// destSession is the session of the user receiving the notification
		destSession *state.Session
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:          "send buddy arrival notification to user",
			sourceSession: newTestSession("src_screen_name"),
			destSession:   newTestSession("dest_screen_name"),
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: state.NewIdentScreenName("src_screen_name"),
							results:    []wire.FeedbagItem{},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("dest_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("src_screen_name").TLVUserInfo(),
								},
							},
						},
					},
				},
			},
		},
		{
			name:          "send buddy arrival notification containing buddy icon to user",
			sourceSession: newTestSession("src_screen_name"),
			destSession:   newTestSession("dest_screen_name"),
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: state.NewIdentScreenName("src_screen_name"),
							results: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBuddy,
									Name:    "friend10",
								},
								{
									ClassID: wire.FeedbagClassIdBart,
									Name:    strconv.Itoa(int(wire.BARTTypesBadgeUrl)),
								},
								{
									ClassID: wire.FeedbagClassIdBart,
									Name:    strconv.Itoa(int(wire.BARTTypesBuddyIcon)),
									TLVLBlock: wire.TLVLBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
												Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											}),
										},
									},
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("dest_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: userInfoWithBARTIcon(
										newTestSession("src_screen_name"),
										wire.BARTID{
											Type: wire.BARTTypesBuddyIcon,
											BARTInfo: wire.BARTInfo{
												Flags: wire.BARTFlagsKnown,
												Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											},
										},
									),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := NewBuddyService(messageRelayer, feedbagManager, nil)

			err := svc.UnicastBuddyArrived(nil, tc.sourceSession, tc.destSession)
			assert.NoError(t, err)
		})
	}
}
