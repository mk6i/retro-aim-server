package foodgroup

import (
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICBMService_ChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		senderSession    *state.Session
		recipientSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient wire.SNACMessage
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectOutput *wire.SNACMessage
	}{
		{
			name:             "transmit message from sender to recipient, ack message back to sender",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectSNACToClient: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMChannelMsgToClient,
				},
				Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVWantEvents,
								Value: []byte{},
							},
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMHostAck,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x0C_ICBMHostAck{
					ScreenName: "recipient-screen-name",
				},
			},
		},
		{
			name:             "transmit message from sender to recipient, don't ack message back to sender",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{},
					},
				},
			},
			expectSNACToClient: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMChannelMsgToClient,
				},
				Body: wire.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVWantEvents,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     state.BlockedA,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeInLocalPermitDeny,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient has blocked sender",
			blockedState:     state.BlockedB,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient doesn't exist",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: nil,
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ICBMTLVRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
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
			feedbagManager.EXPECT().
				BlockedState(tc.senderSession.IdentScreenName(),
					state.NewIdentScreenName(tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName)).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(state.NewIdentScreenName(tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName)).
				Return(tc.recipientSession).
				Maybe()
			if tc.recipientSession != nil {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, tc.recipientSession.IdentScreenName(), tc.expectSNACToClient).
					Maybe()
			}
			//
			// send input SNAC
			//
			svc := NewICBMService(messageRelayer, feedbagManager, nil)
			outputSNAC, err := svc.ChannelMsgToHost(nil, tc.senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost))
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMService_ClientEvent(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// senderScreenName is the screen name of the user sending the event
		senderScreenName state.DisplayScreenName
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient wire.SNACMessage
	}{
		{
			name:             "transmit message from sender to recipient",
			blockedState:     state.BlockedNo,
			senderScreenName: "sender-screen-name",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     12345678,
					ChannelID:  42,
					ScreenName: "recipient-screen-name",
					Event:      12,
				},
			},
			expectSNACToClient: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMClientEvent,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     12345678,
					ChannelID:  42,
					ScreenName: "sender-screen-name",
					Event:      12,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     state.BlockedA,
			senderScreenName: "sender-screen-name",
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
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
			feedbagManager.EXPECT().
				BlockedState(tc.senderScreenName.IdentScreenName(),
					state.NewIdentScreenName(tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent).ScreenName)).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			if tc.blockedState == state.BlockedNo {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything,
						state.NewIdentScreenName(tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent).ScreenName),
						tc.expectSNACToClient)
			}
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderScreenName)
			svc := NewICBMService(messageRelayer, feedbagManager, nil)
			assert.NoError(t, svc.ClientEvent(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent)))
		})
	}
}

func TestICBMService_EvilRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		// senderScreenName is the session of the user sending the EvilRequest
		senderSession *state.Session
		// recipientSession is the session of the user receiving the EvilNotification
		recipientSession *state.Session
		// recipientScreenName is the screen name of the user receiving the EvilNotification
		recipientScreenName state.IdentScreenName
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient wire.SNACMessage
		// expectOutput is the SNAC sent from the server to client
		expectOutput wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
	}{
		{
			name:                "transmit anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    newTestSession("recipient-screen-name", sessOptCannedSignonTime),
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     1, // make it anonymous
					ScreenName: "recipient-screen-name",
				},
			},
			expectSNACToClient: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceEvilNotification,
				},
				Body: wire.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
					NewEvil: evilDeltaAnon,
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 30,
					UpdatedEvilValue: 30,
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
						},
					},
				},
			},
		},
		{
			name:                "transmit non-anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    newTestSession("recipient-screen-name", sessOptCannedSignonTime),
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectSNACToClient: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceEvilNotification,
				},
				Body: wire.SNAC_0x01_0x10_OServiceEvilNotification{
					NewEvil: evilDelta,
					TLVUserInfo: wire.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 100,
					},
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
					UpdatedEvilValue: 100,
				},
			},
			mockParams: mockParams{
				buddyBroadcasterParams: buddyBroadcasterParams{
					broadcastBuddyArrivedParams: broadcastBuddyArrivedParams{
						{
							screenName: state.NewIdentScreenName("recipient-screen-name"),
						},
					},
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because sender has blocked recipient",
			blockedState:        state.BlockedA,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    nil,
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because recipient has blocked sender",
			blockedState:        state.BlockedB,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    nil,
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't let users warn themselves",
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    nil,
			recipientScreenName: state.NewIdentScreenName("sender-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "sender-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotSupportedByHost,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because recipient is offline",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    nil,
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't transmit anonymous warning from sender to recipient because recipient is offline",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientSession:    nil,
			recipientScreenName: state.NewIdentScreenName("recipient-screen-name"),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     1, // make it anonymous
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMErr,
					RequestID: 1234,
				},
				Body: wire.SNACError{
					Code: wire.ErrorCodeNotLoggedOn,
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
			feedbagManager.EXPECT().
				BlockedState(tc.senderSession.IdentScreenName(), tc.recipientScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(tc.recipientScreenName).
				Return(tc.recipientSession).
				Maybe()
			messageRelayer.EXPECT().
				RelayToScreenName(mock.Anything, tc.recipientScreenName, tc.expectSNACToClient).
				Maybe()
			buddyUpdateBroadcaster := newMockbuddyBroadcaster(t)
			for _, params := range tc.mockParams.broadcastBuddyArrivedParams {
				p := params
				buddyUpdateBroadcaster.EXPECT().
					BroadcastBuddyArrived(mock.Anything, mock.MatchedBy(func(s *state.Session) bool {
						return s.IdentScreenName() == p.screenName
					})).
					Return(nil)
			}
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderSession.DisplayScreenName())
			svc := NewICBMService(messageRelayer, feedbagManager, nil)
			svc.buddyUpdateBroadcaster = buddyUpdateBroadcaster
			outputSNAC, err := svc.EvilRequest(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x08_ICBMEvilRequest))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMService_ParameterQuery(t *testing.T) {
	svc := NewICBMService(nil, nil, nil)

	have := svc.ParameterQuery(nil, wire.SNACFrame{RequestID: 1234})
	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICBM,
			SubGroup:  wire.ICBMParameterReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x04_0x05_ICBMParameterReply{
			MaxSlots:             100,
			ICBMFlags:            3,
			MaxIncomingICBMLen:   512,
			MaxSourceEvil:        999,
			MaxDestinationEvil:   999,
			MinInterICBMInterval: 0,
		},
	}

	assert.Equal(t, want, have)
}
