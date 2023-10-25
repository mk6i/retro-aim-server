package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC    oscar.SNAC_0x02_0x15_LocateUserInfoQuery2
		expectOutput XMessage
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				Type2:      0,
				ScreenName: "requested-user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}).GetTLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{},
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}).GetTLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.LocateTLVTagsInfoSigMime,
								Val:   `text/aolrtf; charset="us-ascii"`,
							},
							{
								TType: oscar.LocateTLVTagsInfoSigData,
								Val:   "this is my profile!",
							},
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Sig | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}).GetTLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.LocateTLVTagsInfoSigMime,
								Val:   `text/aolrtf; charset="us-ascii"`,
							},
							{
								TType: oscar.LocateTLVTagsInfoSigData,
								Val:   "this is my profile!",
							},
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				// 2048 is a dummy to make sure bitmask check works
				Type2:      oscar.LocateType2Unavailable | 2048,
				ScreenName: "requested-user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: newTestSession(Session{
						ScreenName:  "requested-user",
						AwayMessage: "this is my away message!",
					}).GetTLVUserInfo(),
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.LocateTLVTagsInfoUnavailableMime,
								Val:   `text/aolrtf; charset="us-ascii"`,
							},
							{
								TType: oscar.LocateTLVTagsInfoUnavailableData,
								Val:   "this is my away message!",
							},
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
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "requested-user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateErr,
				},
				snacOut: oscar.SnacError{
					Code: ErrorCodeNotLoggedOn,
				},
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
					err: ErrSessNotFound,
				},
			},
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
				ScreenName: "non_existent_requested_user",
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateErr,
				},
				snacOut: oscar.SnacError{
					Code: ErrorCodeNotLoggedOn,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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

			svc := LocateService{}
			outputSNAC, err := svc.UserInfoQuery2Handler(tc.userSession, sm, fm, pm, tc.inputSNAC)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestLocateRouter_RouteLocate(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input XMessage
		// output is the response payload
		output XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive LocateRightsQuery, return LocateRightsReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateRightsQuery,
				},
				snacOut: struct{}{},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateRightsReply,
				},
				snacOut: oscar.SNAC_0x02_0x03_LocateRightsReply{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   uint16(1000),
							},
						},
					},
				},
			},
		},
		{
			name: "receive LocateSetInfo, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateSetInfo,
				},
				snacOut: oscar.SNAC_0x02_0x04_LocateSetInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{},
			},
		},
		{
			name: "receive LocateSetDirInfo, return LocateSetDirReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateSetDirInfo,
				},
				snacOut: oscar.SNAC_0x02_0x09_LocateSetDirInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateSetDirReply,
				},
				snacOut: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
					Result: 1,
				},
			},
		},
		{
			name: "receive LocateGetDirInfo, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateGetDirInfo,
				},
				snacOut: oscar.SNAC_0x02_0x0B_LocateGetDirInfo{
					WatcherScreenNames: "screen-name",
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{},
			},
		},
		{
			name: "receive LocateSetKeywordInfo, return LocateSetKeywordReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateSetKeywordInfo,
				},
				snacOut: oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateSetKeywordReply,
				},
				snacOut: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
					Unknown: 1,
				},
			},
		},
		{
			name: "receive LocateUserInfoQuery2, return LocateUserInfoReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoQuery2,
				},
				snacOut: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					Type2: 1,
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				snacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
					LocateInfo: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
		},
		{
			name: "receive LocateGetKeywordInfo, expect ErrUnsupportedSubGroup",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: LOCATE,
					SubGroup:  oscar.LocateGetKeywordInfo,
				},
				snacOut: struct{}{}, // empty SNAC
			},
			output:    XMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockLocateHandler(t)
			svc.EXPECT().
				RightsQueryHandler().
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetDirInfoHandler().
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetInfoHandler(mock.Anything, mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.handlerErr).
				Maybe()
			svc.EXPECT().
				SetKeywordInfoHandler().
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQuery2Handler(mock.Anything, mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := LocateRouter{
				LocateHandler: svc,
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.snacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteLocate(nil, nil, nil, tc.input.snacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.snacFrame == (oscar.SnacFrame{}) {
				return // handler doesn't return response
			}

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))
			assert.Equal(t, uint16(1), flap.Sequence)

			// verify the SNAC frame
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, bufOut))
			assert.Equal(t, tc.output.snacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.snacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), bufOut.Bytes())
		})
	}
}
