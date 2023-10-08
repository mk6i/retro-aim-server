package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// newTestSession returns a copy of the prototype session with fixed time
//
//goland:noinspection GoVetCopyLock
func newTestSession(prototype Session) *Session {
	prototype.SignonTime = time.UnixMilli(1696790127565)
	return &prototype
}

func TestSendAndReceiveUserInfoQuery2(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState BlockedState
		// screenNameLookups is the list of user session lookups
		screenNameLookups map[string]struct {
			sess *Session
			err  error
		}
		// screenNameLookups is the list of user session lookups
		profileLookups map[string]struct {
			payload string
			err     error
		}
		// userSession is the session of the user requesting the user info
		userSession *Session
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x02_0x15_LocateUserInfoQuery2
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
	}{
		{
			name:         "request user info, expect user info response",
			blockedState: BlockedNo,
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}),
				},
			},
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateUserInfoReply,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				Type2:      0,
				ScreenName: "requested-user",
			},
			expectSNACBody: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
				TLVUserInfo: newTestSession(Session{
					ScreenName:  "requested-user",
					AwayMessage: "this is my away message!",
				}).GetTLVUserInfo(),
				LocateInfo: oscar.TLVRestBlock{},
			},
		},
		{
			name:         "request user info + profile, expect user info response + profile",
			blockedState: BlockedNo,
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}),
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
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateUserInfoReply,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectSNACBody: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
				TLVUserInfo: newTestSession(Session{
					ScreenName:  "requested-user",
					AwayMessage: "this is my away message!",
				}).GetTLVUserInfo(),
				LocateInfo: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: LocateTLVTagsInfoSigMime,
							Val:   `text/aolrtf; charset="us-ascii"`,
						},
						{
							TType: LocateTLVTagsInfoSigData,
							Val:   "this is my profile!",
						},
					},
				},
			},
		},
		{
			name:         "request user info + profile, expect user info response + profile",
			blockedState: BlockedNo,
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}),
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
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateUserInfoReply,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectSNACBody: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
				TLVUserInfo: newTestSession(Session{
					ScreenName:  "requested-user",
					AwayMessage: "this is my away message!",
				}).GetTLVUserInfo(),
				LocateInfo: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: LocateTLVTagsInfoSigMime,
							Val:   `text/aolrtf; charset="us-ascii"`,
						},
						{
							TType: LocateTLVTagsInfoSigData,
							Val:   "this is my profile!",
						},
					},
				},
			},
		},
		{
			name:         "request user info + away message, expect user info response + away message",
			blockedState: BlockedNo,
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"requested-user": {
					sess: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}),
				},
			},
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateUserInfoReply,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Unavailable | 2048,
				ScreenName: "requested-user",
			},
			expectSNACBody: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
				TLVUserInfo: newTestSession(Session{
					ScreenName:  "requested-user",
					AwayMessage: "this is my away message!",
				}).GetTLVUserInfo(),
				LocateInfo: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: LocateTLVTagsInfoUnavailableMime,
							Val:   `text/aolrtf; charset="us-ascii"`,
						},
						{
							TType: LocateTLVTagsInfoUnavailableData,
							Val:   "this is my away message!",
						},
					},
				},
			},
		},
		{
			name:         "request user info of user who blocked requester, expect not logged in error",
			blockedState: BlockedB,
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateErr,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "requested-user",
			},
			expectSNACBody: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			},
		},
		{
			name:         "request user info of user who does not exist, expect not logged in error",
			blockedState: BlockedNo,
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"non_existent_requested_user": {
					err: errSessNotFound,
				},
			},
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateErr,
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "non_existent_requested_user",
			},
			expectSNACBody: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			fm := NewMockFeedbagManager(t)
			fm.EXPECT().
				Blocked(tc.userSession.ScreenName, tc.inputSNAC.ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			sm := NewMockSessionManager(t)
			for screenName, val := range tc.screenNameLookups {
				sm.EXPECT().
					RetrieveByScreenName(screenName).
					Return(val.sess, val.err).
					Maybe()
			}
			pm := NewMockProfileManager(t)
			for screenName, val := range tc.profileLookups {
				pm.EXPECT().
					RetrieveProfile(screenName).
					Return(val.payload, val.err).
					Maybe()
			}
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			snac := oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateUserInfoQuery2,
			}
			assert.NoError(t, SendAndReceiveUserInfoQuery2(tc.userSession, sm, fm, pm, snac, input, output, &seq))
			//
			// verify output
			//
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, output))
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, output))
			assert.Equal(t, tc.expectSNACFrame, snacFrame)
			//
			// verify output SNAC body
			//
			switch v := tc.expectSNACBody.(type) {
			case oscar.SNAC_0x02_0x06_LocateUserInfoReply:
				assert.NoError(t, v.SerializeInPlace())
				assert.NoError(t, v.LocateInfo.SerializeInPlace())
				outputSNAC := oscar.SNAC_0x02_0x06_LocateUserInfoReply{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			case oscar.SnacError:
				outputSNAC := oscar.SnacError{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
		})
	}
}
