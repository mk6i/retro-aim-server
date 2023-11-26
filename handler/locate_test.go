package handler

import (
	"context"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
)

func TestSendAndReceiveUserInfoQuery2(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// screenNameLookups is the list of user session lookups
		screenNameLookups map[string]struct {
			sess *state.Session
			err  error
		}
		// screenNameLookups is the list of user session lookups
		profileLookups map[string]struct {
			payload string
			err     error
		}
		// userSession is the session of the user requesting the user info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// expectOutput is the SNAC sent from the server to the
		// recipient client
		expectOutput oscar.SNACMessage
	}{
		{
			name:         "request user info, expect user info response",
			blockedState: state.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
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
			name:         "request user info + profile, expect user info response + profile",
			blockedState: state.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
				},
			},
			profileLookups: map[string]struct {
				payload string
				err     error
			}{
				"requested-user": {
					payload: "this is my profile!",
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
			name:         "request user info + profile, expect user info response + profile",
			blockedState: state.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
				},
			},
			profileLookups: map[string]struct {
				payload string
				err     error
			}{
				"requested-user": {
					payload: "this is my profile!",
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
			name:         "request user info + away message, expect user info response + away message",
			blockedState: state.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
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
			name:         "request user info of user who blocked requester, expect not logged in error",
			blockedState: state.BlockedB,
			userSession:  newTestSession("user_screen_name"),
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
			name:         "request user info of user who does not exist, expect not logged in error",
			blockedState: state.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"non_existent_requested_user": {
					sess: nil,
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
			feedbagManager.EXPECT().
				Blocked(tc.userSession.ScreenName(),
					tc.inputSNAC.Body.(oscar.SNAC_0x02_0x15_LocateUserInfoQuery2).ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			for screenName, val := range tc.screenNameLookups {
				messageRelayer.EXPECT().
					RetrieveByScreenName(screenName).
					Return(val.sess).
					Maybe()
			}
			profileManager := newMockProfileManager(t)
			for screenName, val := range tc.profileLookups {
				profileManager.EXPECT().
					RetrieveProfile(screenName).
					Return(val.payload, val.err).
					Maybe()
			}
			svc := LocateService{
				sessionManager: messageRelayer,
				feedbagManager: feedbagManager,
				profileManager: profileManager,
			}
			outputSNAC, err := svc.UserInfoQuery2Handler(context.Background(), tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x02_0x15_LocateUserInfoQuery2))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
