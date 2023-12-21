package handler

import (
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbagService_QueryHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name:        "retrieve empty feedbag",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					retrieveParams: retrieveParams{
						{
							screenName: "user_screen_name",
							results:    []oscar.FeedbagItem{},
						},
					},
					lastModifiedParams: lastModifiedParams{},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
					Items: []oscar.FeedbagItem{},
				},
			},
		},
		{
			name:        "retrieve feedbag with items",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					retrieveParams: retrieveParams{
						{
							screenName: "user_screen_name",
							results: []oscar.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					lastModifiedParams: lastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(1696472198082),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.retrieveParams {
				feedbagManager.EXPECT().
					Retrieve(params.screenName).
					Return(params.results, nil)
			}
			for _, params := range tc.mockParams.lastModifiedParams {
				feedbagManager.EXPECT().
					LastModified(params.screenName).
					Return(params.result, nil)
			}

			svc := FeedbagService{
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.QueryHandler(nil, tc.userSession, tc.inputSNAC.Frame)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestFeedbagService_QueryIfModifiedHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name:        "retrieve empty feedbag",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					retrieveParams: retrieveParams{
						{
							screenName: "user_screen_name",
							results:    []oscar.FeedbagItem{},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
					Items: []oscar.FeedbagItem{},
				},
			},
		},
		{
			name:        "retrieve feedbag with items",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					retrieveParams: retrieveParams{
						{
							screenName: "user_screen_name",
							results: []oscar.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					lastModifiedParams: lastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(200000),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
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
			name:        "retrieve not-modified response",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(time.UnixMilli(200000).Unix()),
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					retrieveParams: retrieveParams{
						{
							screenName: "user_screen_name",
							results: []oscar.FeedbagItem{
								{
									Name: "buddy_1",
								},
								{
									Name: "buddy_2",
								},
							},
						},
					},
					lastModifiedParams: lastModifiedParams{
						{
							screenName: "user_screen_name",
							result:     time.UnixMilli(100000),
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReplyNotModified,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
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
			for _, params := range tc.mockParams.retrieveParams {
				feedbagManager.EXPECT().
					Retrieve(params.screenName).
					Return(params.results, nil)
			}
			for _, params := range tc.mockParams.lastModifiedParams {
				feedbagManager.EXPECT().
					LastModified(params.screenName).
					Return(params.result, nil)
			}
			//
			// send input SNAC
			//
			svc := FeedbagService{
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.QueryIfModifiedHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x13_0x05_FeedbagQueryIfModified))
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestFeedbagService_RightsQueryHandler(t *testing.T) {
	svc := NewFeedbagService(nil, nil)

	outputSNAC := svc.RightsQueryHandler(nil, oscar.SNACFrame{RequestID: 1234})
	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagRightsReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.FeedbagRightsMaxItemAttrs, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxItemsByClass, []uint16{
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
					oscar.NewTLV(oscar.FeedbagRightsMaxClientItems, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxItemNameLen, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxRecentBuddies, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionBuddies, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionHalfLife, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsInteractionMaxScore, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxBuddiesPerGroup, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxMegaBots, uint16(200)),
					oscar.NewTLV(oscar.FeedbagRightsMaxSmartGroups, uint16(100)),
				},
			},
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}

func TestFeedbagService_InsertItemHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name:        "user adds online buddies to feedbag, receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_1_online",
						},
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
									Name:    "buddy_1_online",
								},
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_1_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "user_screen_name",
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user adds offline buddy to feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user adds invisible buddy to feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks online buddy, buddy receives buddy departure notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
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
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyDeparted,
								},
								Body: oscar.SNAC_0x03_0x0C_BuddyDeparted{
									TLVUserInfo: oscar.TLVUserInfo{
										ScreenName:   "user_screen_name",
										WarningLevel: 0,
									},
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks offline buddy, no buddy departure notification sent",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "invisible user blocks online buddy, no buddy departure notification sent",
			userSession: newTestSession("user_screen_name", sessOptInvisible),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_1",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
									Name:    "buddy_1",
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user blocks themselves, receives error",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "user_screen_name",
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.upsertParams {
				feedbagManager.EXPECT().
					Upsert(params.screenName, params.items).
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
			output, err := svc.InsertItemHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x13_0x08_FeedbagInsertItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestFeedbagService_UpdateItemHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name:        "user updates online buddies in feedbag, receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_1_online",
						},
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
									Name:    "buddy_1_online",
								},
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_1_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "user_screen_name",
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user updates offline buddy in feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user updates an invisible buddy in feedbag, receives no buddy arrival notification",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDPermit,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					upsertParams: upsertParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDPermit,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.upsertParams {
				feedbagManager.EXPECT().
					Upsert(params.screenName, params.items).
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
			output, err := svc.UpdateItemHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x13_0x09_FeedbagUpdateItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestFeedbagService_DeleteItemHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user adding to feedbag
		userSession *state.Session
		// inputSNAC is the SNAC sent from the client to the server
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectOutput is the SNAC sent from the server to client
		expectOutput oscar.SNACMessage
	}{
		{
			name:        "user deletes buddy",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIdBuddy,
							Name:    "buddy_1_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					deleteParams: deleteParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIdBuddy,
									Name:    "buddy_1_online",
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user unblocks buddies, user and buddies receive buddy arrival notifications",
			userSession: newTestSession("user_screen_name", sessOptCannedSignonTime),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_1_online",
						},
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_2_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					deleteParams: deleteParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
									Name:    "buddy_1_online",
								},
								{
									ClassID: oscar.FeedbagClassIDDeny,
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
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_1_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "buddy_1_online",
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "user_screen_name",
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("buddy_2_online", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
						{
							screenName: "buddy_2_online",
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000, 0x0000},
				},
			},
		},
		{
			name:        "user unblocks offline buddy, receives no buddy arrival notifications",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "buddy_offline",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					deleteParams: deleteParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
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
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
		{
			name:        "user unblocks invisible buddy, user receives no buddy arrival notification, buddy receives buddy arrival notifications",
			userSession: newTestSession("user_screen_name", sessOptCannedSignonTime),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							ClassID: oscar.FeedbagClassIDDeny,
							Name:    "invisible_buddy_online",
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					deleteParams: deleteParams{
						{
							screenName: "user_screen_name",
							items: []oscar.FeedbagItem{
								{
									ClassID: oscar.FeedbagClassIDDeny,
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
							message: oscar.SNACMessage{
								Frame: oscar.SNACFrame{
									FoodGroup: oscar.Buddy,
									SubGroup:  oscar.BuddyArrived,
								},
								Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
									TLVUserInfo: newTestSession("user_screen_name", sessOptCannedSignonTime).TLVUserInfo(),
								},
							},
						},
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{0x0000},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tc.mockParams.feedbagManagerParams.deleteParams {
				feedbagManager.EXPECT().
					Delete(params.screenName, params.items).
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
			output, err := svc.DeleteItemHandler(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x13_0x0A_FeedbagDeleteItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}
