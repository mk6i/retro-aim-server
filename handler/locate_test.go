package handler

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendAndReceiveUserInfoQuery2(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState server.BlockedState
		// screenNameLookups is the list of user session lookups
		screenNameLookups map[string]struct {
			sess *server.Session
			err  error
		}
		// screenNameLookups is the list of user session lookups
		profileLookups map[string]struct {
			payload string
			err     error
		}
		// userSession is the session of the user requesting the user info
		userSession *server.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC    oscar.SNAC_0x02_0x15_LocateUserInfoQuery2
		expectOutput oscar.XMessage
	}{
		{
			name:         "request user info, expect user info response",
			blockedState: server.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *server.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				Type2:      0,
				ScreenName: "requested-user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
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
			blockedState: server.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *server.Session
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
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
			blockedState: server.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *server.Session
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
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
			blockedState: server.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *server.Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession("requested-user",
						sessOptCannedSignonTime,
						sessOptCannedAwayMessage),
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Unavailable | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
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
			blockedState: server.BlockedB,
			userSession:  newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "requested-user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:         "request user info of user who does not exist, expect not logged in error",
			blockedState: server.BlockedNo,
			screenNameLookups: map[string]struct {
				sess *server.Session
				err  error
			}{
				"non_existent_requested_user": {
					sess: nil,
				},
			},
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "non_existent_requested_user",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fm := server.NewMockFeedbagManager(t)
			fm.EXPECT().
				Blocked(tc.userSession.ScreenName(), tc.inputSNAC.ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			sm := server.NewMockSessionManager(t)
			for screenName, val := range tc.screenNameLookups {
				sm.EXPECT().
					RetrieveByScreenName(screenName).
					Return(val.sess).
					Maybe()
			}
			pm := server.NewMockProfileManager(t)
			for screenName, val := range tc.profileLookups {
				pm.EXPECT().
					RetrieveProfile(screenName).
					Return(val.payload, val.err).
					Maybe()
			}

			svc := LocateService{
				sm: sm,
				fm: fm,
				pm: pm,
			}
			outputSNAC, err := svc.UserInfoQuery2Handler(context.Background(), tc.userSession, tc.inputSNAC)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
