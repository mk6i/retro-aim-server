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
								Tag:   wire.ICBMTLVTagRequestHostAck,
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
								Tag:   wire.ICBMTLVTagsWantEvents,
								Value: []byte{},
							},
							{
								Tag:   wire.ICBMTLVTagRequestHostAck,
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
								Tag:   wire.ICBMTLVTagsWantEvents,
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
								Tag:   wire.ICBMTLVTagRequestHostAck,
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
								Tag:   wire.ICBMTLVTagRequestHostAck,
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
								Tag:   wire.ICBMTLVTagRequestHostAck,
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
				BlockedState(tc.senderSession.ScreenName(),
					tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(tc.inputSNAC.Body.(wire.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName).
				Return(tc.recipientSession).
				Maybe()
			if tc.recipientSession != nil {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, tc.recipientSession.ScreenName(), tc.expectSNACToClient).
					Maybe()
			}
			//
			// send input SNAC
			//
			svc := NewICBMService(messageRelayer, feedbagManager)
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
		senderScreenName string
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
				BlockedState(tc.senderScreenName, tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent).ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			if tc.blockedState == state.BlockedNo {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, tc.inputSNAC.Body.(wire.SNAC_0x04_0x14_ICBMClientEvent).ScreenName,
						tc.expectSNACToClient)
			}
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderScreenName)
			svc := NewICBMService(messageRelayer, feedbagManager)
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
		// senderScreenName is the session name of the user sending the IM
		senderSession *state.Session
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientBuddies is a list of the recipient's buddies that get
		// updated warning level
		recipientBuddies []string
		broadcastMessage wire.SNACMessage
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient wire.SNACMessage

		expectOutput wire.SNACMessage
	}{
		{
			name:                "transmit anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			broadcastMessage: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyArrived,
				},
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("recipient-screen-name", sessOptCannedSignonTime, sessOptWarning(evilDeltaAnon)).TLVUserInfo(),
				},
			},
			recipientBuddies: []string{"buddy1", "buddy2"},
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
		},
		{
			name:                "transmit non-anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyArrived,
				},
				Body: wire.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("recipient-screen-name", sessOptCannedSignonTime, sessOptWarning(evilDelta)).TLVUserInfo(),
				},
			},
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
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because sender has blocked recipient",
			blockedState:        state.BlockedA,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
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
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
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
			name:                "don't let users block themselves",
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "sender-screen-name",
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				BlockedState(tc.senderSession.ScreenName(), tc.recipientScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			feedbagManager.EXPECT().
				AdjacentUsers(tc.recipientScreenName).
				Return(tc.recipientBuddies, nil).
				Maybe()
			recipSess := newTestSession(tc.recipientScreenName, sessOptCannedSignonTime)
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(tc.recipientScreenName).
				Return(recipSess).
				Maybe()
			messageRelayer.EXPECT().
				RelayToScreenName(mock.Anything, tc.recipientScreenName, tc.expectSNACToClient).
				Maybe()
			messageRelayer.EXPECT().
				RelayToScreenNames(mock.Anything, tc.recipientBuddies, tc.broadcastMessage).
				Maybe()
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderSession.ScreenName())
			svc := NewICBMService(messageRelayer, feedbagManager)
			outputSNAC, err := svc.EvilRequest(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x04_0x08_ICBMEvilRequest))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMService_ParameterQuery(t *testing.T) {
	svc := NewICBMService(nil, nil)

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
