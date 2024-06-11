package foodgroup

import (
	"log/slog"
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
	svc := NewFeedbagService(nil, nil, nil, nil, nil)

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
						61,
						61,
						100,
						100,
						1,
						1,
						50,
						0x00,
						0x00,
						3,
						0x00,
						0x00,
						0x00,
						128,
						255,
						20,
						200,
						1,
						0x00,
						1,
						200,
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

func TestFeedbagService_UpsertItem(t *testing.T) {
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_1_online"),
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: state.NewIdentScreenName("buddy_2_online"),
							sess:       newTestSession("buddy_2_online", sessOptCannedSignonTime),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					unicastBuddyArrivedParams: unicastBuddyArrivedParams{
						{
							from: state.NewIdentScreenName("buddy_1_online"),
							to:   state.NewIdentScreenName("user_screen_name"),
						},
						{
							from: state.NewIdentScreenName("buddy_2_online"),
							to:   state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_offline"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("invisible_buddy_online"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_1"),
							sess:       newTestSession("buddy_1"),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					unicastBuddyDepartedParams: unicastBuddyDepartedParams{
						{
							from: state.NewIdentScreenName("user_screen_name"),
							to:   state.NewIdentScreenName("buddy_1"),
						},
						{
							from: state.NewIdentScreenName("buddy_1"),
							to:   state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_1"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
		{
			name:        "add icon hash to feedbag, icon doesn't exist in BART store, instruct client to upload icon",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIdBart,
							TLVLBlock: wire.TLVLBlock{
								TLVList: wire.TLVList{
									wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
										Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
									}),
								},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				bartManagerParams: bartManagerParams{
					bartManagerRetrieveParams: bartManagerRetrieveParams{
						{
							itemHash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
							result:   []byte{}, // icon doesn't exist
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBart,
									TLVLBlock: wire.TLVLBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
												Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											}),
										},
									},
								},
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.OService,
									SubGroup:  wire.OServiceBartReply,
								},
								Body: wire.SNAC_0x01_0x21_OServiceBARTReply{
									BARTID: wire.BARTID{
										Type: wire.BARTTypesBuddyIcon,
										BARTInfo: wire.BARTInfo{
											Flags: wire.BARTFlagsCustom | wire.BARTFlagsUnknown,
											Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
										},
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
			name:        "add icon hash to feedbag, icon already exists in BART store, notify buddies about icon change",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIdBart,
							TLVLBlock: wire.TLVLBlock{
								TLVList: wire.TLVList{
									wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
										Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
									}),
								},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				bartManagerParams: bartManagerParams{
					bartManagerRetrieveParams: bartManagerRetrieveParams{
						{
							itemHash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
							result:   []byte{'i', 'c', 'o', 'n', 'd', 'a', 't', 'a'},
						},
					},
				},
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBart,
									TLVLBlock: wire.TLVLBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
												Hash: []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
											}),
										},
									},
								},
							},
						},
					},
					adjacentUsersParams: adjacentUsersParams{},
					feedbagParams:       feedbagParams{},
				},
				messageRelayerParams: messageRelayerParams{
					relayToScreenNameParams: relayToScreenNameParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.OService,
									SubGroup:  wire.OServiceBartReply,
								},
								Body: wire.SNAC_0x01_0x21_OServiceBARTReply{
									BARTID: wire.BARTID{
										Type: wire.BARTTypesBuddyIcon,
										BARTInfo: wire.BARTInfo{
											Flags: wire.BARTFlagsCustom,
											Hash:  []byte{'t', 'h', 'e', 'h', 'a', 's', 'h'},
										},
									},
								},
							},
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
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
			name:        "clear icon, notify buddies about icon change",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []wire.FeedbagItem{
						{
							ClassID: wire.FeedbagClassIdBart,
							TLVLBlock: wire.TLVLBlock{
								TLVList: wire.TLVList{
									wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
										Hash: wire.GetClearIconHash(),
									}),
								},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					feedbagUpsertParams: feedbagUpsertParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
							items: []wire.FeedbagItem{
								{
									ClassID: wire.FeedbagClassIdBart,
									TLVLBlock: wire.TLVLBlock{
										TLVList: wire.TLVList{
											wire.NewTLV(wire.FeedbagAttributesBartInfo, wire.BARTInfo{
												Hash: wire.GetClearIconHash(),
											}),
										},
									},
								},
							},
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("user_screen_name"),
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
			for _, params := range tc.mockParams.feedbagManagerParams.feedbagParams {
				feedbagManager.EXPECT().
					Feedbag(params.screenName).
					Return(params.results, nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tc.mockParams.messageRelayerParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNameParams {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, params.screenName, params.message)
			}
			for _, params := range tc.mockParams.messageRelayerParams.relayToScreenNamesParams {
				messageRelayer.EXPECT().
					RelayToScreenNames(mock.Anything, params.screenNames, params.message)
			}
			bartManager := newMockBARTManager(t)
			for _, params := range tc.mockParams.bartManagerParams.bartManagerRetrieveParams {
				bartManager.EXPECT().
					BARTRetrieve(params.itemHash).
					Return(params.result, nil)
			}
			legacyBuddyListManager := newMockLegacyBuddyListManager(t)
			for _, params := range tc.mockParams.whoAddedUserParams {
				legacyBuddyListManager.EXPECT().
					WhoAddedUser(params.userScreenName).
					Return(params.result)
			}
			buddyUpdateBroadcaster := newMockBuddyBroadcaster(t)
			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.IdentScreenName() == p.screenName
					})).
					Return(nil)
			}
			for _, params := range tc.mockParams.unicastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					UnicastBuddyArrived(mock.Anything,
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.from
						}),
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.to
						})).
					Return(nil)
			}
			for _, params := range tc.mockParams.unicastBuddyDepartedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					UnicastBuddyDeparted(mock.Anything,
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.from
						}),
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.to
						}))
			}

			svc := NewFeedbagService(slog.Default(), messageRelayer, feedbagManager, bartManager, buddyUpdateBroadcaster)
			output, err := svc.UpsertItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x08_FeedbagInsertItem).Items)
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_1_online"),
							sess:       newTestSession("buddy_1_online", sessOptCannedSignonTime),
						},
						{
							screenName: state.NewIdentScreenName("buddy_2_online"),
							sess:       newTestSession("buddy_2_online", sessOptCannedSignonTime),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					unicastBuddyArrivedParams: unicastBuddyArrivedParams{
						{
							from: state.NewIdentScreenName("buddy_1_online"),
							to:   state.NewIdentScreenName("user_screen_name"),
						},
						{
							from: state.NewIdentScreenName("buddy_2_online"),
							to:   state.NewIdentScreenName("user_screen_name"),
						},
						{
							from: state.NewIdentScreenName("user_screen_name"),
							to:   state.NewIdentScreenName("buddy_1_online"),
						},
						{
							from: state.NewIdentScreenName("user_screen_name"),
							to:   state.NewIdentScreenName("buddy_2_online"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("buddy_offline"),
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
							screenName: state.NewIdentScreenName("user_screen_name"),
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
							screenName: state.NewIdentScreenName("invisible_buddy_online"),
							sess:       newTestSession("invisible_buddy_online", sessOptInvisible),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					unicastBuddyArrivedParams: unicastBuddyArrivedParams{
						{
							from: state.NewIdentScreenName("user_screen_name"),
							to:   state.NewIdentScreenName("invisible_buddy_online"),
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
			buddyUpdateBroadcast := newMockBuddyBroadcaster(t)
			for _, params := range tc.mockParams.unicastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcast.EXPECT().
					UnicastBuddyArrived(mock.Anything,
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.from
						}),
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.to
						})).
					Return(nil)
			}

			svc := FeedbagService{
				buddyUpdateBroadcaster: buddyUpdateBroadcast,
				feedbagManager:         feedbagManager,
				messageRelayer:         messageRelayer,
			}
			output, err := svc.DeleteItem(nil, tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x13_0x0A_FeedbagDeleteItem))
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expectOutput)
		})
	}
}

func TestFeedbagService_Use(t *testing.T) {
	tests := []struct {
		// name is the name of the test
		name string
		// joiningChatter is the session of the arriving user
		sess *state.Session
		// bodyIn is the SNAC body sent from the arriving user's client to the
		// server
		bodyIn wire.SNAC_0x01_0x02_OServiceClientOnline
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		wantErr    error
	}{
		{
			name:   "notify arriving user's buddies of its arrival and populate the arriving user's buddy list",
			sess:   newTestSession("test-user"),
			bodyIn: wire.SNAC_0x01_0x02_OServiceClientOnline{},
			mockParams: mockParams{
				feedbagManagerParams: feedbagManagerParams{
					buddiesParams: buddiesParams{
						{
							screenName: state.NewIdentScreenName("test-user"),
							results: []state.IdentScreenName{
								state.NewIdentScreenName("buddy1"),
								state.NewIdentScreenName("buddy3"),
							},
						},
					},
				},
				messageRelayerParams: messageRelayerParams{
					retrieveByScreenNameParams: retrieveByScreenNameParams{
						{
							screenName: state.NewIdentScreenName("buddy1"),
							sess:       newTestSession("buddy1"),
						},
						{
							screenName: state.NewIdentScreenName("buddy3"),
							sess:       newTestSession("buddy3"),
						},
					},
				},
				buddyBroadcasterParams: buddyBroadcasterParams{
					unicastBuddyArrivedParams: unicastBuddyArrivedParams{
						{
							from: state.NewIdentScreenName("buddy1"),
							to:   state.NewIdentScreenName("test-user"),
						},
						{
							from: state.NewIdentScreenName("buddy3"),
							to:   state.NewIdentScreenName("test-user"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedbagManager := newMockFeedbagManager(t)
			for _, params := range tt.mockParams.buddiesParams {
				feedbagManager.EXPECT().
					Buddies(params.screenName).
					Return(params.results, nil)
			}
			messageRelayer := newMockMessageRelayer(t)
			for _, params := range tt.mockParams.retrieveByScreenNameParams {
				messageRelayer.EXPECT().
					RetrieveByScreenName(params.screenName).
					Return(params.sess)
			}
			buddyUpdateBroadcast := newMockBuddyBroadcaster(t)
			for _, params := range tt.mockParams.unicastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcast.EXPECT().
					UnicastBuddyArrived(mock.Anything,
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.from
						}),
						mock.MatchedBy(func(s *state.Session) bool {
							return s.IdentScreenName() == p.to
						})).
					Return(nil)
			}

			svc := NewFeedbagService(slog.Default(), messageRelayer, feedbagManager, nil, buddyUpdateBroadcast)

			haveErr := svc.Use(nil, tt.sess)
			assert.ErrorIs(t, tt.wantErr, haveErr)
		})
	}
}
