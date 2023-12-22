package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
)

func TestLocateService_UserInfoQuery2Handler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// userSession is the session of the user requesting user info
		userSession *state.Session
		// inputSNAC is the SNAC sent from client to server
		inputSNAC oscar.SNACMessage
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name: "request user info, expect user info response",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "requested-user",
							result:      state.BlockedNo,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "requested-user",
							sess: newTestSession("requested-user",
								sessOptCannedSignonTime,
								sessOptCannedAwayMessage),
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					Type2:      0,
					ScreenName: "requested-user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage).
						TLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{},
				},
			},
		},
		{
			name: "request user info + profile, expect user info response + profile",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "requested-user",
							result:      state.BlockedNo,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "requested-user",
							sess: newTestSession("requested-user",
								sessOptCannedSignonTime,
								sessOptCannedAwayMessage),
						},
					},
				},
				profileManagerParams: profileManagerParams{
					retrieveProfileParams: retrieveProfileParams{
						{
							screenName: "requested-user",
							result:     "this is my profile!",
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					// 2048 is a dummy to make sure bitmask check works
					Type2:      oscar.LocateType2Sig | 2048,
					ScreenName: "requested-user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage).
						TLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
							oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, "this is my profile!"),
						},
					},
				},
			},
		},
		{
			name: "request user info + profile, expect user info response + profile",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "requested-user",
							result:      state.BlockedNo,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "requested-user",
							sess: newTestSession("requested-user",
								sessOptCannedSignonTime,
								sessOptCannedAwayMessage),
						},
					},
				},
				profileManagerParams: profileManagerParams{
					retrieveProfileParams: retrieveProfileParams{
						{
							screenName: "requested-user",
							result:     "this is my profile!",
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					// 2048 is a dummy to make sure bitmask check works
					Type2:      oscar.LocateType2Sig | 2048,
					ScreenName: "requested-user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage).
						TLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.LocateTLVTagsInfoSigMime, `text/aolrtf; charset="us-ascii"`),
							oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, "this is my profile!"),
						},
					},
				},
			},
		},
		{
			name: "request user info + away message, expect user info response + away message",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "requested-user",
							result:      state.BlockedNo,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "requested-user",
							sess: newTestSession("requested-user",
								sessOptCannedSignonTime,
								sessOptCannedAwayMessage),
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					// 2048 is a dummy to make sure bitmask check works
					Type2:      oscar.LocateType2Unavailable | 2048,
					ScreenName: "requested-user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage).
						TLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableMime, `text/aolrtf; charset="us-ascii"`),
							oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableData, "this is my away message!"),
						},
					},
				},
			},
		},
		{
			name: "request user info of user who blocked requester, expect not logged in error",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "requested-user",
							result:      state.BlockedB,
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					ScreenName: "requested-user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name: "request user info of user who does not exist, expect not logged in error",
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					blockedStateParams: blockedStateParams{
						{
							screenName1: "user_screen_name",
							screenName2: "non_existent_requested_user",
							result:      state.BlockedNo,
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "non_existent_requested_user",
							sess:       nil,
						},
					},
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					ScreenName: "non_existent_requested_user",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.blockedStateParams {
				feedbagManager.EXPECT().
					BlockedState(params.screenName1, params.screenName2).
					Return(params.result, nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, val := range tc.mockParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(val.screenName).
					Return(val.sess)
			}
			profileManager := newMockProfileManager(t)
			for _, val := range tc.mockParams.retrieveProfileParams {
				profileManager.EXPECT().
					Profile(val.screenName).
					Return(val.result, val.err)
			}
			svc := NewLocateService(messageRelayer, feedbagManager, profileManager)
			outputSNAC, err := svc.UserInfoQuery2Handler(context.Background(), tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x02_0x15_LocateUserInfoQuery2))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestLocateService_SetKeywordInfoHandler(t *testing.T) {
	svc := NewLocateService(nil, nil, nil)

	outputSNAC := svc.SetKeywordInfoHandler(nil, oscar.SNACFrame{RequestID: 1234})
	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetKeywordReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
			Unknown: 1,
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}

func TestLocateService_SetDirInfoHandler(t *testing.T) {
	svc := NewLocateService(nil, nil, nil)

	outputSNAC := svc.SetDirInfoHandler(nil, oscar.SNACFrame{RequestID: 1234})
	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateSetDirReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
			Result: 1,
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}

func TestLocateService_SetInfoHandler(t *testing.T) {
	tests := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user setting info
		userSession *state.Session
		// inBody is the message sent from client to server
		inBody oscar.SNAC_0x02_0x04_LocateSetInfo
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		wantErr    error
	}{
		{
			name:        "set profile",
			userSession: newTestSession("test-user"),
			inBody: oscar.SNAC_0x02_0x04_LocateSetInfo{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.LocateTLVTagsInfoSigData, "profile-result"),
					},
				},
			},
		},
		{
			name:        "set away message",
			userSession: newTestSession("user_screen_name"),
			inBody: oscar.SNAC_0x02_0x04_LocateSetInfo{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(oscar.LocateTLVTagsInfoUnavailableData, "this is my away message!"),
					},
				},
			},
			mockParams: mockParams{
				messageRelayerParams: messageRelayerParams{
					broadcastToScreenNamesParams: broadcastToScreenNamesParams{
						{
							screenNames: []string{"friend1", "friend2"},
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptAwayMessage("this is my away message!")).TLVUserInfo(),
								},
							},
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					interestedUsersParams: interestedUsersParams{
						{
							screenName: "user_screen_name",
							users:      []string{"friend1", "friend2"},
						},
					},
				},
				profileManagerParams: profileManagerParams{
					upsertProfileParams: upsertProfileParams{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.broadcastToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.interestedUsersParams {
				feedbagManager.EXPECT().
					InterestedUsers(params.screenName).
					Return(params.users, nil)
			}
			profileManager := newMockProfileManager(t)
			if msg, hasProf := tt.inBody.String(oscar.LocateTLVTagsInfoSigData); hasProf {
				profileManager.EXPECT().
					UpsertProfile(tt.userSession.ScreenName(), msg).
					Return(nil)
			}
			svc := NewLocateService(messageRelayer, feedbagManager, profileManager)
			assert.Equal(t, tt.wantErr, svc.SetInfoHandler(nil, tt.userSession, tt.inBody))
		})
	}
}

func TestLocateService_RightsQueryHandler(t *testing.T) {
	svc := NewLocateService(nil, nil, nil)

	outputSNAC := svc.RightsQueryHandler(nil, oscar.SNACFrame{RequestID: 1234})
	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Locate,
			SubGroup:  oscar.LocateRightsReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x02_0x03_LocateRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxSigLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxCapabilitiesLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxFindByEmailList, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxCertsLen, uint16(1000)),
					oscar.NewTLV(oscar.LocateTLVTagsRightsMaxMaxShortCapabilities, uint16(1000)),
				},
			},
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}
