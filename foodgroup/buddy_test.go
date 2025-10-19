package foodgroup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestBuddyService_RightsQuery(t *testing.T) {
	svc := NewBuddyService(nil, nil, nil, nil, nil)

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					wire.NewTLVBE(wire.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					wire.NewTLVBE(wire.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					wire.NewTLVBE(wire.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
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
				clientSideBuddyListManagerParams: clientSideBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastVisibilityParams: broadcastVisibilityParams{
						{
							from: state.NewIdentScreenName("user_screen_name"),
							filter: []state.IdentScreenName{
								state.NewIdentScreenName("buddy_1_online"),
								state.NewIdentScreenName("buddy_2_offline"),
							},
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
				clientSideBuddyListManagerParams: clientSideBuddyListManagerParams{
					addBuddyParams: addBuddyParams{
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientSideBuddyListManager := newMockClientSideBuddyListManager(t)
			for _, params := range tt.mockParams.addBuddyParams {
				clientSideBuddyListManager.EXPECT().
					AddBuddy(matchContext(), params.me, params.them).
					Return(params.err)
			}
			mockBuddyBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastVisibilityParams {
				mockBuddyBroadcaster.EXPECT().
					BroadcastVisibility(mock.Anything, matchSession(params.from), params.filter, true).
					Return(params.err)
			}

			svc := BuddyService{
				clientSideBuddyListManager: clientSideBuddyListManager,
				buddyBroadcaster:           mockBuddyBroadcaster,
			}

			haveErr := svc.AddBuddies(context.Background(), tt.sess, tt.bodyIn)
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
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastVisibilityParams: broadcastVisibilityParams{
						{
							from: state.NewIdentScreenName("user_screen_name"),
							filter: []state.IdentScreenName{
								state.NewIdentScreenName("buddy_1_online"),
								state.NewIdentScreenName("buddy_2_offline"),
							},
						},
					},
				},
				clientSideBuddyListManagerParams: clientSideBuddyListManagerParams{
					deleteBuddyParams: deleteBuddyParams{
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_1_online"),
						},
						{
							me:   state.NewIdentScreenName("user_screen_name"),
							them: state.NewIdentScreenName("buddy_2_offline"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBuddyBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tt.mockParams.broadcastVisibilityParams {
				mockBuddyBroadcaster.EXPECT().
					BroadcastVisibility(mock.Anything, matchSession(params.from), params.filter, true).
					Return(params.err)
			}
			clientSideBuddyListManager := newMockClientSideBuddyListManager(t)
			for _, params := range tt.mockParams.deleteBuddyParams {
				clientSideBuddyListManager.EXPECT().
					RemoveBuddy(matchContext(), params.me, params.them).
					Return(params.err)
			}

			svc := BuddyService{
				buddyBroadcaster:           mockBuddyBroadcaster,
				clientSideBuddyListManager: clientSideBuddyListManager,
			}

			assert.ErrorIs(t, tt.wantErr, svc.DelBuddies(context.Background(), tt.sess, tt.bodyIn))
		})
	}
}

func TestBuddyNotifier_BroadcastBuddyArrived(t *testing.T) {
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
			name:        "happy path",
			userSession: newTestSession("me"),
			mockParams: mockParams{
				bartItemManagerParams: bartItemManagerParams{
					buddyIconMetadataParams: buddyIconMetadataParams{
						{
							screenName: state.NewIdentScreenName("me"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: wire.BARTFlagsKnown,
									Hash:  []byte{'m', 'y', 'i', 'c', 'o', 'n'},
								},
							},
						},
					},
				},
				relationshipFetcherParams: relationshipFetcherParams{
					allRelationshipsParams: allRelationshipsParams{
						{
							screenName: state.NewIdentScreenName("me"),
							filter:     nil,
							result: []state.Relationship{
								{
									User:          state.NewIdentScreenName("friend1-visible"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend2-visible"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend3-you-block"),
									BlocksYou:     false,
									YouBlock:      true,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend4-blocks-you"),
									BlocksYou:     true,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend5-not-on-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: false,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []state.IdentScreenName{
								state.NewIdentScreenName("friend1-visible"),
								state.NewIdentScreenName("friend2-visible"),
							},
							message: newBuddyArrivedNotif(userInfoWithBARTIcon(
								newTestSession("me"),
								wire.BARTID{
									Type: wire.BARTTypesBuddyIcon,
									BARTInfo: wire.BARTInfo{
										Flags: wire.BARTFlagsKnown,
										Hash:  []byte{'m', 'y', 'i', 'c', 'o', 'n'},
									},
								},
							)),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			relationshipFetcher := newMockRelationshipFetcher(t)
			for _, params := range tc.mockParams.allRelationshipsParams {
				relationshipFetcher.EXPECT().
					AllRelationships(matchContext(), params.screenName, params.filter).
					Return(params.result, params.err)
			}
			bartItemManager := newMockBARTItemManager(t)
			for _, params := range tc.mockParams.buddyIconMetadataParams {
				bartItemManager.EXPECT().
					BuddyIconMetadata(matchContext(), params.screenName).
					Return(params.result, params.err)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(matchContext(), params.screenNames, params.message)
			}

			svc := buddyNotifier{
				bartItemManager:     bartItemManager,
				relationshipFetcher: relationshipFetcher,
				messageRelayer:      messageRelayer,
			}

			err := svc.BroadcastBuddyArrived(context.Background(), tc.userSession.IdentScreenName(), tc.userSession.TLVUserInfo())
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
			name:        "happy path",
			userSession: newTestSession("me"),
			mockParams: mockParams{
				relationshipFetcherParams: relationshipFetcherParams{
					allRelationshipsParams: allRelationshipsParams{
						{
							screenName: state.NewIdentScreenName("me"),
							filter:     nil,
							result: []state.Relationship{
								{
									User:          state.NewIdentScreenName("friend1-visible"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend2-visible"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend3-you-block"),
									BlocksYou:     false,
									YouBlock:      true,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend4-blocks-you"),
									BlocksYou:     true,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend5-not-on-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: false,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNamesParams: relayToScreenNamesParams{
						{
							screenNames: []state.IdentScreenName{
								state.NewIdentScreenName("friend1-visible"),
								state.NewIdentScreenName("friend2-visible"),
							},
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyDeparted,
									RequestID: wire.ReqIDFromServer,
								},
								Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
									TLVUserInfo: wire.TLVUserInfo{
										ScreenName:   "me",
										WarningLevel: 0,
										TLVBlock: wire.TLVBlock{
											TLVList: wire.TLVList{
												wire.NewTLVBE(wire.OServiceUserInfoUserFlags, uint16(0)),
											},
										},
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
			relationshipFetcher := newMockRelationshipFetcher(t)
			for _, params := range tc.mockParams.allRelationshipsParams {
				relationshipFetcher.EXPECT().
					AllRelationships(matchContext(), params.screenName, params.filter).
					Return(params.result, params.err)
			}
			bartItemManager := newMockBARTItemManager(t)
			for _, params := range tc.mockParams.buddyIconMetadataParams {
				bartItemManager.EXPECT().
					BuddyIconMetadata(matchContext(), params.screenName).
					Return(params.result, params.err)
			}

			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}

			svc := buddyNotifier{
				relationshipFetcher: relationshipFetcher,
				messageRelayer:      messageRelayer,
			}

			err := svc.BroadcastBuddyDeparted(context.Background(), tc.userSession)
			assert.NoError(t, err)
		})
	}
}

func Test_buddyNotifier_BroadcastVisibility(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// sourceSession is the session of the user
		userSession *state.Session
		// filter limits specific users that can be notified
		filter []state.IdentScreenName
		// doSendDepartures indicates whether departure messages should be sent
		doSendDepartures bool
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:        "happy path",
			userSession: newTestSession("me"),
			mockParams: mockParams{
				bartItemManagerParams: bartItemManagerParams{
					buddyIconMetadataParams: buddyIconMetadataParams{
						{
							screenName: state.NewIdentScreenName("me"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("friend3-visible-on-your-list"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							result:     nil,
						},
					},
				},
				relationshipFetcherParams: relationshipFetcherParams{
					allRelationshipsParams: allRelationshipsParams{
						{
							screenName: state.NewIdentScreenName("me"),
							filter:     nil,
							result: []state.Relationship{
								{
									User:          state.NewIdentScreenName("friend1-blocks-you"),
									BlocksYou:     true,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend2-visible-on-their-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  false,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend3-visible-on-your-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: false,
								},
								{
									User:          state.NewIdentScreenName("friend4-visible-on-both-lists"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend5-blocked-on-their-list"),
									BlocksYou:     false,
									YouBlock:      true,
									IsOnYourList:  false,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend6-blocked-on-your-list"),
									BlocksYou:     false,
									YouBlock:      true,
									IsOnYourList:  true,
									IsOnTheirList: false,
								},
								{
									User:          state.NewIdentScreenName("friend7-blocked-on-both-lists"),
									BlocksYou:     false,
									YouBlock:      true,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend7-visible-offline"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("friend2-visible-on-their-list"),
							message:    newBuddyArrivedNotif(newTestSession("me").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyArrivedNotif(newTestSession("friend3-visible-on-your-list").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							message:    newBuddyArrivedNotif(newTestSession("me").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyArrivedNotif(newTestSession("friend4-visible-on-both-lists").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("friend5-blocked-on-their-list"),
							message:    newBuddyDepartedNotif(newTestSession("me")),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyDepartedNotif(newTestSession("friend6-blocked-on-your-list")),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyDepartedNotif(newTestSession("friend7-blocked-on-both-lists")),
						},
						{
							screenName: state.NewIdentScreenName("friend7-blocked-on-both-lists"),
							message:    newBuddyDepartedNotif(newTestSession("me")),
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams: retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("friend2-visible-on-their-list"),
							result:     newTestSession("friend2-visible-on-their-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend3-visible-on-your-list"),
							result:     newTestSession("friend3-visible-on-your-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							result:     newTestSession("friend4-visible-on-both-lists"),
						},
						{
							screenName: state.NewIdentScreenName("friend5-blocked-on-their-list"),
							result:     newTestSession("friend5-blocked-on-their-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend6-blocked-on-your-list"),
							result:     newTestSession("friend6-blocked-on-your-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend7-blocked-on-both-lists"),
							result:     newTestSession("friend7-blocked-on-both-lists"),
						},
						{
							screenName: state.NewIdentScreenName("friend7-visible-offline"),
							result:     nil,
						},
					},
				},
			},
			doSendDepartures: true,
		},
		{
			name:        "don't send departure notifications",
			userSession: newTestSession("me"),
			mockParams: mockParams{
				bartItemManagerParams: bartItemManagerParams{
					buddyIconMetadataParams: buddyIconMetadataParams{
						{
							screenName: state.NewIdentScreenName("me"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("friend3-visible-on-your-list"),
							result:     nil,
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							result:     nil,
						},
					},
				},
				relationshipFetcherParams: relationshipFetcherParams{
					allRelationshipsParams: allRelationshipsParams{
						{
							screenName: state.NewIdentScreenName("me"),
							filter:     nil,
							result: []state.Relationship{
								{
									User:          state.NewIdentScreenName("friend2-visible-on-their-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  false,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend3-visible-on-your-list"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: false,
								},
								{
									User:          state.NewIdentScreenName("friend4-visible-on-both-lists"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
								{
									User:          state.NewIdentScreenName("friend7-visible-offline"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("friend2-visible-on-their-list"),
							message:    newBuddyArrivedNotif(newTestSession("me").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyArrivedNotif(newTestSession("friend3-visible-on-your-list").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							message:    newBuddyArrivedNotif(newTestSession("me").TLVUserInfo()),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message:    newBuddyArrivedNotif(newTestSession("friend4-visible-on-both-lists").TLVUserInfo()),
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams: retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("friend2-visible-on-their-list"),
							result:     newTestSession("friend2-visible-on-their-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend3-visible-on-your-list"),
							result:     newTestSession("friend3-visible-on-your-list"),
						},
						{
							screenName: state.NewIdentScreenName("friend4-visible-on-both-lists"),
							result:     newTestSession("friend4-visible-on-both-lists"),
						},
						{
							screenName: state.NewIdentScreenName("friend7-visible-offline"),
							result:     nil,
						},
					},
				},
			},
			doSendDepartures: false,
		},
		{
			name:        "users have buddy icons",
			userSession: newTestSession("me"),
			mockParams: mockParams{
				bartItemManagerParams: bartItemManagerParams{
					buddyIconMetadataParams: buddyIconMetadataParams{
						{
							screenName: state.NewIdentScreenName("me"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: wire.BARTFlagsKnown,
									Hash:  []byte{'m', 'y', 'i', 'c', 'o', 'n'},
								},
							},
						},
						{
							screenName: state.NewIdentScreenName("friend-visible-on-both-lists"),
							result: &wire.BARTID{
								Type: wire.BARTTypesBuddyIcon,
								BARTInfo: wire.BARTInfo{
									Flags: wire.BARTFlagsKnown,
									Hash:  []byte{'t', 'h', 'e', 'i', 'r', 'i', 'c', 'o', 'n'},
								},
							},
						},
					},
				},
				relationshipFetcherParams: relationshipFetcherParams{
					allRelationshipsParams: allRelationshipsParams{
						{
							screenName: state.NewIdentScreenName("me"),
							filter:     nil,
							result: []state.Relationship{
								{
									User:          state.NewIdentScreenName("friend-visible-on-both-lists"),
									BlocksYou:     false,
									YouBlock:      false,
									IsOnYourList:  true,
									IsOnTheirList: true,
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("friend-visible-on-both-lists"),
							message: newBuddyArrivedNotif(userInfoWithBARTIcon(
								newTestSession("me"),
								wire.BARTID{
									Type: wire.BARTTypesBuddyIcon,
									BARTInfo: wire.BARTInfo{
										Flags: wire.BARTFlagsKnown,
										Hash:  []byte{'m', 'y', 'i', 'c', 'o', 'n'},
									},
								},
							)),
						},
						{
							screenName: state.NewIdentScreenName("me"),
							message: newBuddyArrivedNotif(userInfoWithBARTIcon(
								newTestSession("friend-visible-on-both-lists"),
								wire.BARTID{
									Type: wire.BARTTypesBuddyIcon,
									BARTInfo: wire.BARTInfo{
										Flags: wire.BARTFlagsKnown,
										Hash:  []byte{'t', 'h', 'e', 'i', 'r', 'i', 'c', 'o', 'n'},
									},
								},
							)),
						},
					},
				},
				sessionRetrieverParams: sessionRetrieverParams{
					retrieveSessionParams: retrieveSessionParams{
						{
							screenName: state.NewIdentScreenName("friend-visible-on-both-lists"),
							result:     newTestSession("friend-visible-on-both-lists"),
						},
					},
				},
			},
			doSendDepartures: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			relationshipFetcher := newMockRelationshipFetcher(t)
			for _, params := range tc.mockParams.allRelationshipsParams {
				relationshipFetcher.EXPECT().
					AllRelationships(matchContext(), params.screenName, params.filter).
					Return(params.result, params.err)
			}
			bartItemManager := newMockBARTItemManager(t)
			for _, params := range tc.mockParams.buddyIconMetadataParams {
				bartItemManager.EXPECT().
					BuddyIconMetadata(matchContext(), params.screenName).
					Return(params.result, params.err)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}
			sessionRetriever := newMockSessionRetriever(t)
			for _, params := range tc.mockParams.retrieveSessionParams {
				sessionRetriever.EXPECT().
					RetrieveSession(params.screenName, params.sessionNum).
					Return(params.result)
			}

			svc := buddyNotifier{
				bartItemManager:     bartItemManager,
				relationshipFetcher: relationshipFetcher,
				messageRelayer:      messageRelayer,
				sessionRetriever:    sessionRetriever,
			}

			err := svc.BroadcastVisibility(context.Background(), tc.userSession, tc.filter, tc.doSendDepartures)
			assert.NoError(t, err)
		})
	}
}

func newBuddyDepartedNotif(me *state.Session) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDeparted,
			RequestID: wire.ReqIDFromServer,
		},
		Body: wire.SNAC_0x03_0x0C_BuddyDeparted{
			TLVUserInfo: wire.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   me.IdentScreenName().String(),
				WarningLevel: me.Warning(),
			},
		},
	}
}

func newBuddyArrivedNotif(userInfo wire.TLVUserInfo) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyArrived,
			RequestID: wire.ReqIDFromServer,
		},
		Body: wire.SNAC_0x03_0x0B_BuddyArrived{
			TLVUserInfo: userInfo,
		},
	}
}
