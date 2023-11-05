package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestQueryHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// expectOutput is the SNAC payload sent from the server to the
		// recipient client
		expectOutput XMessage
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Items: []oscar.FeedbagItem{},
				},
			},
		},
		{
			name:       "retrieve feedbag with items",
			screenName: "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(1696472198082),
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Version: 0,
					Items: []oscar.FeedbagItem{
						{
							Name: "buddy_1",
						},
						{
							Name: "buddy_2",
						},
					},
					LastUpdate: uint32(time.UnixMilli(1696472198082).Unix()),
				},
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
				Retrieve(tc.screenName).
				Return(tc.feedbagItems, nil).
				Maybe()
			fm.EXPECT().
				LastModified(tc.screenName).
				Return(tc.lastModified, nil).
				Maybe()
			//
			// send input SNAC
			//
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			svc := FeedbagService{}
			outputSNAC, err := svc.QueryHandler(nil, senderSession, fm)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestQueryIfModifiedHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x13_0x05_FeedbagQueryIfModified
		// expectOutput is the SNAC payload sent from the server to the
		// recipient client
		expectOutput XMessage
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Items: []oscar.FeedbagItem{},
				},
			},
		},
		{
			name:       "retrieve feedbag with items",
			screenName: "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(200000),
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Version: 0,
					Items: []oscar.FeedbagItem{
						{
							Name: "buddy_1",
						},
						{
							Name: "buddy_2",
						},
					},
					LastUpdate: uint32(time.UnixMilli(200000).Unix()),
				},
			},
		},
		{
			name:       "retrieve not-modified response",
			screenName: "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(100000),
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(200000).Unix()),
			},
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReplyNotModified,
				},
				snacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(100000).Unix()),
					Count:      2,
				},
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
				Retrieve(tc.screenName).
				Return(tc.feedbagItems, nil).
				Maybe()
			fm.EXPECT().
				LastModified(tc.screenName).
				Return(tc.lastModified, nil).
				Maybe()
			//
			// send input SNAC
			//
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			svc := FeedbagService{}
			outputSNAC, err := svc.QueryIfModifiedHandler(nil, senderSession, fm, tc.inputSNAC)
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestInsertItemHandler(t *testing.T) {
	defaultSess := &Session{}
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user managing buddy list
		userSession *Session
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x13_0x08_FeedbagInsertItem
		// screenNameLookups is the list of user's online buddies
		screenNameLookups map[string]struct {
			sess *Session
			err  error
		}
		// clientResponse is the message returned to the client
		clientResponse XMessage
		// buddyMessages are events forwarded to buddy clients
		buddyMessages []struct {
			user string
			msg  XMessage
		}
	}{
		{
			name: "user adds 2 online buddies, expect OK response",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 2,
						Name:    "buddy_1_online",
					},
					{
						ClassID: 2,
						Name:    "buddy_2_online",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"user_screen_name": {
					sess: &Session{ScreenName: "user_screen_name"},
				},
				"buddy_1_online": {
					sess: &Session{ScreenName: "buddy_1_online"},
				},
				"buddy_2_online": {
					sess: &Session{ScreenName: "buddy_2_online"},
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
			buddyMessages: []struct {
				user string
				msg  XMessage
			}{
				{
					user: "user_screen_name",
					msg: XMessage{
						snacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyArrived,
						},
						snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName: "buddy_1_online",
								TLVBlock: oscar.TLVBlock{
									TLVList: defaultSess.GetUserInfo(),
								},
							},
						},
					},
				},
				{
					user: "user_screen_name",
					msg: XMessage{
						snacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyArrived,
						},
						snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName: "buddy_2_online",
								TLVBlock: oscar.TLVBlock{
									TLVList: defaultSess.GetUserInfo(),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "user adds an offline buddy, expect OK response and 0 buddy arrived events",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 2,
						Name:    "buddy_offline",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"buddy_offline": {
					err: ErrSessNotFound,
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name: "users adds an invisible buddy, expect OK response and 0 buddy arrived events",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 2,
						Name:    "invisible_buddy_online",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"invisible_buddy_online": {
					sess: &Session{
						ScreenName: "invisible_buddy_online",
						invisible:  true,
					},
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name: "user blocks buddy currently online, expect OK response, buddy departed event client, 1 buddy " +
				"departed event sent to buddy",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "buddy_1",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"user_screen_name": {
					sess: &Session{ScreenName: "user_screen_name"},
				},
				"buddy_1": {
					sess: &Session{ScreenName: "buddy_1"},
				},
			},
			buddyMessages: []struct {
				user string
				msg  XMessage
			}{
				{
					user: "buddy_1",
					msg: XMessage{
						snacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyDeparted,
						},
						snacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "user_screen_name",
								WarningLevel: 0,
							},
						},
					},
				},
				{
					user: "user_screen_name",
					msg: XMessage{
						snacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyDeparted,
						},
						snacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "buddy_1",
								WarningLevel: 0,
							},
						},
					},
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name: "user blocks buddy currently offline, expect OK response and no buddy departed events",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "buddy_1",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *Session
				err  error
			}{
				"user_screen_name": {
					sess: &Session{ScreenName: "user_screen_name"},
				},
				"buddy_1": {
					err: ErrSessNotFound,
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name: "user tries to block themselves, expect feedback error",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "user_screen_name",
					},
				},
			},
			clientResponse: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagErr,
				},
				snacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
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
				Upsert(tc.userSession.ScreenName, tc.inputSNAC.Items).
				Return(nil).
				Maybe()
			fm.EXPECT().
				Buddies(tc.userSession.ScreenName).
				Return([]string{}, nil).
				Maybe()
			sm := NewMockSessionManager(t)
			for screenName, val := range tc.screenNameLookups {
				sm.EXPECT().
					RetrieveByScreenName(screenName).
					Return(val.sess, val.err).
					Maybe()
			}
			for _, n := range tc.buddyMessages {
				sm.EXPECT().
					SendToScreenName(mock.Anything, n.user, n.msg).
					Maybe()
			}
			//
			// send input SNAC
			//
			svc := FeedbagService{}
			output, err := svc.InsertItemHandler(nil, sm, tc.userSession, fm, tc.inputSNAC)
			assert.NoError(t, err)
			//
			// verify response
			//
			assert.Equal(t, output, tc.clientResponse)
		})
	}
}

func TestFeedbagRouter_RouteFeedbag(t *testing.T) {
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
			name: "receive FeedbagRightsQuery, return FeedbagRightsReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagRightsQuery,
				},
				snacOut: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
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
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagRightsReply,
				},
				snacOut: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
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
		},
		{
			name: "receive FeedbagQuery, return FeedbagReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagQuery,
				},
				snacOut: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
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
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Version: 4,
				},
			},
		},
		{
			name: "receive FeedbagQueryIfModified, return FeedbagRightsReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagQueryIfModified,
				},
				snacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: 1234,
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					LastUpdate: 1234,
				},
			},
		},
		{
			name: "receive FeedbagUse, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagUse,
				},
				snacOut: struct{}{},
			},
			output: XMessage{},
		},
		{
			name: "receive FeedbagInsertItem, return BuddyArrived and FeedbagStatus",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagInsertItem,
				},
				snacOut: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagUpdateItem, return BuddyArrived and FeedbagStatus",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagUpdateItem,
				},
				snacOut: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagDeleteItem, return FeedbagStatus",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagDeleteItem,
				},
				snacOut: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagStartCluster, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStartCluster,
				},
				snacOut: oscar.SNAC_0x13_0x11_FeedbagStartCluster{
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
			output: XMessage{},
		},
		{
			name: "receive FeedbagEndCluster, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagEndCluster,
				},
				snacOut: struct{}{},
			},
			output: XMessage{},
		},
		{
			name: "receive FeedbagDeleteUser, return ErrUnsupportedSubGroup",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagDeleteUser,
				},
				snacOut: struct{}{},
			},
			output:    XMessage{},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockFeedbagHandler(t)
			svc.EXPECT().
				DeleteItemHandler(mock.Anything, mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryHandler(mock.Anything, mock.Anything, mock.Anything).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryIfModifiedHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RightsQueryHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				InsertItemHandler(mock.Anything, mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				UpdateItemHandler(mock.Anything, mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				StartClusterHandler(mock.Anything, tc.input.snacOut).
				Maybe()

			router := FeedbagRouter{
				FeedbagHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.snacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteFeedbag(nil, nil, nil, nil, tc.input.snacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.snacFrame == (oscar.SnacFrame{}) {
				return
			}

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence increments
			assert.Equal(t, seq, uint32(1))
			assert.Equal(t, flap.Sequence, uint16(0))

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.snacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.snacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
