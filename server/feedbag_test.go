package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestReceiveAndSendFeedbagQuery(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody oscar.SNAC_0x13_0x06_FeedbagReply
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{},
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
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			var seq uint32
			snac := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagQuery,
			}
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			output := &bytes.Buffer{}
			assert.NoError(t, ReceiveAndSendFeedbagQuery(senderSession, fm, snac, output, &seq))
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
			actual := oscar.SNAC_0x13_0x06_FeedbagReply{}
			assert.NoError(t, oscar.Unmarshal(&actual, output))
			assert.Equal(t, tc.expectSNACBody, actual)
		})
	}
}

func TestReceiveAndSendFeedbagQueryIfModified(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x13_0x05_FeedbagQueryIfModified
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{},
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
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReplyNotModified,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(200000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				Count:      2,
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
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			snac := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagQuery,
			}
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			assert.NoError(t, ReceiveAndSendFeedbagQueryIfModified(senderSession, fm, snac, input, output, &seq))
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
			case oscar.SNAC_0x13_0x06_FeedbagReply:
				outputSNAC := oscar.SNAC_0x13_0x06_FeedbagReply{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			case oscar.SNAC_0x13_0x05_FeedbagQueryIfModified:
				outputSNAC := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
		})
	}
}

func TestReceiveInsertItem(t *testing.T) {
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
		// clientResponses is messages returned to the client
		clientResponses []XMessage
		// buddyMessages are events forwarded to buddy clients
		buddyMessages map[string]XMessage
	}{
		{
			name: "user adds 2 online buddies, expect OK response and 2 buddy arrived client events",
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
				"buddy_1_online": {
					sess: &Session{ScreenName: "buddy_1_online"},
				},
				"buddy_2_online": {
					sess: &Session{ScreenName: "buddy_2_online"},
				},
			},
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagStatus,
					},
					snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
						Results: []uint16{0x0000, 0x0000},
					},
				},
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: BUDDY,
						SubGroup:  BuddyArrived,
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
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: BUDDY,
						SubGroup:  BuddyArrived,
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
					err: errSessNotFound,
				},
			},
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagStatus,
					},
					snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
						Results: []uint16{0x0000},
					},
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
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagStatus,
					},
					snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
						Results: []uint16{0x0000},
					},
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
				"buddy_1": {
					sess: &Session{ScreenName: "buddy_1"},
				},
			},
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagStatus,
					},
					snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
						Results: []uint16{0x0000},
					},
				},
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: BUDDY,
						SubGroup:  BuddyDeparted,
					},
					snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
						TLVUserInfo: oscar.TLVUserInfo{
							ScreenName: "buddy_1",
						},
					},
				},
			},
			buddyMessages: map[string]XMessage{
				"buddy_1": {
					snacFrame: oscar.SnacFrame{
						FoodGroup: BUDDY,
						SubGroup:  BuddyDeparted,
					},
					snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
						TLVUserInfo: oscar.TLVUserInfo{
							ScreenName: "user_screen_name",
						},
					},
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
				"buddy_1": {
					err: errSessNotFound,
				},
			},
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagStatus,
					},
					snacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
						Results: []uint16{0x0000},
					},
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
			clientResponses: []XMessage{
				{
					snacFrame: oscar.SnacFrame{
						FoodGroup: FEEDBAG,
						SubGroup:  FeedbagErr,
					},
					snacOut: oscar.SnacError{
						Code: ErrorCodeNotSupportedByHost,
					},
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
			for _, msg := range tc.buddyMessages {
				sm.EXPECT().
					SendToScreenName(mock.Anything, msg).
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
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagQuery,
			}

			assert.NoError(t, ReceiveInsertItem(sm, tc.userSession, fm, snac, input, output, &seq))
			//
			// verify response
			//
			for _, xMsg := range tc.clientResponses {
				flap := oscar.FlapFrame{}
				assert.NoError(t, oscar.Unmarshal(&flap, output))
				snacBuf, err := flap.SNACBuffer(output)
				assert.NoError(t, err)

				snacFrame := oscar.SnacFrame{}
				assert.NoError(t, oscar.Unmarshal(&snacFrame, snacBuf))
				assert.Equal(t, xMsg.snacFrame, snacFrame)

				switch v := xMsg.snacOut.(type) {
				case oscar.SNAC_0x13_0x0E_FeedbagStatus:
					outputSNAC := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
					assert.NoError(t, oscar.Unmarshal(&outputSNAC, snacBuf))
					assert.Equal(t, v, outputSNAC)
				case oscar.SNAC_0x03_0x0A_BuddyArrived:
					assert.NoError(t, v.SerializeInPlace())
					outputSNAC := oscar.SNAC_0x03_0x0A_BuddyArrived{}
					assert.NoError(t, oscar.Unmarshal(&outputSNAC, snacBuf))
					assert.Equal(t, v, outputSNAC)
				case oscar.SnacError:
					outputSNAC := oscar.SnacError{}
					assert.NoError(t, oscar.Unmarshal(&outputSNAC, snacBuf))
					assert.Equal(t, v, outputSNAC)
				default:
					t.Fatalf("unexpected output SNAC type")
				}
			}
			assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
		})
	}
}
