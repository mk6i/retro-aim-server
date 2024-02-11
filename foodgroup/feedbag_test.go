package foodgroup

import (
	"testing"
	"time"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbagService_Query(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "retrieve empty feedbag",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "user_screen_name",
							results:    []wire.FeedbagItem{},
						},
					},
					feedbagLastModifiedParams: feedbagLastModifiedParams{},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x06_FeedbagReply{
					Items: []wire.FeedbagItem{},
				},
			},
		},
		{
			name:        "retrieve feedbag with items",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "user_screen_name",
							results: []wire.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					feedbagLastModifiedParams: feedbagLastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(1696472198082),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x06_FeedbagReply{
					Version: 0,
					Items: []wire.FeedbagItem{
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
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}
			for _, params := range tc.mockParams.feedbagLastModifiedParams {
				feedbagManager.EXPECT().
					FeedbagLastModified(params.screenName).
					Return(params.result, nil)
			}

			svc := FeedbagService{
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.Query(nil, tc.userSession, tc.inputSNAC.Frame)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestFeedbagService_QueryIfModified(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "retrieve empty feedbag",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "user_screen_name",
							results:    []wire.FeedbagItem{},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x06_FeedbagReply{
					Items: []wire.FeedbagItem{},
				},
			},
		},
		{
			name:        "retrieve feedbag with items",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "user_screen_name",
							results: []wire.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					feedbagLastModifiedParams: feedbagLastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(200000),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x06_FeedbagReply{
					Version: 0,
					Items: []wire.FeedbagItem{
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
			name:        "retrieve not-modified response",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(200000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagParams: feedbagParams{
						{
							screenName: "user_screen_name",
							results: []wire.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					feedbagLastModifiedParams: feedbagLastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(100000),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagReplyNotModified,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x05_FeedbagQueryIfModified{
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
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}
			for _, params := range tc.mockParams.feedbagLastModifiedParams {
				feedbagManager.EXPECT().
					FeedbagLastModified(params.screenName).
					Return(params.result, nil)
			}
			//
			// send input SNAC
			//
			svc := FeedbagService{
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.QueryIfModified(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x05_FeedbagQueryIfModified))
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestFeedbagService_RightsQuery(t *testing.T) {
	svc := NewFeedbagService(nil, nil)

	outputSNAC := svc.RightsQuery(nil, wire.SNACFrame{RequestID: 1234})
	expectSNAC := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Feedbag,
			SubGroup:  wire.FeedbagRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.FeedbagRightsMaxItemAttrs, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxItemsByClass, []uint16{
						0x3D,
						0x3D,
						0x64,
						0x64,
						0x01,
						0x01,
						0x32,
						0x00,
						0x00,
						0x03,
						0x00,
						0x00,
						0x00,
						0x80,
						0xFF,
						0x14,
						0xC8,
						0x01,
						0x00,
						0x01,
						0x00,
					}),
					wire.NewTLV(wire.FeedbagRightsMaxClientItems, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxItemNameLen, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxRecentBuddies, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionBuddies, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionHalfLife, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsInteractionMaxScore, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxBuddiesPerGroup, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxMegaBots, uint16(200)),
					wire.NewTLV(wire.FeedbagRightsMaxSmartGroups, uint16(100)),
				},
			},
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}

func TestFeedbagService_InsertItem(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "user adds online buddies to feedbag, receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_1_online",
						},
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_1_online",
								},
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_2_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1_online",
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: "buddy_2_online",
							sess:       newTestSession("buddy_2_online", sessOptCannedSignonTime),
						},
					},
					sendToScreenNameParams: sendToScreenNameParams{
						{
							screenName: "user_screen_name",
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
						{
							screenName: "user_screen_name",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user adds offline buddy to feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_offline",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_offline",
							sess:       nil,
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user adds invisible buddy to feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "invisible_buddy_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "invisible_buddy_online",
							sess:       newTestSession("invisible_buddy_online", sessOptInvisible),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks online buddy, buddy receives buddy departure notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_1",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1",
							sess:       newTestSession("buddy_1"),
						},
					},
					sendToScreenNameParams: sendToScreenNameParams{
						{
							screenName: "buddy_1",
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
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks offline buddy, no buddy departure notification sent",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_1",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1",
							sess:       nil,
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "invisible user blocks online buddy, no buddy departure notification sent",
			userSession: newTestSession("user_screen_name", sessOptInvisible),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_1",
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks themselves, receives error",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "user_screen_name",
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagUpsertParams {
				feedbagManager.EXPECT().
					FeedbagUpsert(params.screenName, params.items).
					Return(nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tc.mockParams.messageRelayerParams.sendToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := FeedbagService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}
			output, err := svc.InsertItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x08_FeedbagInsertItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestFeedbagService_UpdateItem(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "user updates online buddies in feedbag, receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_1_online",
						},
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_1_online",
								},
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_2_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1_online",
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: "buddy_2_online",
							sess:       newTestSession("buddy_2_online", sessOptCannedSignonTime),
						},
					},
					sendToScreenNameParams: sendToScreenNameParams{
						{
							screenName: "user_screen_name",
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
						{
							screenName: "user_screen_name",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user updates offline buddy in feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "buddy_offline",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_offline",
							sess:       nil,
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user updates an invisible buddy in feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDPermit,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDPermit,
									Name:    "invisible_buddy_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "invisible_buddy_online",
							sess:       newTestSession("invisible_buddy_online", sessOptInvisible),
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagUpsertParams {
				feedbagManager.EXPECT().
					FeedbagUpsert(params.screenName, params.items).
					Return(nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tc.mockParams.messageRelayerParams.sendToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := FeedbagService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}
			output, err := svc.UpdateItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x09_FeedbagUpdateItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestFeedbagService_DeleteItem(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
	}{
		{
			name:        "user deletes buddy",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIdBuddy,
							Name:    "buddy_1_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagDeleteParams: feedbagDeleteParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBuddy,
									Name:    "buddy_1_online",
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user unblocks buddies, user and buddies receive buddy arrival notifications",
			userSession: newTestSession("user_screen_name", sessOptCannedSignonTime),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_1_online",
						},
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagDeleteParams: feedbagDeleteParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_1_online",
								},
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_2_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_1_online",
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: "buddy_2_online",
							sess:       newTestSession("buddy_2_online", sessOptCannedSignonTime),
						},
					},
					sendToScreenNameParams: sendToScreenNameParams{
						{
							screenName: "user_screen_name",
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
						{
							screenName: "buddy_1_online",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "user_screen_name",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "buddy_2_online",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user unblocks offline buddy, receives no buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagDeleteParams: feedbagDeleteParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "buddy_offline",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "buddy_offline",
							sess:       nil,
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user unblocks invisible buddy, user receives no buddy arrival notification, buddy receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name", sessOptCannedSignonTime),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIDDeny,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagDeleteParams: feedbagDeleteParams{
						{
							screenName: "user_screen_name",
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIDDeny,
									Name:    "invisible_buddy_online",
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: "invisible_buddy_online",
							sess:       newTestSession("invisible_buddy_online", sessOptInvisible),
						},
					},
					sendToScreenNameParams: sendToScreenNameParams{
						{
							screenName: "invisible_buddy_online",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Buddy,
									SubGroup:  wire.BuddyArrived,
								},
								Body: wire.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Feedbag,
					SubGroup:  wire.FeedbagStatus,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagDeleteParams {
				feedbagManager.EXPECT().
					FeedbagDelete(params.screenName, params.items).
					Return(nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tc.mockParams.messageRelayerParams.sendToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}

			svc := FeedbagService{
				feedbagManager: feedbagManager,
				messageRelayer: messageRelayer,
			}
			output, err := svc.DeleteItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x0A_FeedbagDeleteItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}
