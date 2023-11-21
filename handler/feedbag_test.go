package handler

import (
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
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
		expectOutput oscar.XMessage
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			fm := newMockFeedbagManager(t)
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
			senderSession := newTestSession(tc.screenName)
			svc := FeedbagService{
				feedbagManager: fm,
			}
			outputSNAC, err := svc.QueryHandler(nil, senderSession)
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
		expectOutput oscar.XMessage
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "user_screen_name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReplyNotModified,
				},
				SnacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
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
			fm := newMockFeedbagManager(t)
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
			senderSession := newTestSession(tc.screenName)
			svc := FeedbagService{
				feedbagManager: fm,
			}
			outputSNAC, err := svc.QueryIfModifiedHandler(nil, senderSession, tc.inputSNAC)
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestInsertItemHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user managing buddy list
		userSession *state.Session
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x13_0x08_FeedbagInsertItem
		// screenNameLookups is the list of user's online buddies
		screenNameLookups map[string]struct {
			sess *state.Session
			err  error
		}
		// clientResponse is the message returned to the client
		clientResponse oscar.XMessage
		// buddyMessages are events forwarded to buddy clients
		buddyMessages []struct {
			user string
			msg  oscar.XMessage
		}
	}{
		{
			name:        "user adds 2 online buddies, expect OK response",
			userSession: newTestSession("user_screen_name"),
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
				sess *state.Session
				err  error
			}{
				"user_screen_name": {
					sess: newTestSession("user_screen_name", sessOptCannedSignonTime),
				},
				"buddy_1_online": {
					sess: newTestSession("buddy_1_online", sessOptCannedSignonTime),
				},
				"buddy_2_online": {
					sess: newTestSession("buddy_2_online", sessOptCannedSignonTime),
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
			buddyMessages: []struct {
				user string
				msg  oscar.XMessage
			}{
				{
					user: "user_screen_name",
					msg: oscar.XMessage{
						SnacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyArrived,
						},
						SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName: "buddy_1_online",
								TLVBlock: oscar.TLVBlock{
									TLVList: newTestSession("", sessOptCannedSignonTime).UserInfo(),
								},
							},
						},
					},
				},
				{
					user: "user_screen_name",
					msg: oscar.XMessage{
						SnacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyArrived,
						},
						SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName: "buddy_2_online",
								TLVBlock: oscar.TLVBlock{
									TLVList: newTestSession("", sessOptCannedSignonTime).UserInfo(),
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "user adds an offline buddy, expect OK response and 0 buddy arrived events",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 2,
						Name:    "buddy_offline",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"buddy_offline": {
					sess: nil,
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "users adds an invisible buddy, expect OK response and 0 buddy arrived events",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 2,
						Name:    "invisible_buddy_online",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"invisible_buddy_online": {
					sess: newTestSession("invisible_buddy_online", sessOptInvisible),
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name: "user blocks buddy currently online, expect OK response, buddy departed event client, 1 buddy " +
				"departed event sent to buddy",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "buddy_1",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"user_screen_name": {
					sess: newTestSession("user_screen_name"),
				},
				"buddy_1": {
					sess: newTestSession("buddy_1"),
				},
			},
			buddyMessages: []struct {
				user string
				msg  oscar.XMessage
			}{
				{
					user: "buddy_1",
					msg: oscar.XMessage{
						SnacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyDeparted,
						},
						SnacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "user_screen_name",
								WarningLevel: 0,
							},
						},
					},
				},
				{
					user: "user_screen_name",
					msg: oscar.XMessage{
						SnacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyDeparted,
						},
						SnacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "buddy_1",
								WarningLevel: 0,
							},
						},
					},
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks buddy currently offline, expect OK response and a superfluous buddy departed events",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "buddy_1",
					},
				},
			},
			screenNameLookups: map[string]struct {
				sess *state.Session
				err  error
			}{
				"user_screen_name": {
					sess: newTestSession("user_screen_name"),
				},
				"buddy_1": {
					sess: nil,
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
			buddyMessages: []struct {
				user string
				msg  oscar.XMessage
			}{
				{
					user: "buddy_1",
					msg: oscar.XMessage{
						SnacFrame: oscar.SnacFrame{
							FoodGroup: oscar.BUDDY,
							SubGroup:  oscar.BuddyDeparted,
						},
						SnacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
							TLVUserInfo: oscar.TLVUserInfo{
								ScreenName:   "user_screen_name",
								WarningLevel: 0,
							},
						},
					},
				},
			},
		},
		{
			name:        "user tries to block themselves, expect feedback error",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
				Items: []oscar.FeedbagItem{
					{
						ClassID: 3,
						Name:    "user_screen_name",
					},
				},
			},
			clientResponse: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagErr,
				},
				SnacOut: oscar.SnacError{
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
			fm := newMockFeedbagManager(t)
			fm.EXPECT().
				Upsert(tc.userSession.ScreenName(), tc.inputSNAC.Items).
				Return(nil).
				Maybe()
			fm.EXPECT().
				Buddies(tc.userSession.ScreenName()).
				Return([]string{}, nil).
				Maybe()
			sm := newMockSessionManager(t)
			for screenName, val := range tc.screenNameLookups {
				sm.EXPECT().
					RetrieveByScreenName(screenName).
					Return(val.sess).
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
			svc := FeedbagService{
				feedbagManager: fm,
				sessionManager: sm,
			}
			output, err := svc.InsertItemHandler(nil, tc.userSession, tc.inputSNAC)
			assert.NoError(t, err)
			//
			// verify response
			//
			assert.Equal(t, output, tc.clientResponse)
		})
	}
}
